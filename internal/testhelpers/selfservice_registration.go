package testhelpers

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/ory/kratos-client-go"

	"github.com/ory/x/ioutilx"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/driver"
	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/selfservice/flow/registration"
	"github.com/ory/kratos/x"
)

func NewRegistrationUIFlowEchoServer(t *testing.T, reg driver.Registry) *httptest.Server {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		e, err := reg.RegistrationFlowPersister().GetRegistrationFlow(r.Context(), x.ParseUUID(r.URL.Query().Get("flow")))
		require.NoError(t, err)
		reg.Writer().Write(w, r, e)
	}))
	reg.Config(context.Background()).MustSet(config.ViperKeySelfServiceRegistrationUI, ts.URL+"/registration-ts")
	t.Cleanup(ts.Close)
	return ts
}

func InitializeRegistrationFlowViaBrowser(t *testing.T, client *http.Client, ts *httptest.Server) *kratos.RegistrationFlow {
	res, err := client.Get(ts.URL + registration.RouteInitBrowserFlow)
	require.NoError(t, err)
	require.NoError(t, res.Body.Close())

	rs, _, err := NewSDKCustomClient(ts, client).PublicApi.GetSelfServiceRegistrationFlow(context.Background()).Id(res.Request.URL.Query().Get("flow")).Execute()
	require.NoError(t, err)
	assert.Empty(t, rs.Active)

	return rs
}

func InitializeRegistrationFlowViaAPI(t *testing.T, client *http.Client, ts *httptest.Server) *kratos.RegistrationFlow {
	rs, _, err := NewSDKCustomClient(ts, client).PublicApi.InitializeSelfServiceRegistrationViaAPIFlow(context.Background()).Execute()
	require.NoError(t, err)
	assert.Empty(t, rs.Active)
	return rs
}

func GetRegistrationFlowMethodConfig(t *testing.T, rs *kratos.RegistrationFlow, id string) *kratos.RegistrationFlowMethodConfig {
	require.NotEmpty(t, rs.Methods[id])
	require.NotEmpty(t, rs.Methods[id].Config)
	require.NotEmpty(t, rs.Methods[id].Config.Action)
	c := rs.Methods[id].Config
	return &c
}

func RegistrationMakeRequest(
	t *testing.T,
	isAPI bool,
	f *kratos.RegistrationFlowMethodConfig,
	hc *http.Client,
	values string,
) (string, *http.Response) {
	require.NotEmpty(t, f.Action)

	res, err := hc.Do(NewRequest(t, isAPI, "POST", f.Action, bytes.NewBufferString(values)))
	require.NoError(t, err)
	defer res.Body.Close()

	return string(ioutilx.MustReadAll(res.Body)), res
}

// SubmitRegistrationForm (for Browser and API!), fills out the form and modifies
// // the form values with `withValues`, and submits the form. Returns the body and checks for expectedStatusCode and
// // expectedURL on completion
func SubmitRegistrationForm(
	t *testing.T,
	isAPI bool,
	hc *http.Client,
	publicTS *httptest.Server,
	withValues func(v url.Values),
	method identity.CredentialsType,
	expectedStatusCode int,
	expectedURL string,
) string {
	if hc == nil {
		hc = new(http.Client)
	}

	hc.Transport = NewTransportWithLogger(hc.Transport, t)
	var payload *kratos.RegistrationFlow
	if isAPI {
		payload = InitializeRegistrationFlowViaAPI(t, hc, publicTS)
	} else {
		payload = InitializeRegistrationFlowViaBrowser(t, hc, publicTS)
	}

	time.Sleep(time.Millisecond) // add a bit of delay to allow `1ns` to time out.

	config := GetRegistrationFlowMethodConfig(t, payload, method.String())
	values := SDKFormFieldsToURLValues(config.Nodes)
	withValues(values)
	b, res := RegistrationMakeRequest(t, isAPI, config, hc, EncodeFormAsJSON(t, isAPI, values))
	assert.EqualValues(t, expectedStatusCode, res.StatusCode, "%s", b)
	assert.Contains(t, res.Request.URL.String(), expectedURL, "%+v\n\t%s", res.Request, b)
	return b
}
