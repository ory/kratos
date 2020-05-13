package testhelpers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"testing"

	"github.com/gobuffalo/httptest"
	"github.com/julienschmidt/httprouter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/negroni"

	"github.com/ory/x/urlx"

	"github.com/ory/viper"
	"github.com/ory/x/pointerx"

	"github.com/ory/kratos/driver"
	"github.com/ory/kratos/driver/configuration"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/internal/httpclient/client"
	"github.com/ory/kratos/internal/httpclient/client/common"
	"github.com/ory/kratos/internal/httpclient/models"
	"github.com/ory/kratos/selfservice/flow/settings"
	"github.com/ory/kratos/x"
)

func HookConfigRedirectTo(t *testing.T, u string) (m []map[string]interface{}) {
	var b bytes.Buffer
	_, err := fmt.Fprintf(&b, `[
	{
		"hook": "redirect",
		"config": {
          "default_redirect_url": "%s",
          "allow_user_defined_redirect": true
		}
	}
]`, u)
	require.NoError(t, err)
	require.NoError(t, json.NewDecoder(&b).Decode(&m))

	return m
}

func HookVerify(t *testing.T) (m []map[string]interface{}) {
	require.NoError(t, json.NewDecoder(bytes.NewBufferString(`[{"job": "verify"}]`)).Decode(&m))
	return m
}

func GetSettingsRequest(t *testing.T, primaryUser *http.Client, ts *httptest.Server) *common.GetSelfServiceBrowserSettingsRequestOK {
	publicClient := NewSDKClient(ts)

	res, err := primaryUser.Get(ts.URL + settings.PublicPath)
	require.NoError(t, err)
	require.NoError(t, res.Body.Close())

	rs, err := publicClient.Common.GetSelfServiceBrowserSettingsRequest(
		common.NewGetSelfServiceBrowserSettingsRequestParams().WithHTTPClient(primaryUser).
			WithRequest(res.Request.URL.Query().Get("request")),
	)
	require.NoError(t, err)
	assert.Empty(t, rs.Payload.Active)

	return rs
}

func GetSettingsMethodConfig(t *testing.T, primaryUser *http.Client, ts *httptest.Server, id string) *models.RequestMethodConfig {
	rs := GetSettingsRequest(t, primaryUser, ts)

	require.NotEmpty(t, rs.Payload.Methods[id])
	require.NotEmpty(t, rs.Payload.Methods[id].Config)
	require.NotEmpty(t, rs.Payload.Methods[id].Config.Action)

	return rs.Payload.Methods[id].Config
}

func NewSettingsUITestServer(t *testing.T) *httptest.Server {
	router := httprouter.New()
	router.GET("/settings", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		w.WriteHeader(http.StatusNoContent)
	})
	router.GET("/login", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		w.WriteHeader(http.StatusUnauthorized)
	})
	ts := httptest.NewServer(router)
	t.Cleanup(ts.Close)

	viper.Set(configuration.ViperKeyURLsSettings, ts.URL+"/settings")
	viper.Set(configuration.ViperKeyURLsLogin, ts.URL+"/login")

	return ts
}

func NewSettingsLoginAcceptAPIServer(t *testing.T, adminClient *client.OryKratos) *httptest.Server {
	var called int
	loginTS := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, 0, called)
		called++

		viper.Set(configuration.ViperKeySelfServicePrivilegedAuthenticationAfter, "5m")

		res, err := adminClient.Common.GetSelfServiceBrowserLoginRequest(common.NewGetSelfServiceBrowserLoginRequestParams().WithRequest(r.URL.Query().Get("request")))
		require.NoError(t, err)
		require.NotEmpty(t, res.Payload.RequestURL)

		redir := urlx.ParseOrPanic(*res.Payload.RequestURL).Query().Get("return_to")
		t.Logf("Redirecting to: %s", redir)
		http.Redirect(w, r, redir, http.StatusFound)
	}))
	t.Cleanup(func() {
		loginTS.Close()
	})
	viper.Set(configuration.ViperKeyURLsLogin, loginTS.URL+"/login")
	return loginTS
}

