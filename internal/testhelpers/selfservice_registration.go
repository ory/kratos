// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package testhelpers

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/ory/kratos/selfservice/flow/registration"

	"github.com/tidwall/gjson"

	"github.com/ory/x/assertx"

	kratos "github.com/ory/kratos/internal/httpclient"

	"github.com/ory/x/ioutilx"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/driver"
	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/x"
)

func NewRegistrationUIFlowEchoServer(t *testing.T, reg driver.Registry) *httptest.Server {
	ctx := context.Background()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		e, err := reg.RegistrationFlowPersister().GetRegistrationFlow(r.Context(), x.ParseUUID(r.URL.Query().Get("flow")))
		require.NoError(t, err)
		reg.Writer().Write(w, r, e)
	}))
	reg.Config().MustSet(ctx, config.ViperKeySelfServiceRegistrationUI, ts.URL+"/registration-ts")
	t.Cleanup(ts.Close)
	return ts
}

// Deprecated: Use context-based InitializeRegistrationFlowViaBrowserCtx instead
func InitializeRegistrationFlowViaBrowser(t *testing.T, client *http.Client, ts *httptest.Server, isSPA bool, expectInitError bool, expectGetError bool, opts ...InitFlowWithOption) *kratos.RegistrationFlow {
	return InitializeRegistrationFlowViaBrowserCtx(t.Context(), t, client, ts, isSPA, expectInitError, expectGetError, opts...)
}

func InitializeRegistrationFlowViaBrowserCtx(ctx context.Context, t *testing.T, client *http.Client, ts *httptest.Server, isSPA bool, expectInitError bool, expectGetError bool, opts ...InitFlowWithOption) *kratos.RegistrationFlow {
	req, err := http.NewRequest("GET", getURLFromInitOptions(ts, registration.RouteInitBrowserFlow, false, opts...), nil)
	require.NoError(t, err)
	req = req.WithContext(ctx)

	if isSPA {
		req.Header.Set("Accept", "application/json")
	}

	res, err := client.Do(req)
	require.NoError(t, err)
	body := x.MustReadAll(res.Body)
	if isSPA {
		require.True(t, gjson.ValidBytes(body), "body is not valid JSON: %s", string(body))
	}
	require.NoError(t, res.Body.Close())
	if expectInitError {
		require.Equal(t, 200, res.StatusCode)
		require.NotNil(t, res.Request.URL)
		require.Contains(t, res.Request.URL.String(), "error-ts")
	}

	flowID := res.Request.URL.Query().Get("flow")
	if isSPA {
		flowID = gjson.GetBytes(body, "id").String()
	}

	rs, _, err := NewSDKCustomClient(ts, client).FrontendAPI.GetRegistrationFlow(ctx).Id(flowID).Execute()
	if expectGetError {
		require.Error(t, err)
		require.Nil(t, rs)
	} else {
		require.NoError(t, err)
		assert.Empty(t, rs.Active)
	}
	return rs
}

func InitializeRegistrationFlowViaAPIExpectError(t *testing.T, client *http.Client, ts *httptest.Server, opts ...InitFlowWithOption) {
	o := new(initFlowOptions).apply(opts)

	_, _, err := NewSDKCustomClient(ts, client).FrontendAPI.CreateNativeRegistrationFlow(context.Background()).IdentitySchema(o.identitySchema).Execute()
	require.Error(t, err)
}

// Deprecated: Use context-based InitializeRegistrationFlowViaAPICtx instead
func InitializeRegistrationFlowViaAPI(t *testing.T, client *http.Client, ts *httptest.Server, opts ...InitFlowWithOption) *kratos.RegistrationFlow {
	return InitializeRegistrationFlowViaAPICtx(context.Background(), t, client, ts, opts...)
}

