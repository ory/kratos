// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package testhelpers

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/tidwall/gjson"

	kratos "github.com/ory/kratos/internal/httpclient"

	"github.com/ory/x/ioutilx"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/driver"
	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/selfservice/flow/recovery"
	"github.com/ory/kratos/x"
)

func NewVerifyAfterHookWebHookTarget(ctx context.Context, t *testing.T, conf *config.Config, assert func(t *testing.T, body []byte)) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		msg, err := io.ReadAll(r.Body)
		require.NoError(t, err)

		assert(t, msg)
	}))
	before := conf.GetProvider(ctx).Get(config.ViperKeySelfServiceVerificationAfter + ".hooks")
	// A hook to ensure that the verification hook is called with the correct data
	conf.MustSet(ctx, config.ViperKeySelfServiceVerificationAfter+".hooks", []map[string]interface{}{
		{
			"hook": "web_hook",
			"config": map[string]interface{}{
				"url":    ts.URL,
				"method": "POST",
				"body":   "base64://ZnVuY3Rpb24oY3R4KSB7CiAgICBpZGVudGl0eTogY3R4LmlkZW50aXR5Cn0=",
			},
		},
	})

	t.Cleanup(ts.Close)
	t.Cleanup(func() {
		conf.MustSet(ctx, config.ViperKeySelfServiceVerificationAfter+".hooks", before)
	})
}

func NewRecoveryAfterHookWebHookTarget(ctx context.Context, t *testing.T, conf *config.Config, assert func(t *testing.T, body []byte)) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		msg, err := io.ReadAll(r.Body)
		require.NoError(t, err)

		assert(t, msg)
	}))

	// A hook to ensure that the recovery hook is called with the correct data
	conf.MustSet(ctx, config.ViperKeySelfServiceRecoveryAfter+".hooks", []map[string]interface{}{
		{
			"hook": "web_hook",
			"config": map[string]interface{}{
				"url":    ts.URL,
				"method": "POST",
				"body":   "base64://ZnVuY3Rpb24oY3R4KSB7CiAgICBpZGVudGl0eTogY3R4LmlkZW50aXR5Cn0=",
			},
		},
	})

	t.Cleanup(ts.Close)
	t.Cleanup(func() {
		conf.MustSet(ctx, config.ViperKeySelfServiceRecoveryAfter+".hooks", []map[string]interface{}{})
	})
}

func NewRecoveryUIFlowEchoServer(t *testing.T, reg driver.Registry) *httptest.Server {
	ctx := context.Background()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		e, err := reg.RecoveryFlowPersister().GetRecoveryFlow(r.Context(), x.ParseUUID(r.URL.Query().Get("flow")))
		require.NoError(t, err)
		reg.Writer().Write(w, r, e)
	}))
	reg.Config().MustSet(ctx, config.ViperKeySelfServiceRecoveryUI, ts.URL+"/recovery-ts")
	t.Cleanup(ts.Close)
	return ts
}

func GetRecoveryFlowForType(t *testing.T, client *http.Client, ts *httptest.Server, ft flow.Type) *kratos.RecoveryFlow {
	publicClient := NewSDKCustomClient(ts, client)

	var url string
	switch ft {
	case flow.TypeBrowser:
		url = ts.URL + recovery.RouteInitBrowserFlow
	case flow.TypeAPI:
		url = ts.URL + recovery.RouteInitAPIFlow
	default:
		t.Errorf("unknown type: %s", ft)
		t.FailNow()
	}

	res, err := client.Get(url)
	require.NoError(t, err)
	defer func() { _ = res.Body.Close() }()

	var flowID string
	switch ft {
	case flow.TypeBrowser:
		flowID = res.Request.URL.Query().Get("flow")
	case flow.TypeAPI:
		flowID = gjson.GetBytes(ioutilx.MustReadAll(res.Body), "id").String()
	default:
		t.Errorf("unknown type: %s", ft)
		t.FailNow()
	}
	require.NotEmpty(t, flowID, "expected to receive a flow id, got none. %s", ioutilx.MustReadAll(res.Body))

	rs, _, err := publicClient.FrontendAPI.GetRecoveryFlow(context.Background()).
		Id(flowID).
		Execute()
	require.NoError(t, err, "expected no error when fetching recovery flow: %s", err)
	assert.NotEmpty(t, rs.Active)

	return rs
}

