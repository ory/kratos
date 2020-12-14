// nolint
package testhelpers

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/ory/x/ioutilx"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/x/pointerx"

	"github.com/ory/kratos/driver"
	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/internal/httpclient/client/public"
	"github.com/ory/kratos/internal/httpclient/models"
	"github.com/ory/kratos/selfservice/flow/verification"
	"github.com/ory/kratos/x"
)

func NewVerificationUIFlowEchoServer(t *testing.T, reg driver.Registry) *httptest.Server {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		e, err := reg.VerificationFlowPersister().GetVerificationFlow(r.Context(), x.ParseUUID(r.URL.Query().Get("flow")))
		require.NoError(t, err)
		reg.Writer().Write(w, r, e)
	}))
	reg.Configuration().MustSet(config.ViperKeySelfServiceVerificationUI, ts.URL+"/verification-ts")
	t.Cleanup(ts.Close)
	return ts
}

func GetVerificationFlow(t *testing.T, client *http.Client, ts *httptest.Server) *public.GetSelfServiceVerificationFlowOK {
	publicClient := NewSDKClient(ts)

	res, err := client.Get(ts.URL + verification.RouteInitBrowserFlow)
	require.NoError(t, err)
	require.NoError(t, res.Body.Close())

	rs, err := publicClient.Public.GetSelfServiceVerificationFlow(
		public.NewGetSelfServiceVerificationFlowParams().WithHTTPClient(client).
			WithID(res.Request.URL.Query().Get("flow")),
	)
	require.NoError(t, err, "%s", res.Request.URL.String())
	assert.Empty(t, rs.Payload.Active)

	return rs
}

func InitializeVerificationFlowViaBrowser(t *testing.T, client *http.Client, ts *httptest.Server) *public.GetSelfServiceVerificationFlowOK {
	publicClient := NewSDKClient(ts)
	res, err := client.Get(ts.URL + verification.RouteInitBrowserFlow)
	require.NoError(t, err)
	require.NoError(t, res.Body.Close())

	rs, err := publicClient.Public.GetSelfServiceVerificationFlow(
		public.NewGetSelfServiceVerificationFlowParams().WithHTTPClient(client).
			WithID(res.Request.URL.Query().Get("flow")),
	)
	require.NoError(t, err)
	assert.Empty(t, rs.Payload.Active)

	return rs
}

func InitializeVerificationFlowViaAPI(t *testing.T, client *http.Client, ts *httptest.Server) *public.InitializeSelfServiceVerificationViaAPIFlowOK {
	publicClient := NewSDKClient(ts)

	rs, err := publicClient.Public.InitializeSelfServiceVerificationViaAPIFlow(public.
		NewInitializeSelfServiceVerificationViaAPIFlowParams().WithHTTPClient(client))
	require.NoError(t, err)
	assert.Empty(t, rs.Payload.Active)

	return rs
}

func GetVerificationFlowMethodConfig(t *testing.T, rs *models.VerificationFlow, id string) *models.VerificationFlowMethodConfig {
	require.NotEmpty(t, rs.Methods[id])
	require.NotEmpty(t, rs.Methods[id].Config)
	require.NotEmpty(t, rs.Methods[id].Config.Action)
	return rs.Methods[id].Config
}

func VerificationMakeRequest(
	t *testing.T,
	isAPI bool,
	f *models.VerificationFlowMethodConfig,
	hc *http.Client,
	values string,
) (string, *http.Response) {
	require.NotEmpty(t, f.Action)

	res, err := hc.Do(NewRequest(t, isAPI, "POST", pointerx.StringR(f.Action), bytes.NewBufferString(values)))
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
	method identity.CredentialsType,
	expectedStatusCode int,
	expectedURL string,
) string {
	hc.Transport = NewTransportWithLogger(hc.Transport, t)
	var f *models.VerificationFlow
	if isAPI {
		f = InitializeVerificationFlowViaAPI(t, hc, publicTS).Payload
	} else {
		f = InitializeVerificationFlowViaBrowser(t, hc, publicTS).Payload
	}

	time.Sleep(time.Millisecond) // add a bit of delay to allow `1ns` to time out.

	config := GetVerificationFlowMethodConfig(t, f, method.String())
	formPayload := SDKFormFieldsToURLValues(config.Fields)
	withValues(formPayload)

	b, res := VerificationMakeRequest(t, isAPI, config, hc, EncodeFormAsJSON(t, isAPI, formPayload))
	assert.EqualValues(t, expectedStatusCode, res.StatusCode, "%s", b)
	assert.Contains(t, res.Request.URL.String(), expectedURL, "%+v\n\t%s", res.Request, b)

	return b
}
