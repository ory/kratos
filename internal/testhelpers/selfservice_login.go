// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package testhelpers

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/ory/kratos/identity"
	"github.com/ory/x/urlx"

	"github.com/tidwall/gjson"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/driver"
	"github.com/ory/kratos/driver/config"
	kratos "github.com/ory/kratos/internal/httpclient"
	"github.com/ory/kratos/selfservice/flow/login"
	"github.com/ory/kratos/x"
	"github.com/ory/x/ioutilx"
)

func NewLoginUIFlowEchoServer(t *testing.T, reg driver.Registry) *httptest.Server {
	ctx := context.Background()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		e, err := reg.LoginFlowPersister().GetLoginFlow(r.Context(), x.ParseUUID(r.URL.Query().Get("flow")))
		require.NoError(t, err)
		reg.Writer().Write(w, r, e)
	}))
	ts.URL = strings.Replace(ts.URL, "127.0.0.1", "localhost", -1)
	reg.Config().MustSet(ctx, config.ViperKeySelfServiceLoginUI, ts.URL+"/login-ts")
	t.Cleanup(ts.Close)
	return ts
}

func NewLoginUIWith401Response(t *testing.T, c *config.Config) *httptest.Server {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	ts.URL = strings.Replace(ts.URL, "127.0.0.1", "localhost", -1)
	ctx := context.Background()
	c.MustSet(ctx, config.ViperKeySelfServiceLoginUI, ts.URL+"/login-ts")
	t.Cleanup(ts.Close)
	return ts
}

type initFlowOptions struct {
	aal                  identity.AuthenticatorAssuranceLevel
	returnTo             string
	refresh              bool
	oauth2LoginChallenge string
	via                  string
	ctx                  context.Context
}

func newInitFlowOptions(opts []InitFlowWithOption) *initFlowOptions {
	return new(initFlowOptions).apply(opts)
}

func (o *initFlowOptions) Context() context.Context {
	if o.ctx == nil {
		return context.Background()
	}
	return o.ctx
}

func (o *initFlowOptions) apply(opts []InitFlowWithOption) *initFlowOptions {
	for _, opt := range opts {
		opt(o)
	}
	return o
}

func getURLFromInitOptions(ts *httptest.Server, path string, forced bool, opts ...InitFlowWithOption) string {
	o := new(initFlowOptions).apply(opts)
	q := url.Values{}

	if forced || o.refresh {
		q.Set("refresh", "true")
	}

	if o.aal != "" {
		q.Set("aal", string(o.aal))
	}

	if o.returnTo != "" {
		q.Set("return_to", string(o.returnTo))
	}

	if o.oauth2LoginChallenge != "" {
		q.Set("login_challenge", o.oauth2LoginChallenge)
	}
	if o.via != "" {
		q.Set("via", o.via)
	}

	u := urlx.ParseOrPanic(ts.URL + path)
	u.RawQuery = q.Encode()
	return u.String()
}

type InitFlowWithOption func(*initFlowOptions)

func InitFlowWithAAL(aal identity.AuthenticatorAssuranceLevel) InitFlowWithOption {
	return func(o *initFlowOptions) {
		o.aal = aal
	}
}

func InitFlowWithReturnTo(returnTo string) InitFlowWithOption {
	return func(o *initFlowOptions) {
		o.returnTo = returnTo
	}
}

func InitFlowWithRefresh() InitFlowWithOption {
	return func(o *initFlowOptions) {
		o.refresh = true
	}
}

func InitFlowWithContext(ctx context.Context) InitFlowWithOption {
	return func(o *initFlowOptions) {
		o.ctx = ctx
	}
}

func InitFlowWithOAuth2LoginChallenge(hlc string) InitFlowWithOption {
	return func(o *initFlowOptions) {
		o.oauth2LoginChallenge = hlc
	}
}

// InitFlowWithVia sets the `via` query parameter which is used by the code MFA flows to determine the trait to use to send the code to the user
func InitFlowWithVia(via string) InitFlowWithOption {
	return func(o *initFlowOptions) {
		o.via = via
	}
}

