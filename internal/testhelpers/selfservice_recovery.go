// Copyright © 2022 Ory Corp
// SPDX-License-Identifier: Apache-2.0

// nolint
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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/driver"
	"github.com/ory/kratos/driver/config"
	kratos "github.com/ory/kratos/internal/httpclient"
	"github.com/ory/kratos/selfservice/flow/verification"
	"github.com/ory/kratos/x"
	"github.com/ory/x/ioutilx"
)

func NewVerificationUIFlowEchoServer(t *testing.T, reg driver.Registry) *httptest.Server {
	return createVerificationUIFlowEchoServer(t, reg, func(w http.ResponseWriter, r *http.Request) {
		e, err := reg.VerificationFlowPersister().GetVerificationFlow(r.Context(), x.ParseUUID(r.URL.Query().Get("flow")))
		require.NoError(t, err)
		reg.Writer().Write(w, r, e)
	})
}
func NewVerificationUIFlowEchoServerRespectAcceptHeader(t *testing.T, reg driver.Registry) *httptest.Server {
	return createVerificationUIFlowEchoServer(t, reg, func(w http.ResponseWriter, r *http.Request) {
		e, err := reg.VerificationFlowPersister().GetVerificationFlow(r.Context(), x.ParseUUID(r.URL.Query().Get("flow")))
		require.NoError(t, err)
		if x.AcceptsJSON(r) {
			reg.Writer().Write(w, r, e)
		} else {
			type Html struct {
				Flow *verification.Flow `json:"browser_flow"`
			}
			reg.Writer().Write(w, r, Html{Flow: e})
		}
	})
}

func createVerificationUIFlowEchoServer(t *testing.T, reg driver.Registry, handler http.HandlerFunc) *httptest.Server {
	ctx := context.Background()
	ts := httptest.NewServer(handler)
	reg.Config().MustSet(ctx, config.ViperKeySelfServiceVerificationUI, ts.URL+"/verification-ts")
	t.Cleanup(ts.Close)
	return ts
}

func GetVerificationFlow(t *testing.T, client *http.Client, isAPI bool, ts *httptest.Server) *kratos.VerificationFlow {
	publicClient := NewSDKCustomClient(ts, client)

	var initFlowRoute string
	if isAPI {
		initFlowRoute = verification.RouteInitAPIFlow
	} else {
		initFlowRoute = verification.RouteInitBrowserFlow
	}
	res, err := client.Get(ts.URL + initFlowRoute)
	require.NoError(t, err)
	defer res.Body.Close()

	var rs kratos.VerificationFlow
	if isAPI {
		require.NoError(t, json.NewDecoder(res.Body).Decode(&rs))
	} else {
		f, _, err := publicClient.FrontendApi.GetVerificationFlow(context.Background()).Id(res.Request.URL.Query().Get("flow")).Execute()
		require.NoError(t, err, "%s", res.Request.URL.String())
		rs = *f
	}
	assert.NotEmpty(t, rs.Active)

	return &rs
}

func InitializeVerificationFlowViaBrowser(t *testing.T, client *http.Client, isSPA bool, ts *httptest.Server) *kratos.VerificationFlow {
	publicClient := NewSDKCustomClient(ts, client)
	req, err := http.NewRequest("GET", ts.URL+verification.RouteInitBrowserFlow, nil)
	require.NoError(t, err)

	if isSPA {
		req.Header.Set("Accept", "application/json")
	}
	require.NoError(t, err)

	res, err := client.Do(req)
	require.NoError(t, err)
	defer res.Body.Close()

	if isSPA {
		var f kratos.VerificationFlow
		require.NoError(t, json.NewDecoder(res.Body).Decode(&f))
		return &f
	}

	rs, _, err := publicClient.FrontendApi.GetVerificationFlow(context.Background()).Id(res.Request.URL.Query().Get("flow")).Execute()
	require.NoError(t, err)
	assert.NotEmpty(t, rs.Active)

	return rs
}

func InitializeVerificationFlowViaAPI(t *testing.T, client *http.Client, ts *httptest.Server) *kratos.VerificationFlow {
	publicClient := NewSDKCustomClient(ts, client)

	rs, _, err := publicClient.FrontendApi.CreateNativeVerificationFlow(context.Background()).Execute()
	require.NoError(t, err)
	assert.NotEmpty(t, rs.Active)

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
	isSPA bool,
	hc *http.Client,
	publicTS *httptest.Server,
	withValues func(v url.Values),
	expectedStatusCode int,
	expectedURL string,
) string {
	var f *kratos.VerificationFlow
	if isAPI {
		f = InitializeVerificationFlowViaAPI(t, hc, publicTS)
	} else {
		f = InitializeVerificationFlowViaBrowser(t, hc, isSPA, publicTS)
	}

	time.Sleep(time.Millisecond) // add a bit of delay to allow `1ns` to time out.

	formPayload := SDKFormFieldsToURLValues(f.Ui.Nodes)
	withValues(formPayload)

	b, res := VerificationMakeRequest(t, isAPI || isSPA, f, hc, EncodeFormAsJSON(t, isAPI || isSPA, formPayload))
	assert.EqualValues(t, expectedStatusCode, res.StatusCode, "%s", b)
	assert.Contains(t, res.Request.URL.String(), expectedURL, "%+v\n\t%s", res.Request, b)

	return b
}
