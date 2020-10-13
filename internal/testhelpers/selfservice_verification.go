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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/viper"
	"github.com/ory/x/pointerx"

	"github.com/ory/kratos/driver"
	"github.com/ory/kratos/driver/configuration"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/internal/httpclient/client/public"
	"github.com/ory/kratos/internal/httpclient/models"
	"github.com/ory/kratos/selfservice/flow/recovery"
	"github.com/ory/kratos/x"
)

func NewRecoveryUIFlowEchoServer(t *testing.T, reg driver.Registry) *httptest.Server {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		e, err := reg.RecoveryFlowPersister().GetRecoveryFlow(r.Context(), x.ParseUUID(r.URL.Query().Get("flow")))
		require.NoError(t, err)
		reg.Writer().Write(w, r, e)
	}))
	viper.Set(configuration.ViperKeySelfServiceRecoveryUI, ts.URL+"/recovery-ts")
	t.Cleanup(ts.Close)
	return ts
}

func GetRecoveryFlow(t *testing.T, client *http.Client, ts *httptest.Server) *public.GetSelfServiceRecoveryFlowOK {
	publicClient := NewSDKClient(ts)

	res, err := client.Get(ts.URL + recovery.RouteInitBrowserFlow)
	require.NoError(t, err)
	require.NoError(t, res.Body.Close())

	rs, err := publicClient.Public.GetSelfServiceRecoveryFlow(
		public.NewGetSelfServiceRecoveryFlowParams().WithHTTPClient(client).
			WithID(res.Request.URL.Query().Get("flow")),
	)
	require.NoError(t, err, "%s", res.Request.URL.String())
	assert.Empty(t, rs.Payload.Active)

	return rs
}

func RecoverySubmitForm(
	t *testing.T,
	f *models.RecoveryFlowMethodConfig,
	hc *http.Client,
	values url.Values,
) (string, *public.GetSelfServiceRecoveryFlowOK) {
	require.NotEmpty(t, f.Action)

	res, err := hc.PostForm(pointerx.StringR(f.Action), values)
	require.NoError(t, err)
	defer res.Body.Close()

	b, err := ioutil.ReadAll(res.Body)
	require.NoError(t, err)
	assert.EqualValues(t, http.StatusOK, res.StatusCode, "%s", b)

	assert.Equal(t, viper.GetString(configuration.ViperKeySelfServiceRecoveryUI), res.Request.URL.Scheme+"://"+res.Request.URL.Host+res.Request.URL.Path, "should end up at the settings URL, used: %s", pointerx.StringR(f.Action))

	rs, err := NewSDKClientFromURL(viper.GetString(configuration.ViperKeyPublicBaseURL)).Public.GetSelfServiceRecoveryFlow(
		public.NewGetSelfServiceRecoveryFlowParams().WithHTTPClient(hc).
			WithID(res.Request.URL.Query().Get("flow")),
	)
	require.NoError(t, err)
	body, err := json.Marshal(rs.Payload)
	require.NoError(t, err)
	return string(body), rs
}

func InitializeRecoveryFlowViaBrowser(t *testing.T, client *http.Client, ts *httptest.Server) *public.GetSelfServiceRecoveryFlowOK {
	publicClient := NewSDKClient(ts)
	res, err := client.Get(ts.URL + recovery.RouteInitBrowserFlow)
	require.NoError(t, err)
	require.NoError(t, res.Body.Close())

	rs, err := publicClient.Public.GetSelfServiceRecoveryFlow(
		public.NewGetSelfServiceRecoveryFlowParams().WithHTTPClient(client).
			WithID(res.Request.URL.Query().Get("flow")),
	)
	require.NoError(t, err)
	assert.Empty(t, rs.Payload.Active)

	return rs
}

func InitializeRecoveryFlowViaAPI(t *testing.T, client *http.Client, ts *httptest.Server) *public.InitializeSelfServiceRecoveryViaAPIFlowOK {
	publicClient := NewSDKClient(ts)

	rs, err := publicClient.Public.InitializeSelfServiceRecoveryViaAPIFlow(public.
		NewInitializeSelfServiceRecoveryViaAPIFlowParams().WithHTTPClient(client))
	require.NoError(t, err)
	assert.Empty(t, rs.Payload.Active)

	return rs
}

func GetRecoveryFlowMethodConfig(t *testing.T, rs *models.RecoveryFlow, id string) *models.RecoveryFlowMethodConfig {
	require.NotEmpty(t, rs.Methods[id])
	require.NotEmpty(t, rs.Methods[id].Config)
	require.NotEmpty(t, rs.Methods[id].Config.Action)
	return rs.Methods[id].Config
}

func RecoveryMakeRequest(
	t *testing.T,
	isAPI bool,
	f *models.RecoveryFlowMethodConfig,
	hc *http.Client,
	values string,
) (string, *http.Response) {
	require.NotEmpty(t, f.Action)

	res, err := hc.Do(NewRequest(t, isAPI, "POST", pointerx.StringR(f.Action), bytes.NewBufferString(values)))
	require.NoError(t, err)
	defer res.Body.Close()

	return string(x.MustReadAll(res.Body)), res
}

// SubmitRecoveryForm initiates a registration flow (for Browser and API!), fills out the form and modifies
// the form values with `withValues`, and submits the form. If completed, it will return the flow as JSON.
func SubmitRecoveryForm(
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
	var f *models.RecoveryFlow
	if isAPI {
		f = InitializeRecoveryFlowViaAPI(t, hc, publicTS).Payload
	} else {
		f = InitializeRecoveryFlowViaBrowser(t, hc, publicTS).Payload
	}

	time.Sleep(time.Millisecond) // add a bit of delay to allow `1ns` to time out.

	config := GetRecoveryFlowMethodConfig(t, f, method.String())
	formPayload := SDKFormFieldsToURLValues(config.Fields)
	withValues(formPayload)

	b, res := RecoveryMakeRequest(t, isAPI, config, hc, EncodeFormAsJSON(t, isAPI, formPayload))
	assert.EqualValues(t, expectedStatusCode, res.StatusCode, "%s", b)
	assert.Contains(t, res.Request.URL.String(), expectedURL, "%+v\n\t%s", res.Request, b)

	return b
}