func InitializeLoginFlowViaBrowser(t *testing.T, client *http.Client, ts *httptest.Server, forced bool, isSPA bool, expectInitError bool, expectGetError bool, opts ...InitFlowWithOption) *kratos.LoginFlow {
	publicClient := NewSDKCustomClient(ts, client)

	req, err := http.NewRequest("GET", getURLFromInitOptions(ts, login.RouteInitBrowserFlow, forced, opts...), nil)
	require.NoError(t, err)
	o := newInitFlowOptions(opts)

	if isSPA {
		req.Header.Set("Accept", "application/json")
	}

	res, err := client.Do(req.WithContext(o.Context()))
	require.NoError(t, err)
	body := x.MustReadAll(res.Body)
	require.NoError(t, res.Body.Close())
	require.Equal(t, 200, res.StatusCode, "%s", body)
	if expectInitError {
		require.NotNil(t, res.Request.URL)
		require.Contains(t, res.Request.URL.String(), "error-ts")
	}

	flowID := res.Request.URL.Query().Get("flow")
	if isSPA {
		flowID = gjson.GetBytes(body, "id").String()
	}
	require.NotEmpty(t, flowID)

	rs, r, err := publicClient.FrontendAPI.GetLoginFlow(context.Background()).Id(flowID).Execute()
	if expectGetError {
		require.Error(t, err)
		require.Nil(t, rs)
	} else {
		require.NoError(t, err, "%s", ioutilx.MustReadAll(r.Body))
		assert.Empty(t, rs.Active)
	}

	return rs
}

func InitializeLoginFlowViaAPIWithContext(t *testing.T, ctx context.Context, client *http.Client, ts *httptest.Server, forced bool, opts ...InitFlowWithOption) *kratos.LoginFlow {
	publicClient := NewSDKCustomClient(ts, client)

	o := new(initFlowOptions).apply(opts)
	req := publicClient.FrontendAPI.CreateNativeLoginFlow(ctx).Refresh(forced)
	if o.aal != "" {
		req = req.Aal(string(o.aal))
	}
	if o.via != "" {
		req = req.Via(o.via)
	}

	rs, res, err := req.Execute()
	require.NoError(t, err, "%s", ioutilx.MustReadAll(res.Body))
	assert.Empty(t, rs.Active)

	return rs
}

func InitializeLoginFlowViaAPI(t *testing.T, client *http.Client, ts *httptest.Server, forced bool, opts ...InitFlowWithOption) *kratos.LoginFlow {
	return InitializeLoginFlowViaAPIWithContext(t, context.Background(), client, ts, forced, opts...)
}

func LoginMakeRequest(
	t *testing.T,
	isAPI bool,
	isSPA bool,
	f *kratos.LoginFlow,
	hc *http.Client,
	values string,
) (string, *http.Response) {
	return LoginMakeRequestWithContext(t, context.Background(), isAPI, isSPA, f, hc, values)
}

func LoginMakeRequestWithContext(
	t *testing.T,
	ctx context.Context,
	isAPI bool,
	isSPA bool,
	f *kratos.LoginFlow,
	hc *http.Client,
	values string,
) (string, *http.Response) {
	require.NotEmpty(t, f.Ui.Action)

	req := NewRequest(t, isAPI, "POST", f.Ui.Action, bytes.NewBufferString(values))
	if isSPA && !isAPI {
		req.Header.Set("Accept", "application/json")
	}

	res, err := hc.Do(req.WithContext(ctx))
	require.NoError(t, err, "action: %s", f.Ui.Action)
	defer res.Body.Close()

	return string(ioutilx.MustReadAll(res.Body)), res
}

func GetLoginFlow(t *testing.T, client *http.Client, ts *httptest.Server, flowID string) *kratos.LoginFlow {
	publicClient := NewSDKCustomClient(ts, client)
	rs, _, err := publicClient.FrontendAPI.GetLoginFlow(context.Background()).Id(flowID).Execute()
	require.NoError(t, err)
	return rs
}

// SubmitLoginForm initiates a login flow (for Browser and API!), fills out the form and modifies
// the form values with `withValues`, and submits the form. Returns the body and checks for expectedStatusCode and
// expectedURL on completion
func SubmitLoginForm(
	t *testing.T,
	isAPI bool,
	hc *http.Client,
	publicTS *httptest.Server,
	withValues func(v url.Values),
	isSPA bool,
	forced bool,
	expectedStatusCode int,
	expectedURL string,
) string {
	if hc == nil {
		hc = new(http.Client)
		if !isAPI {
			hc = NewClientWithCookies(t)
		}
	}

	hc.Transport = NewTransportWithLogger(hc.Transport, t)
	var f *kratos.LoginFlow
	if isAPI {
		f = InitializeLoginFlowViaAPI(t, hc, publicTS, forced)
	} else {
		f = InitializeLoginFlowViaBrowser(t, hc, publicTS, forced, isSPA, false, false)
	}

	time.Sleep(time.Millisecond) // add a bit of delay to allow `1ns` to time out.

	payload := SDKFormFieldsToURLValues(f.Ui.Nodes)
	withValues(payload)
	b, res := LoginMakeRequest(t, isAPI, isSPA, f, hc, EncodeFormAsJSON(t, isAPI, payload))
	assert.EqualValues(t, expectedStatusCode, res.StatusCode, "%s", b)
	assert.Contains(t, res.Request.URL.String(), expectedURL, "%+v\n\t%s", res.Request, b)

	t.Logf("%+v", res.Header)

	return b
}