func GetRecoveryFlow(t *testing.T, client *http.Client, ts *httptest.Server) *kratos.RecoveryFlow {
	return GetRecoveryFlowForType(t, client, ts, flow.TypeBrowser)
}

func InitializeRecoveryFlowViaBrowser(t *testing.T, client *http.Client, isSPA bool, ts *httptest.Server, values url.Values) *kratos.RecoveryFlow {
	publicClient := NewSDKCustomClient(ts, client)

	u := ts.URL + recovery.RouteInitBrowserFlow
	if values != nil {
		u += "?" + values.Encode()
	}
	req, err := http.NewRequest("GET", u, nil)
	require.NoError(t, err)

	if isSPA {
		req.Header.Set("Accept", "application/json")
	}

	res, err := client.Do(req)
	require.NoError(t, err)
	defer func() { _ = res.Body.Close() }()

	if isSPA {
		var f kratos.RecoveryFlow
		require.NoError(t, json.NewDecoder(res.Body).Decode(&f))
		return &f
	}

	require.NoError(t, res.Body.Close())
	rs, _, err := publicClient.FrontendAPI.GetRecoveryFlow(context.Background()).Id(res.Request.URL.Query().Get("flow")).Execute()
	require.NoError(t, err)
	assert.NotEmpty(t, rs.Active)

	return rs
}

func InitializeRecoveryFlowViaAPI(t *testing.T, client *http.Client, ts *httptest.Server) *kratos.RecoveryFlow {
	publicClient := NewSDKCustomClient(ts, client)

	rs, _, err := publicClient.FrontendAPI.CreateNativeRecoveryFlow(context.Background()).Execute()
	require.NoError(t, err)
	assert.NotEmpty(t, rs.Active)

	return rs
}

func RecoveryMakeRequest(
	t *testing.T,
	isAPI bool,
	f *kratos.RecoveryFlow,
	hc *http.Client,
	values string,
) (string, *http.Response) {
	require.NotEmpty(t, f.Ui.Action)

	res, err := hc.Do(NewRequest(t, isAPI, "POST", f.Ui.Action, bytes.NewBufferString(values)))
	require.NoError(t, err)
	defer func() { _ = res.Body.Close() }()

	return string(ioutilx.MustReadAll(res.Body)), res
}

// SubmitRecoveryForm initiates a registration flow (for Browser and API!), fills out the form and modifies
// the form values with `withValues`, and submits the form. If completed, it will return the flow as JSON.
func SubmitRecoveryForm(
	t *testing.T,
	isAPI bool,
	isSPA bool,
	hc *http.Client,
	publicTS *httptest.Server,
	withValues func(v url.Values),
	expectedStatusCode int,
	expectedURL string,
) string {
	hc.Transport = NewTransportWithLogger(hc.Transport, t)
	var f *kratos.RecoveryFlow
	if isAPI {
		f = InitializeRecoveryFlowViaAPI(t, hc, publicTS)
	} else {
		f = InitializeRecoveryFlowViaBrowser(t, hc, isSPA, publicTS, nil)
	}

	time.Sleep(time.Millisecond) // add a bit of delay to allow `1ns` to time out.

	formPayload := SDKFormFieldsToURLValues(f.Ui.Nodes)
	withValues(formPayload)

	b, res := RecoveryMakeRequest(t, isAPI || isSPA, f, hc, EncodeFormAsJSON(t, isAPI || isSPA, formPayload))
	assert.EqualValues(t, expectedStatusCode, res.StatusCode, "%s", b)
	assert.Contains(t, res.Request.URL.String(), expectedURL, "%+v\n\t%s", res.Request, b)

	return b
}

func PersistNewRecoveryFlow(t *testing.T, strategy recovery.Strategy, conf *config.Config, reg *driver.RegistryDefault) *recovery.Flow {
	t.Helper()
	req := NewTestHTTPRequest(t, "GET", conf.SelfPublicURL(context.Background()).String()+"/test", nil)
	f, err := recovery.NewFlow(conf, conf.SelfServiceFlowRecoveryRequestLifespan(context.Background()), reg.GenerateCSRFToken(req), req, strategy, flow.TypeBrowser)
	require.NoError(t, err, "Expected no error when creating a new recovery flow: %s", err)

	err = reg.RecoveryFlowPersister().CreateRecoveryFlow(context.Background(), f)
	require.NoError(t, err, "Expected no error when persisting a new recover flow: %s", err)
	return f
}
