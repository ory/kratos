// nolint
package testhelpers

import (
	"bytes"
	"context"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/ory/x/ioutilx"

	"github.com/gobuffalo/httptest"
	"github.com/julienschmidt/httprouter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/sjson"
	"github.com/urfave/negroni"

	"github.com/ory/x/urlx"

	"github.com/ory/x/pointerx"

	"github.com/ory/kratos-client-go/client"
	"github.com/ory/kratos-client-go/client/public"
	"github.com/ory/kratos-client-go/models"
	"github.com/ory/kratos/driver"
	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/selfservice/flow/settings"
	"github.com/ory/kratos/x"
)

func NewSettingsUIFlowEchoServer(t *testing.T, reg driver.Registry) *httptest.Server {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		e, err := reg.SettingsFlowPersister().GetSettingsFlow(r.Context(), x.ParseUUID(r.URL.Query().Get("flow")))
		require.NoError(t, err)
		reg.Writer().Write(w, r, e)
	}))
	reg.Configuration().MustSet(config.ViperKeySelfServiceSettingsURL, ts.URL+"/settings-ts")
	t.Cleanup(ts.Close)
	return ts
}

func InitializeSettingsFlowViaBrowser(t *testing.T, client *http.Client, ts *httptest.Server) *public.GetSelfServiceSettingsFlowOK {
	publicClient := NewSDKClient(ts)

	res, err := client.Get(ts.URL + settings.RouteInitBrowserFlow)
	require.NoError(t, err)
	require.NoError(t, res.Body.Close())

	rs, err := publicClient.Public.GetSelfServiceSettingsFlow(
		public.NewGetSelfServiceSettingsFlowParams().WithHTTPClient(client).
			WithID(res.Request.URL.Query().Get("flow")), nil,
	)
	require.NoError(t, err)
	assert.Empty(t, rs.Payload.Active)

	return rs
}

func InitializeSettingsFlowViaAPI(t *testing.T, client *http.Client, ts *httptest.Server) *public.InitializeSelfServiceSettingsViaAPIFlowOK {
	publicClient := NewSDKClient(ts)

	rs, err := publicClient.Public.InitializeSelfServiceSettingsViaAPIFlow(public.
		NewInitializeSelfServiceSettingsViaAPIFlowParams().WithHTTPClient(client), nil)
	require.NoError(t, err)
	assert.Empty(t, rs.Payload.Active)

	return rs
}

func EncodeFormAsJSON(t *testing.T, isApi bool, values url.Values) (payload string) {
	if !isApi {
		return values.Encode()
	}
	payload = "{}"
	for k := range values {
		var err error
		payload, err = sjson.Set(payload, strings.ReplaceAll(k, ".", "\\."), values.Get(k))
		require.NoError(t, err)
	}
	return payload
}

func ExpectStatusCode(isAPI bool, api, browser int) int {
	if isAPI {
		return api
	}
	return browser
}

func ExpectURL(isAPI bool, api, browser string) string {
	if isAPI {
		return api
	}
	return browser
}

func GetSettingsFlowMethodConfig(t *testing.T, rs *models.SettingsFlow, id string) *models.SettingsFlowMethodConfig {
	require.NotEmpty(t, rs.Methods[id])
	require.NotEmpty(t, rs.Methods[id].Config)
	require.NotEmpty(t, rs.Methods[id].Config.Action)
	return rs.Methods[id].Config
}

func GetSettingsFlowMethodConfigDeprecated(t *testing.T, primaryUser *http.Client, ts *httptest.Server, id string) *models.SettingsFlowMethodConfig {
	rs := InitializeSettingsFlowViaBrowser(t, primaryUser, ts)

	require.NotEmpty(t, rs.Payload.Methods[id])
	require.NotEmpty(t, rs.Payload.Methods[id].Config)
	require.NotEmpty(t, rs.Payload.Methods[id].Config.Action)

	return rs.Payload.Methods[id].Config
}

func NewSettingsUITestServer(t *testing.T, conf *config.Provider) *httptest.Server {
	router := httprouter.New()
	router.GET("/settings", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		w.WriteHeader(http.StatusNoContent)
	})
	router.GET("/login", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		w.WriteHeader(http.StatusUnauthorized)
	})
	ts := httptest.NewServer(router)
	t.Cleanup(ts.Close)

	conf.MustSet(config.ViperKeySelfServiceSettingsURL, ts.URL+"/settings")
	conf.MustSet(config.ViperKeySelfServiceLoginUI, ts.URL+"/login")

	return ts
}

