package testhelpers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/viper"
	"github.com/ory/x/pointerx"

	"github.com/ory/kratos/driver/configuration"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/internal/httpclient/client/common"
	"github.com/ory/kratos/internal/httpclient/models"
	"github.com/ory/kratos/selfservice/flow/registration"
	"github.com/ory/kratos/selfservice/strategy/password"
	"github.com/ory/kratos/x"
)

func InitializeRegistrationFlowViaBrowser(t *testing.T, client *http.Client, ts *httptest.Server) *common.GetSelfServiceRegistrationFlowOK {
	res, err := client.Get(ts.URL + registration.RouteInitBrowserFlow)
	require.NoError(t, err)
	require.NoError(t, res.Body.Close())

	rs, err := NewSDKClient(ts).Common.GetSelfServiceRegistrationFlow(
		common.NewGetSelfServiceRegistrationFlowParams().WithHTTPClient(client).
			WithID(res.Request.URL.Query().Get("flow")))
	require.NoError(t, err)
	assert.Empty(t, rs.Payload.Active)

	return rs
}

func InitializeRegistrationFlowViaAPI(t *testing.T, client *http.Client, ts *httptest.Server) *common.InitializeSelfServiceRegistrationViaAPIFlowOK {
	rs, err := NewSDKClient(ts).Common.InitializeSelfServiceRegistrationViaAPIFlow(common.
		NewInitializeSelfServiceRegistrationViaAPIFlowParams().WithHTTPClient(client))
	require.NoError(t, err)
	assert.Empty(t, rs.Payload.Active)

	return rs
}

func GetRegistrationFlowMethodConfig(t *testing.T, rs *models.RegistrationFlow, id string) *models.RegistrationFlowMethodConfig {
	require.NotEmpty(t, rs.Methods[id])
	require.NotEmpty(t, rs.Methods[id].Config)
	require.NotEmpty(t, rs.Methods[id].Config.Action)
	return rs.Methods[id].Config
}

func RegistrationMakeRequest(
	t *testing.T,
	isAPI bool,
	f *models.RegistrationFlowMethodConfig,
	hc *http.Client,
	values string,
) (string, *http.Response) {
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

	return string(x.MustReadAll(res.Body)), res
}

// SubmitRegistrationFormAndExpectValidationError initiates a registration flow (for Browser and API!), fills out the form and modifies
// the form values with `withValues`, and submits the form. If completed, it will return the flow as JSON.
func SubmitRegistrationFormAndExpectValidationError(
	t *testing.T,
	isAPI bool,
	hc *http.Client,
	publicTS *httptest.Server,
	withValues func(v url.Values) url.Values,
	method identity.CredentialsType,
	expectedStatusCode int,
) string {
	hc.Transport = NewTransportWithLogger(hc.Transport, t)
	var payload *models.RegistrationFlow
	if isAPI {
		payload = InitializeRegistrationFlowViaAPI(t, hc, publicTS).Payload
	} else {
		payload = InitializeRegistrationFlowViaBrowser(t, hc, publicTS).Payload
	}

	time.Sleep(time.Millisecond) // add a bit of delay to allow `1ns` to time out.

	config := GetRegistrationFlowMethodConfig(t, payload, method.String())

	b, res := RegistrationMakeRequest(t, isAPI, config, hc, EncodeFormAsJSON(t, isAPI,
		withValues(SDKFormFieldsToURLValues(config.Fields))))
	assert.EqualValues(t, expectedStatusCode, res.StatusCode, "%s", b)

	expectURL := viper.GetString(configuration.ViperKeySelfServiceRegistrationUI)
	if isAPI {
		switch method {
		case identity.CredentialsTypePassword:
			expectURL = password.RouteRegistration
		default:
			t.Logf("Expected method to be profile ior password but got: %s", method)
			t.FailNow()
		}
	}

	assert.Contains(t, res.Request.URL.String(), expectURL, "%+v\n\t%s", res.Request, b)

	if isAPI {
		return b
	}

	rs, err := NewSDKClientFromURL(viper.GetString(configuration.ViperKeyPublicBaseURL)).Common.GetSelfServiceRegistrationFlow(
		common.NewGetSelfServiceRegistrationFlowParams().WithHTTPClient(hc).WithID(res.Request.URL.Query().Get("flow")))
	require.NoError(t, err)

	body, err := json.Marshal(rs.Payload)
	require.NoError(t, err)
	return string(body)
}
