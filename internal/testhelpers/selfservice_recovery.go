// nolint
package testhelpers

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	kratos "github.com/ory/kratos-client-go"
	"github.com/ory/kratos/driver"
	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/selfservice/flow/verification"
	"github.com/ory/kratos/x"
	"github.com/ory/x/ioutilx"
)

func NewVerificationUIFlowEchoServer(t *testing.T, reg driver.Registry) *httptest.Server {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		e, err := reg.VerificationFlowPersister().GetVerificationFlow(r.Context(), x.ParseUUID(r.URL.Query().Get("flow")))
		require.NoError(t, err)
		reg.Writer().Write(w, r, e)
	}))
	reg.Config(context.Background()).MustSet(config.ViperKeySelfServiceVerificationUI, ts.URL+"/verification-ts")
	t.Cleanup(ts.Close)
	return ts
}

func GetVerificationFlow(t *testing.T, client *http.Client, ts *httptest.Server) *kratos.VerificationFlow {
	publicClient := NewSDKCustomClient(ts, client)

	res, err := client.Get(ts.URL + verification.RouteInitBrowserFlow)
	require.NoError(t, err)
	require.NoError(t, res.Body.Close())

	rs, _, err := publicClient.PublicApi.GetSelfServiceVerificationFlow(context.Background()).Id(res.Request.URL.Query().Get("flow")).Execute()
	require.NoError(t, err, "%s", res.Request.URL.String())
	assert.Empty(t, rs.Active)

	return rs
}

func InitializeVerificationFlowViaBrowser(t *testing.T, client *http.Client, ts *httptest.Server) *kratos.VerificationFlow {
	publicClient := NewSDKCustomClient(ts, client)
	res, err := client.Get(ts.URL + verification.RouteInitBrowserFlow)
	require.NoError(t, err)
	require.NoError(t, res.Body.Close())

	rs, _, err := publicClient.PublicApi.GetSelfServiceVerificationFlow(context.Background()).Id(res.Request.URL.Query().Get("flow")).Execute()
	require.NoError(t, err)
	assert.Empty(t, rs.Active)

	return rs
}

func InitializeVerificationFlowViaAPI(t *testing.T, client *http.Client, ts *httptest.Server) *kratos.VerificationFlow {
	publicClient := NewSDKCustomClient(ts, client)

	rs, _, err := publicClient.PublicApi.InitializeSelfServiceVerificationForNativeApps(context.Background()).Execute()
	require.NoError(t, err)
	assert.Empty(t, rs.Active)

	return rs
}

func VerificationMakeRequest(
	t *testing.T,
	isAPI bool,
	f *kratos.VerificationFlow,
	hc *http.Client,
	values string,
) (string, *http.Response) {
	require.NotEmpty(t, f.Ui.Action)

	res, err := hc.Do(NewRequest(t, isAPI, "POST", f.Ui.Action, bytes.NewBufferString(values)))
	require.NoError(t, err)
	defer res.Body.Close()

	return string(ioutilx.MustReadAll(res.Body)), res
}

// SubmitVerificationForm initiates a registration flow (for Browser and API!), fills out the form and modifies
// the form values with `withValues`, and submits the form. If completed, it will return the flow as JSON.
func SubmitVerificationForm(
	t *testing.T,
	isAPI bool,
	hc *http.Client,
	publicTS *httptest.Server,
	withValues func(v url.Values),
	expectedStatusCode int,
	expectedURL string,
) string {
	hc.Transport = NewTransportWithLogger(hc.Transport, t)
	var f *kratos.VerificationFlow
	if isAPI {
		f = InitializeVerificationFlowViaAPI(t, hc, publicTS)
	} else {
		f = InitializeVerificationFlowViaBrowser(t, hc, publicTS)
	}

	time.Sleep(time.Millisecond) // add a bit of delay to allow `1ns` to time out.

	formPayload := SDKFormFieldsToURLValues(f.Ui.Nodes)
	withValues(formPayload)

	b, res := VerificationMakeRequest(t, isAPI, f, hc, EncodeFormAsJSON(t, isAPI, formPayload))
	assert.EqualValues(t, expectedStatusCode, res.StatusCode, "%s", b)
	assert.Contains(t, res.Request.URL.String(), expectedURL, "%+v\n\t%s", res.Request, b)

	return b
}
