// nolint
package testhelpers

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/ory/x/ioutilx"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/viper"
	"github.com/ory/x/pointerx"

	"github.com/ory/kratos/driver"
	"github.com/ory/kratos/driver/configuration"
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
	viper.Set(configuration.ViperKeySelfServiceVerificationUI, ts.URL+"/verification-ts")
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

func VerificationSubmitForm(
	t *testing.T,
	f *models.VerificationFlowMethodConfig,
	hc *http.Client,
	values url.Values,
) (string, *public.GetSelfServiceVerificationFlowOK) {
	require.NotEmpty(t, f.Action)

	res, err := hc.PostForm(pointerx.StringR(f.Action), values)
	require.NoError(t, err)
	defer res.Body.Close()

	b, err := ioutil.ReadAll(res.Body)
	require.NoError(t, err)
	assert.EqualValues(t, http.StatusOK, res.StatusCode, "%s", b)

	assert.Equal(t, viper.GetString(configuration.ViperKeySelfServiceVerificationUI), res.Request.URL.Scheme+"://"+res.Request.URL.Host+res.Request.URL.Path, "should end up at the settings URL, used: %s", pointerx.StringR(f.Action))

	rs, err := NewSDKClientFromURL(viper.GetString(configuration.ViperKeyPublicBaseURL)).Public.GetSelfServiceVerificationFlow(
		public.NewGetSelfServiceVerificationFlowParams().WithHTTPClient(hc).
			WithID(res.Request.URL.Query().Get("flow")),
	)
	require.NoError(t, err)
	body, err := json.Marshal(rs.Payload)
	require.NoError(t, err)
	return string(body), rs
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
