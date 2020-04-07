package testhelpers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"testing"

	"github.com/gobuffalo/httptest"
	"github.com/julienschmidt/httprouter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/negroni"

	"github.com/ory/viper"
	"github.com/ory/x/pointerx"

	"github.com/ory/kratos/driver"
	"github.com/ory/kratos/driver/configuration"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/internal/httpclient/client/common"
	"github.com/ory/kratos/internal/httpclient/models"
	"github.com/ory/kratos/selfservice/flow/settings"
	"github.com/ory/kratos/session"
	"github.com/ory/kratos/x"
)

func SetSettingsStrategyAfterHooks(t *testing.T, strategy string, u string) {
	viper.Set(
		configuration.ViperKeySelfServiceSettingsAfterConfig+"."+strategy,
		HookConfigRedirectTo(t, u))
}

func HookConfigRedirectTo(t *testing.T, u string) (m []map[string]interface{}) {
	var b bytes.Buffer
	_, err := fmt.Fprintf(&b, `[
	{
		"job": "redirect",
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

	viper.Set(configuration.ViperKeyURLsProfile, ts.URL+"/settings")
	viper.Set(configuration.ViperKeyURLsLogin, ts.URL+"/login")

	return ts
}

func NewSettingsAPIServer(t *testing.T, reg *driver.RegistryDefault, ids []identity.Identity) (*httptest.Server, *httptest.Server) {
	public, admin := x.NewRouterPublic(), x.NewRouterAdmin()
	reg.SettingsHandler().RegisterAdminRoutes(admin)

	reg.SettingsHandler().RegisterPublicRoutes(public)
	reg.SettingsStrategies().RegisterPublicRoutes(public)

	n := negroni.Classic()
	n.UseHandler(public)
	hh := x.NewTestCSRFHandler(n, reg)
	reg.WithCSRFHandler(hh)

	tsp, tsa := httptest.NewServer(hh), httptest.NewServer(admin)
	t.Cleanup(tsp.Close)
	t.Cleanup(tsa.Close)

	viper.Set(configuration.ViperKeyURLsSelfPublic, tsp.URL)
	viper.Set(configuration.ViperKeyURLsSelfAdmin, tsa.URL)

	for k := range ids {
		route, _ := session.MockSessionCreateHandlerWithIdentity(t, reg, &ids[k])
		public.GET("/sessions/set/"+strconv.Itoa(k), route)
	}

	return tsp, tsa
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

	assert.Equal(t, viper.GetString(configuration.ViperKeyURLsProfile), res.Request.URL.Scheme+"://"+res.Request.URL.Host+res.Request.URL.Path, "should end up at the settings URL, used: %s", pointerx.StringR(f.Action))

	rs, err := NewSDKClientFromURL(viper.GetString(configuration.ViperKeyURLsSelfPublic)).Common.GetSelfServiceBrowserSettingsRequest(
		common.NewGetSelfServiceBrowserSettingsRequestParams().WithHTTPClient(hc).
			WithRequest(res.Request.URL.Query().Get("request")),
	)
	require.NoError(t, err)
	body, err := json.Marshal(rs.Payload)
	require.NoError(t, err)
	return string(body), rs
}
