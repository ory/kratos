// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package testhelpers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

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

func GetRecoveryFlow(t *testing.T, client *http.Client, ts *httptest.Server) *kratos.RecoveryFlow {
	publicClient := NewSDKCustomClient(ts, client)

	res, err := client.Get(ts.URL + recovery.RouteInitBrowserFlow)
	require.NoError(t, err)
	require.NoError(t, res.Body.Close())

	flowID := res.Request.URL.Query().Get("flow")
	assert.NotEmpty(t, flowID, "expected to receive a flow id, got none")

	rs, _, err := publicClient.FrontendApi.GetRecoveryFlow(context.Background()).
		Id(flowID).
		Execute()
	assert.NotEmpty(t, rs.Active)
	require.NoError(t, err, "expected no error when fetching recovery flow: %s", err)

	return rs
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
	defer res.Body.Close()

	if isSPA {
		var f kratos.RecoveryFlow
		require.NoError(t, json.NewDecoder(res.Body).Decode(&f))
		return &f
	}

	require.NoError(t, res.Body.Close())
	rs, _, err := publicClient.FrontendApi.GetRecoveryFlow(context.Background()).Id(res.Request.URL.Query().Get("flow")).Execute()
	require.NoError(t, err)
	assert.NotEmpty(t, rs.Active)

	return rs
}

func InitializeRecoveryFlowViaAPI(t *testing.T, client *http.Client, ts *httptest.Server) *kratos.RecoveryFlow {
	publicClient := NewSDKCustomClient(ts, client)

	rs, _, err := publicClient.FrontendApi.CreateNativeRecoveryFlow(context.Background()).Execute()
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
	defer res.Body.Close()

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
	req := x.NewTestHTTPRequest(t, "GET", conf.SelfPublicURL(context.Background()).String()+"/test", nil)
	f, err := recovery.NewFlow(conf, conf.SelfServiceFlowRecoveryRequestLifespan(context.Background()), reg.GenerateCSRFToken(req), req, strategy, flow.TypeBrowser)
	require.NoError(t, err, "Expected no error when creating a new recovery flow: %s", err)

	err = reg.RecoveryFlowPersister().CreateRecoveryFlow(context.Background(), f)
	require.NoError(t, err, "Expected no error when persisting a new recover flow: %s", err)
	return f
}
