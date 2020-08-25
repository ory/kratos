// nolint
package testhelpers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
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
	"github.com/ory/kratos/selfservice/strategy/password"
	"github.com/ory/kratos/x"
)

func InitializeSettingsFlowViaBrowser(t *testing.T, client *http.Client, ts *httptest.Server) *common.GetSelfServiceSettingsFlowOK {
	publicClient := NewSDKClient(ts)

	res, err := client.Get(ts.URL + settings.RouteInitBrowserFlow)
	require.NoError(t, err)
	require.NoError(t, res.Body.Close())

	rs, err := publicClient.Common.GetSelfServiceSettingsFlow(
		common.NewGetSelfServiceSettingsFlowParams().WithHTTPClient(client).
			WithID(res.Request.URL.Query().Get("flow")),
	)
	require.NoError(t, err)
	assert.Empty(t, rs.Payload.Active)

	return rs
}

func InitializeSettingsFlowViaAPI(t *testing.T, client *http.Client, ts *httptest.Server) *common.InitializeSelfServiceSettingsViaAPIFlowOK {
	publicClient := NewSDKClient(ts)

	rs, err := publicClient.Common.InitializeSelfServiceSettingsViaAPIFlow(common.
		NewInitializeSelfServiceSettingsViaAPIFlowParams().WithHTTPClient(client))
	require.NoError(t, err)
	assert.Empty(t, rs.Payload.Active)

	return rs
}

func GetSettingsMethodConfig(t *testing.T, primaryUser *http.Client, ts *httptest.Server, id string) *models.FlowMethodConfig {
	rs := InitializeSettingsFlowViaBrowser(t, primaryUser, ts)

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

	viper.Set(configuration.ViperKeySelfServiceSettingsURL, ts.URL+"/settings")
	viper.Set(configuration.ViperKeySelfServiceLoginUI, ts.URL+"/login")

	return ts
}

func NewSettingsLoginAcceptAPIServer(t *testing.T, adminClient *client.OryKratos) *httptest.Server {
	var called int
	loginTS := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, 0, called)
		called++

		viper.Set(configuration.ViperKeySelfServiceSettingsPrivilegedAuthenticationAfter, "5m")

		res, err := adminClient.Common.GetSelfServiceLoginFlow(common.NewGetSelfServiceLoginFlowParams().WithID(r.URL.Query().Get("flow")))

		require.NoError(t, err)
		require.NotEmpty(t, res.Payload.RequestURL)

		redir := urlx.ParseOrPanic(*res.Payload.RequestURL).Query().Get("return_to")
		t.Logf("Redirecting to: %s", redir)
		http.Redirect(w, r, redir, http.StatusFound)
	}))
	t.Cleanup(func() {
		loginTS.Close()
	})
	viper.Set(configuration.ViperKeySelfServiceLoginUI, loginTS.URL+"/login")
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

	viper.Set(configuration.ViperKeyPublicBaseURL, tsp.URL)
	viper.Set(configuration.ViperKeyAdminBaseURL, tsa.URL)
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
	isAPI bool,
	f *models.FlowMethodConfig,
	hc *http.Client,
	values string,
	expectedStatusCode int,
) (string, *common.GetSelfServiceSettingsFlowOK) {
	require.NotEmpty(t, f.Action)

	req, err := http.NewRequest("POST", pointerx.StringR(f.Action), bytes.NewBufferString(values))
	require.NoError(t, err)

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "text/html")
	if isAPI {
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")
	}

	res, err := hc.Do(req)
	require.NoError(t, err)
	defer res.Body.Close()

	b := x.MustReadAll(res.Body)
	assert.EqualValues(t, expectedStatusCode, res.StatusCode, "%s", b)

	expectURL := viper.GetString(configuration.ViperKeySelfServiceSettingsURL)
	if isAPI {
		expectURL = password.RouteSettings
	}
	assert.Contains(t, res.Request.URL.String(), expectURL, "should end up at the settings URL, used: %s\n\t%s", pointerx.StringR(f.Action), b)

	rs, err := NewSDKClientFromURL(viper.GetString(configuration.ViperKeyPublicBaseURL)).Common.GetSelfServiceSettingsFlow(
		common.NewGetSelfServiceSettingsFlowParams().WithHTTPClient(hc).
			WithID(res.Request.URL.Query().Get("flow")),
	)
	require.NoError(t, err)
	body, err := json.Marshal(rs.Payload)
	require.NoError(t, err)
	return string(body), rs
}
