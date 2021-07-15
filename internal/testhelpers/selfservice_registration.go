package testhelpers

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/tidwall/gjson"

	"github.com/ory/x/assertx"

	kratos "github.com/ory/kratos-client-go"

	"github.com/ory/x/ioutilx"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/driver"
	"github.com/ory/kratos/driver/config"
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

func InitializeRegistrationFlowViaBrowser(t *testing.T, client *http.Client, ts *httptest.Server, isSPA bool) *kratos.SelfServiceRegistrationFlow {
	req, err := http.NewRequest("GET", ts.URL+registration.RouteInitBrowserFlow, nil)
	require.NoError(t, err)

	if isSPA {
		req.Header.Set("Accept", "application/json")
	}

	res, err := client.Do(req)
	require.NoError(t, err)
	body := x.MustReadAll(res.Body)
	require.NoError(t, res.Body.Close())

	flowID := res.Request.URL.Query().Get("flow")
	if isSPA {
		flowID = gjson.GetBytes(body, "id").String()
	}

	rs, _, err := NewSDKCustomClient(ts, client).V0alpha1Api.GetSelfServiceRegistrationFlow(context.Background()).Id(flowID).Execute()
	require.NoError(t, err)
	assert.Empty(t, rs.Active)

	return rs
}

func InitializeRegistrationFlowViaAPI(t *testing.T, client *http.Client, ts *httptest.Server) *kratos.SelfServiceRegistrationFlow {
	rs, _, err := NewSDKCustomClient(ts, client).V0alpha1Api.InitializeSelfServiceRegistrationFlowWithoutBrowser(context.Background()).Execute()
	require.NoError(t, err)
	assert.Empty(t, rs.Active)
	return rs
}

func RegistrationMakeRequest(
	t *testing.T,
	isAPI bool,
	isSPA bool,
	f *kratos.SelfServiceRegistrationFlow,
	hc *http.Client,
	values string,
) (string, *http.Response) {
	require.NotEmpty(t, f.Ui.Action)

	req := NewRequest(t, isAPI, "POST", f.Ui.Action, bytes.NewBufferString(values))
	if isSPA {
		req.Header.Set("Accept", "application/json")
	}

	res, err := hc.Do(req)
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
	isSPA bool,
	expectedStatusCode int,
	expectedURL string,
) string {
	if hc == nil {
		hc = new(http.Client)
	}

	hc.Transport = NewTransportWithLogger(hc.Transport, t)
	var payload *kratos.SelfServiceRegistrationFlow
	if isAPI {
		payload = InitializeRegistrationFlowViaAPI(t, hc, publicTS)
	} else {
		payload = InitializeRegistrationFlowViaBrowser(t, hc, publicTS, isSPA)
	}

	time.Sleep(time.Millisecond) // add a bit of delay to allow `1ns` to time out.

	values := SDKFormFieldsToURLValues(payload.Ui.Nodes)
	withValues(values)
	b, res := RegistrationMakeRequest(t, isAPI, isSPA, payload, hc, EncodeFormAsJSON(t, isAPI, values))
	assert.EqualValues(t, expectedStatusCode, res.StatusCode, assertx.PrettifyJSONPayload(t, b))
	assert.Contains(t, res.Request.URL.String(), expectedURL, "%+v\n\t%s", res.Request, assertx.PrettifyJSONPayload(t, b))
	return b
}
