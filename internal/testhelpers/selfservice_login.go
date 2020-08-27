package testhelpers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/ory/x/pointerx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/viper"

	"github.com/ory/kratos/driver"
	"github.com/ory/kratos/driver/configuration"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/internal/httpclient/client/common"
	"github.com/ory/kratos/internal/httpclient/models"
	"github.com/ory/kratos/selfservice/flow/login"
	"github.com/ory/kratos/selfservice/strategy/password"
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

func InitializeLoginFlowViaBrowser(t *testing.T, client *http.Client, ts *httptest.Server, forced bool) *common.GetSelfServiceLoginFlowOK {
	publicClient := NewSDKClient(ts)

	q := ""
	if forced {
		q = "?refresh=true"
	}

	res, err := client.Get(ts.URL + login.RouteInitBrowserFlow + q)
	require.NoError(t, err)
	require.NoError(t, res.Body.Close())

	rs, err := publicClient.Common.GetSelfServiceLoginFlow(
		common.NewGetSelfServiceLoginFlowParams().WithHTTPClient(client).
			WithID(res.Request.URL.Query().Get("flow")),
	)
	require.NoError(t, err)
	assert.Empty(t, rs.Payload.Active)

	return rs
}

func InitializeLoginFlowViaAPI(t *testing.T, client *http.Client, ts *httptest.Server, forced bool) *common.InitializeSelfServiceLoginViaAPIFlowOK {
	publicClient := NewSDKClient(ts)

	rs, err := publicClient.Common.InitializeSelfServiceLoginViaAPIFlow(common.
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

	return string(x.MustReadAll(res.Body)), res
}

// SubmitLoginFormAndExpectValidationError initiates a login flow (for Browser and API!), fills out the form and modifies
// the form values with `withValues`, and submits the form. If completed, it will return the flow as JSON.
func SubmitLoginFormAndExpectValidationError(
	t *testing.T,
	isAPI bool,
	hc *http.Client,
	publicTS *httptest.Server,
	withValues func(v url.Values) url.Values,
	method identity.CredentialsType,
	forced bool,
	expectedStatusCode int,
) string {
	hc.Transport = NewTransportWithLogger(hc.Transport, t)
	var payload *models.LoginFlow
	if isAPI {
		payload = InitializeLoginFlowViaAPI(t, hc, publicTS,forced).Payload
	} else {
		payload = InitializeLoginFlowViaBrowser(t, hc, publicTS,forced).Payload
	}

	time.Sleep(time.Millisecond) // add a bit of delay to allow `1ns` to time out.

	config := GetLoginFlowMethodConfig(t, payload, method.String())

	b, res := LoginMakeRequest(t, isAPI, config, hc, EncodeFormAsJSON(t, isAPI,
		withValues(SDKFormFieldsToURLValues(config.Fields))))
	assert.EqualValues(t, expectedStatusCode, res.StatusCode, "%s", b)

	expectURL := viper.GetString(configuration.ViperKeySelfServiceLoginUI)
	if isAPI {
		switch method {
		case identity.CredentialsTypePassword:
			expectURL = password.RouteLogin
		default:
			t.Logf("Expected method to be profile ior password but got: %s", method)
			t.FailNow()
		}
	}

	assert.Contains(t, res.Request.URL.String(), expectURL, "%+v\n\t%s", res.Request, b)

	if isAPI {
		return b
	}

	rs, err := NewSDKClientFromURL(viper.GetString(configuration.ViperKeyPublicBaseURL)).Common.GetSelfServiceLoginFlow(
		common.NewGetSelfServiceLoginFlowParams().WithHTTPClient(hc).WithID(res.Request.URL.Query().Get("flow")))
	require.NoError(t, err)

	body, err := json.Marshal(rs.Payload)
	require.NoError(t, err)
	return string(body)
}