func NewSettingsUIEchoServer(t *testing.T, reg *driver.RegistryDefault) *httptest.Server {
	router := httprouter.New()
	router.GET("/settings", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		res, err := reg.SettingsFlowPersister().GetSettingsFlow(r.Context(), x.ParseUUID(r.URL.Query().Get("flow")))
		require.NoError(t, err)
		reg.Writer().Write(w, r, res)
	})

	router.GET("/login", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		w.WriteHeader(http.StatusUnauthorized)
	})
	ts := httptest.NewServer(router)
	t.Cleanup(ts.Close)

	reg.Configuration().MustSet(config.ViperKeySelfServiceSettingsURL, ts.URL+"/settings")
	reg.Configuration().MustSet(config.ViperKeySelfServiceLoginUI, ts.URL+"/login")

	return ts
}

func NewSettingsLoginAcceptAPIServer(t *testing.T, adminClient *client.OryKratos, conf *config.Provider) *httptest.Server {
	var called int
	loginTS := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, 0, called)
		called++

		conf.MustSet(config.ViperKeySelfServiceSettingsPrivilegedAuthenticationAfter, "5m")

		res, err := adminClient.Public.GetSelfServiceLoginFlow(public.NewGetSelfServiceLoginFlowParams().WithID(r.URL.Query().Get("flow")))

		require.NoError(t, err)
		require.NotEmpty(t, res.Payload.RequestURL)

		redir := urlx.ParseOrPanic(*res.Payload.RequestURL).Query().Get("return_to")
		t.Logf("Redirecting to: %s", redir)
		http.Redirect(w, r, redir, http.StatusFound)
	}))
	t.Cleanup(func() {
		loginTS.Close()
	})
	conf.MustSet(config.ViperKeySelfServiceLoginUI, loginTS.URL+"/login")
	return loginTS
}

func NewSettingsAPIServer(t *testing.T, reg *driver.RegistryDefault, ids map[string]*identity.Identity) (*httptest.Server, *httptest.Server, map[string]*http.Client) {
	public, admin := x.NewRouterPublic(), x.NewRouterAdmin()
	reg.SettingsHandler().RegisterAdminRoutes(admin)

	n := negroni.Classic()
	n.UseHandler(public)
	hh := x.NewTestCSRFHandler(n, reg)
	reg.WithCSRFHandler(hh)

	reg.SettingsHandler().RegisterPublicRoutes(public)
	reg.SettingsStrategies().RegisterPublicRoutes(public)
	reg.LoginHandler().RegisterPublicRoutes(public)
	reg.LoginHandler().RegisterAdminRoutes(admin)
	reg.LoginStrategies().RegisterPublicRoutes(public)

	tsp, tsa := httptest.NewServer(hh), httptest.NewServer(admin)
	t.Cleanup(tsp.Close)
	t.Cleanup(tsa.Close)

	reg.Configuration().MustSet(config.ViperKeyPublicBaseURL, tsp.URL)
	reg.Configuration().MustSet(config.ViperKeyAdminBaseURL, tsa.URL)
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

func SettingsMakeRequest(
	t *testing.T,
	isAPI bool,
	f *models.SettingsFlowMethodConfig,
	hc *http.Client,
	values string,
) (string, *http.Response) {
	require.NotEmpty(t, f.Action)

	res, err := hc.Do(NewRequest(t, isAPI, "POST", pointerx.StringR(f.Action), bytes.NewBufferString(values)))
	require.NoError(t, err)
	defer res.Body.Close()

	return string(ioutilx.MustReadAll(res.Body)), res
}

// SubmitSettingsForm initiates a settings flow (for Browser and API!), fills out the form and modifies
// the form values with `withValues`, and submits the form. If completed, it will return the flow as JSON.
func SubmitSettingsForm(
	t *testing.T,
	isAPI bool,
	hc *http.Client,
	publicTS *httptest.Server,
	withValues func(v url.Values),
	method string,
	expectedStatusCode int,
	expectedURL string,
) string {
	if hc == nil {
		hc = new(http.Client)
		if !isAPI {
			hc = NewClientWithCookies(t)
		}
	}

	hc.Transport = NewTransportWithLogger(hc.Transport, t)
	var payload *models.SettingsFlow
	if isAPI {
		payload = InitializeSettingsFlowViaAPI(t, hc, publicTS).Payload
	} else {
		payload = InitializeSettingsFlowViaBrowser(t, hc, publicTS).Payload
	}

	time.Sleep(time.Millisecond * 10) // add a bit of delay to allow `1ns` to time out.

	config := GetSettingsFlowMethodConfig(t, payload, method)
	values := SDKFormFieldsToURLValues(config.Fields)
	withValues(values)

	b, res := SettingsMakeRequest(t, isAPI, config, hc, EncodeFormAsJSON(t, isAPI, values))
	assert.EqualValues(t, expectedStatusCode, res.StatusCode, "%s", b)
	assert.Contains(t, res.Request.URL.String(), expectedURL, "%+v\n\t%s", res.Request, b)

	return b
}