func InitializeRegistrationFlowViaAPICtx(ctx context.Context, t *testing.T, client *http.Client, ts *httptest.Server, opts ...InitFlowWithOption) *kratos.RegistrationFlow {
	o := new(initFlowOptions).apply(opts)

	rs, _, err := NewSDKCustomClient(ts, client).FrontendAPI.CreateNativeRegistrationFlow(ctx).IdentitySchema(o.identitySchema).Execute()
	require.NoError(t, err)
	assert.Empty(t, rs.Active)
	return rs
}

func GetRegistrationFlow(t *testing.T, client *http.Client, ts *httptest.Server, flowID string) *kratos.RegistrationFlow {
	rs, _, err := NewSDKCustomClient(ts, client).FrontendAPI.GetRegistrationFlow(context.Background()).Id(flowID).Execute()
	require.NoError(t, err)
	assert.Empty(t, rs.Active)
	return rs
}

func RegistrationMakeRequest(
	t *testing.T,
	isAPI bool,
	isSPA bool,
	f *kratos.RegistrationFlow,
	hc *http.Client,
	values string,
) (string, *http.Response) {
	return RegistrationMakeRequestCtx(t.Context(), t, isAPI, isSPA, f, hc, values)
}

func RegistrationMakeRequestCtx(
	ctx context.Context,
	t *testing.T,
	isAPI bool,
	isSPA bool,
	f *kratos.RegistrationFlow,
	hc *http.Client,
	values string,
) (string, *http.Response) {
	require.NotEmpty(t, f.Ui.Action)

	req := NewRequest(t, isAPI, "POST", f.Ui.Action, bytes.NewBufferString(values))
	if isSPA {
		req.Header.Set("Accept", "application/json")
	}
	req = req.WithContext(ctx)

	res, err := hc.Do(req)
	require.NoError(t, err)
	defer func() { _ = res.Body.Close() }()

	return string(ioutilx.MustReadAll(res.Body)), res
}

// Deprecated: Use context-based SubmitRegistrationFormCtx instead
func SubmitRegistrationForm(
	t *testing.T,
	isAPI bool,
	hc *http.Client,
	publicTS *httptest.Server,
	withValues func(v url.Values),
	isSPA bool,
	expectedStatusCode int,
	expectedURL string,
	opts ...InitFlowWithOption,
) string {
	return SubmitRegistrationFormCtx(context.Background(), t, isAPI, hc, publicTS, withValues, isSPA, expectedStatusCode, expectedURL, opts...)
}

// SubmitRegistrationFormCtx (for Browser and API!), fills out the form and modifies
// the form values with `withValues`, and submits the form. Returns the body and checks for expectedStatusCode and
// expectedURL on completion
func SubmitRegistrationFormCtx(
	ctx context.Context,
	t *testing.T,
	isAPI bool,
	hc *http.Client,
	publicTS *httptest.Server,
	withValues func(v url.Values),
	isSPA bool,
	expectedStatusCode int,
	expectedURL string,
	opts ...InitFlowWithOption,
) string {
	if hc == nil {
		hc = new(http.Client)
	}

	hc.Transport = NewTransportWithLogger(hc.Transport, t)
	var payload *kratos.RegistrationFlow
	if isAPI {
		payload = InitializeRegistrationFlowViaAPICtx(ctx, t, hc, publicTS, opts...)
	} else {
		payload = InitializeRegistrationFlowViaBrowserCtx(ctx, t, hc, publicTS, isSPA, false, false, opts...)
	}

	time.Sleep(time.Millisecond) // add a bit of delay to allow `1ns` to time out.

	values := SDKFormFieldsToURLValues(payload.Ui.Nodes)
	withValues(values)
	b, res := RegistrationMakeRequest(t, isAPI, isSPA, payload, hc, EncodeFormAsJSON(t, isAPI, values))
	assert.EqualValues(t, expectedStatusCode, res.StatusCode, assertx.PrettifyJSONPayload(t, b))
	assert.Containsf(t, res.Request.URL.String(), expectedURL, "%+v\n\t%s", res.Request, assertx.PrettifyJSONPayload(t, b))
	return b
}
