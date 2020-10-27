package testhelpers

import (
	"bytes"
	"github.com/ory/x/ioutilx"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/x/pointerx"

	"github.com/ory/viper"

	"github.com/ory/kratos/driver"
	"github.com/ory/kratos/driver/configuration"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/internal/httpclient/client/public"
	"github.com/ory/kratos/internal/httpclient/models"
	"github.com/ory/kratos/selfservice/flow/login"
	"github.com/ory/kratos/x"
)

func NewLoginUIFlowEchoServer(t *testing.T, reg driver.Registry) *httptest.Server {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		e, err := reg.LoginFlowPersister().GetLoginFlow(r.Context(), x.ParseUUID(r.URL.Query().Get("flow")))
		require.NoError(t, err)
		reg.Writer().Write(w, r, e)
	}))
	viper.Set(configuration.ViperKeySelfServiceLoginUI, ts.URL+"/login-ts")
	t.Cleanup(ts.Close)
	return ts
}

func NewLoginUIWith401Response(t *testing.T) *httptest.Server {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	viper.Set(configuration.ViperKeySelfServiceLoginUI, ts.URL+"/login-ts")
	t.Cleanup(ts.Close)
	return ts
}

func InitializeLoginFlowViaBrowser(t *testing.T, client *http.Client, ts *httptest.Server, forced bool) *public.GetSelfServiceLoginFlowOK {
	publicClient := NewSDKClient(ts)

	q := ""
	if forced {
		q = "?refresh=true"
	}

	res, err := client.Get(ts.URL + login.RouteInitBrowserFlow + q)
	require.NoError(t, err)
	require.NoError(t, res.Body.Close())

	rs, err := publicClient.Public.GetSelfServiceLoginFlow(
		public.NewGetSelfServiceLoginFlowParams().WithHTTPClient(client).
			WithID(res.Request.URL.Query().Get("flow")),
	)
	require.NoError(t, err)
	assert.Empty(t, rs.Payload.Active)

	return rs
}

func InitializeLoginFlowViaAPI(t *testing.T, client *http.Client, ts *httptest.Server, forced bool) *public.InitializeSelfServiceLoginViaAPIFlowOK {
	publicClient := NewSDKClient(ts)

	rs, err := publicClient.Public.InitializeSelfServiceLoginViaAPIFlow(public.
		NewInitializeSelfServiceLoginViaAPIFlowParams().WithHTTPClient(client).WithRefresh(pointerx.Bool(forced)))
	require.NoError(t, err)
	assert.Empty(t, rs.Payload.Active)

	return rs
}

func GetLoginFlowMethodConfig(t *testing.T, rs *models.LoginFlow, id string) *models.LoginFlowMethodConfig {
	require.NotEmpty(t, rs.Methods[id])
	require.NotEmpty(t, rs.Methods[id].Config)
	require.NotEmpty(t, rs.Methods[id].Config.Action)
	return rs.Methods[id].Config
}

func LoginMakeRequest(
	t *testing.T,
	isAPI bool,
	f *models.LoginFlowMethodConfig,
	hc *http.Client,
	values string,
) (string, *http.Response) {
	require.NotEmpty(t, f.Action)

	res, err := hc.Do(NewRequest(t, isAPI, "POST", pointerx.StringR(f.Action), bytes.NewBufferString(values)))
	require.NoError(t, err)
	defer res.Body.Close()

	return string(ioutilx.MustReadAll(res.Body)), res
}

// SubmitLoginForm initiates a login flow (for Browser and API!), fills out the form and modifies
// the form values with `withValues`, and submits the form. Returns the body and checks for expectedStatusCode and
// expectedURL on completion
func SubmitLoginForm(
	t *testing.T,
	isAPI bool,
	hc *http.Client,
	publicTS *httptest.Server,
	withValues func(v url.Values),
	method identity.CredentialsType,
	forced bool,
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
	var f *models.LoginFlow
	if isAPI {
		f = InitializeLoginFlowViaAPI(t, hc, publicTS, forced).Payload
	} else {
		f = InitializeLoginFlowViaBrowser(t, hc, publicTS, forced).Payload
	}

	time.Sleep(time.Millisecond) // add a bit of delay to allow `1ns` to time out.

	config := GetLoginFlowMethodConfig(t, f, method.String())

	payload := SDKFormFieldsToURLValues(config.Fields)
	withValues(payload)
	b, res := LoginMakeRequest(t, isAPI, config, hc, EncodeFormAsJSON(t, isAPI, payload))
	assert.EqualValues(t, expectedStatusCode, res.StatusCode, "%s", b)
	assert.Contains(t, res.Request.URL.String(), expectedURL, "%+v\n\t%s", res.Request, b)

	return b
}