func NewSettingsAPIServer(t *testing.T, reg *driver.RegistryDefault, ids map[string]*identity.Identity) (*httptest.Server, *httptest.Server, map[string]*http.Client) {
	public, admin := x.NewRouterPublic(), x.NewRouterAdmin()
	reg.SettingsHandler().RegisterAdminRoutes(admin)

	reg.SettingsHandler().RegisterPublicRoutes(public)
	reg.SettingsStrategies().RegisterPublicRoutes(public)
	reg.LoginHandler().RegisterPublicRoutes(public)
	reg.LoginHandler().RegisterAdminRoutes(admin)
	reg.LoginStrategies().RegisterPublicRoutes(public)

	n := negroni.Classic()
	n.UseHandler(public)
	hh := x.NewTestCSRFHandler(n, reg)
	reg.WithCSRFHandler(hh)

	tsp, tsa := httptest.NewServer(hh), httptest.NewServer(admin)
	t.Cleanup(tsp.Close)
	t.Cleanup(tsa.Close)

	viper.Set(configuration.ViperKeyURLsSelfPublic, tsp.URL)
	viper.Set(configuration.ViperKeyURLsSelfAdmin, tsa.URL)
	return tsp, tsa, AddAndLoginIdentities(t, reg, &httptest.Server{Config: &http.Server{Handler: public}, URL: tsp.URL}, ids)
}

// AddAndLoginIdentities adds the given identities to the store (like a registration flow) and returns http.Clients
// which contain their sessions.
func AddAndLoginIdentities(t *testing.T, reg *driver.RegistryDefault, public *httptest.Server, ids map[string]*identity.Identity) map[string]*http.Client {
	result := map[string]*http.Client{}
	for k := range ids {
		tid := x.NewUUID().String()
		_ = reg.PrivilegedIdentityPool().DeleteIdentity(context.Background(), ids[k].ID)
		route, _ := MockSessionCreateHandlerWithIdentity(t, reg, ids[k])
		location := "/sessions/set/" + tid

		if router, ok := public.Config.Handler.(*x.RouterPublic); ok {
			router.Router.GET(location, route)
		} else if router, ok := public.Config.Handler.(*httprouter.Router); ok {
			router.GET(location, route)
		} else if router, ok := public.Config.Handler.(*x.RouterAdmin); ok {
			router.GET(location, route)
		} else {
			t.Logf("Got unknown type: %T", public.Config.Handler)
			t.FailNow()
		}
		result[k] = NewSessionClient(t, public.URL+location)
	}
	return result
}

func SettingsSubmitForm(
	t *testing.T,
	f *models.RequestMethodConfig,
	hc *http.Client,
	values url.Values,
) (string, *common.GetSelfServiceBrowserSettingsRequestOK) {
	require.NotEmpty(t, f.Action)

	res, err := hc.PostForm(pointerx.StringR(f.Action), values)
	require.NoError(t, err)
	defer res.Body.Close()

	b, err := ioutil.ReadAll(res.Body)
	require.NoError(t, err)
	assert.EqualValues(t, http.StatusNoContent, res.StatusCode, "%s", b)

	assert.Equal(t, viper.GetString(configuration.ViperKeyURLsSettings), res.Request.URL.Scheme+"://"+res.Request.URL.Host+res.Request.URL.Path, "should end up at the settings URL, used: %s", pointerx.StringR(f.Action))

	rs, err := NewSDKClientFromURL(viper.GetString(configuration.ViperKeyURLsSelfPublic)).Common.GetSelfServiceBrowserSettingsRequest(
		common.NewGetSelfServiceBrowserSettingsRequestParams().WithHTTPClient(hc).
			WithRequest(res.Request.URL.Query().Get("request")),
	)
	require.NoError(t, err)
	body, err := json.Marshal(rs.Payload)
	require.NoError(t, err)
	return string(body), rs
}
