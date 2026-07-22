// Copyright © 2026 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package login_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
	"go.opentelemetry.io/otel/attribute"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"

	"github.com/ory/kratos/driver"
	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/pkg"
	"github.com/ory/kratos/pkg/testhelpers"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/selfservice/flow/login"
	"github.com/ory/kratos/ui/container"
	"github.com/ory/kratos/x"
	"github.com/ory/kratos/x/nosurfx"
	"github.com/ory/x/configx"
	"github.com/ory/x/otelx"
	"github.com/ory/x/sqlxx"
)

// newRegistryWithOIDCProvider returns a fast test registry configured with a
// single generic OIDC provider whose ID is "google". The populator can find
// the provider when the admin-create handler invokes it.
func newRegistryWithOIDCProvider(t *testing.T) *driver.RegistryDefault {
	t.Helper()
	_, reg := pkg.NewFastRegistryWithMocks(t,
		configx.WithValues(testhelpers.DefaultIdentitySchemaConfig("file://../../strategy/oidc/stub/registration.schema.json")),
		configx.WithValues(testhelpers.MethodEnableConfig(identity.CredentialsTypeOIDC, true)),
		configx.WithValue(
			config.ViperKeySelfServiceStrategyConfig+"."+string(identity.CredentialsTypeOIDC)+".config.providers",
			[]map[string]any{{
				"provider":      "generic",
				"id":            "google",
				"label":         "Google",
				"client_id":     "invalid",
				"client_secret": "invalid",
				"issuer_url":    "https://foobar/",
				"mapper_url":    "file://../../strategy/oidc/stub/oidc.facebook.jsonnet",
			}},
		),
	)
	return reg
}

func doAdminPost(t *testing.T, adminURL string, body any) (*http.Response, []byte) {
	t.Helper()
	b, err := json.Marshal(body)
	require.NoError(t, err)
	res, err := testhelpers.NewTestClient(t).Post(adminURL+"/admin"+login.RouteAdminCreateTestFlow, "application/json", bytes.NewReader(b))
	require.NoError(t, err)
	t.Cleanup(func() { _ = res.Body.Close() })
	respBody, err := io.ReadAll(res.Body)
	require.NoError(t, err)
	return res, respBody
}

func doDelete(t *testing.T, c *http.Client, u string) (*http.Response, []byte) {
	t.Helper()
	req, err := http.NewRequest(http.MethodDelete, u, nil)
	require.NoError(t, err)
	res, err := c.Do(req)
	require.NoError(t, err)
	t.Cleanup(func() { _ = res.Body.Close() })
	body, err := io.ReadAll(res.Body)
	require.NoError(t, err)
	return res, body
}

// newSeedTestFlow creates a minimal test login flow directly via the
// persister to support cap-enforcement seeding.
func newSeedTestFlow(t *testing.T, ctx context.Context, reg *driver.RegistryDefault, exp time.Duration) *login.Flow {
	t.Helper()
	f := &login.Flow{
		ID:              x.NewUUID(),
		Type:            flow.TypeBrowser,
		State:           flow.StateChooseMethod,
		RequestURL:      "https://example.com/self-service/login/browser",
		IssuedAt:        time.Now().UTC(),
		ExpiresAt:       time.Now().UTC().Add(exp),
		CSRFToken:       nosurfx.FakeCSRFToken,
		UI:              container.New(""),
		InternalContext: sqlxx.JSONRawMessage(`{}`),
	}
	require.NoError(t, f.SetTestContext(&login.TestContext{ProviderID: "google"}))
	require.NoError(t, reg.LoginFlowPersister().CreateLoginFlow(ctx, f))
	return f
}

func TestHandler_AdminCreateTestLoginFlow_HappyPath(t *testing.T) {
	t.Parallel()
	reg := newRegistryWithOIDCProvider(t)
	_, admin := testhelpers.NewKratosServer(t, reg)

	res, body := doAdminPost(t, admin.URL, map[string]string{"provider_id": "google"})
	require.Equal(t, http.StatusCreated, res.StatusCode, "%s", body)
	assert.NotEmpty(t, gjson.GetBytes(body, "id").String())
	assert.Equal(t, "google", gjson.GetBytes(body, "test_context.provider_id").String())
	assert.Equal(t, string(flow.StateChooseMethod), gjson.GetBytes(body, "state").String())

	// UI contains a provider submit node scoped to google.
	found := false
	for _, n := range gjson.GetBytes(body, "ui.nodes").Array() {
		if n.Get("attributes.name").String() == "provider" && n.Get("attributes.value").String() == "google" {
			found = true
			break
		}
	}
	assert.True(t, found, "expected provider=google submit node in UI: %s", body)
}

// TestHandler_AdminCreateTestLoginFlow_Tracing asserts that the handler
// emits a span recording which strategy served the request and what the
// provider lookup saw. A production incident where the wrong strategy was
// selected (and thus an empty provider list was searched) was invisible in
// traces because this path had no instrumentation.
func TestHandler_AdminCreateTestLoginFlow_Tracing(t *testing.T) {
	t.Parallel()
	reg := newRegistryWithOIDCProvider(t)

	recorder := tracetest.NewSpanRecorder()
	provider := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(recorder))
	reg.SetTracer(otelx.NewNoop().WithOTLP(provider.Tracer("test")))

	_, admin := testhelpers.NewKratosServer(t, reg)

	res, body := doAdminPost(t, admin.URL, map[string]string{"provider_id": "google"})
	require.Equal(t, http.StatusCreated, res.StatusCode, "%s", body)
	res, body = doAdminPost(t, admin.URL, map[string]string{"provider_id": "nonesuch"})
	require.Equal(t, http.StatusNotFound, res.StatusCode, "%s", body)

	var spans []sdktrace.ReadOnlySpan
	for _, sp := range recorder.Ended() {
		if sp.Name() == "login.Handler.adminCreateTestLoginFlow" {
			spans = append(spans, sp)
		}
	}
	require.Len(t, spans, 2, "expected one handler span per request")

	attrs := func(sp sdktrace.ReadOnlySpan) map[attribute.Key]attribute.Value {
		out := make(map[attribute.Key]attribute.Value, len(sp.Attributes()))
		for _, kv := range sp.Attributes() {
			out[kv.Key] = kv.Value
		}
		return out
	}

	found := attrs(spans[0])
	assert.Equal(t, "oidc", found["test_login_flow.strategy"].AsString())
	assert.Equal(t, int64(1), found["test_login_flow.available_providers"].AsInt64())
	assert.True(t, found["test_login_flow.provider_found"].AsBool())

	notFound := attrs(spans[1])
	assert.Equal(t, "oidc", notFound["test_login_flow.strategy"].AsString())
	assert.Equal(t, int64(1), notFound["test_login_flow.available_providers"].AsInt64())
	assert.False(t, notFound["test_login_flow.provider_found"].AsBool())
}

func TestHandler_AdminCreateTestLoginFlow_BadInput(t *testing.T) {
	t.Parallel()

	reg := newRegistryWithOIDCProvider(t)
	_, admin := testhelpers.NewKratosServer(t, reg)

	for _, tc := range []struct {
		name       string
		body       map[string]string
		wantStatus int
	}{
		{name: "empty provider_id", body: map[string]string{}, wantStatus: http.StatusBadRequest},
		{name: "unknown provider_id", body: map[string]string{"provider_id": "nonesuch"}, wantStatus: http.StatusNotFound},
		{name: "provider_id too long", body: map[string]string{"provider_id": strings.Repeat("a", 256)}, wantStatus: http.StatusBadRequest},
	} {
		t.Run("case="+tc.name, func(t *testing.T) {
			t.Parallel()
			res, body := doAdminPost(t, admin.URL, tc.body)
			assert.Equal(t, tc.wantStatus, res.StatusCode, "%s", body)
		})
	}
}

func TestHandler_DeleteTestLoginFlow(t *testing.T) {
	t.Parallel()

	type setup func(t *testing.T, reg *driver.RegistryDefault) *login.Flow

	seedUncaptured := func(t *testing.T, reg *driver.RegistryDefault) *login.Flow {
		return newSeedTestFlow(t, context.Background(), reg, time.Hour)
	}
	// seedCapturedMatching: captured flow whose CSRFToken matches what the
	// fake CSRF generator returns — simulates the browser that holds the
	// post-callback cookie.
	seedCapturedMatching := func(t *testing.T, reg *driver.RegistryDefault) *login.Flow {
		return newSeedCapturedTestFlow(t, context.Background(), reg, time.Hour, nosurfx.FakeCSRFToken)
	}
	// seedCapturedMismatch: captured flow with a CSRFToken that won't match,
	// simulating a request without (or with the wrong) CSRF cookie.
	seedCapturedMismatch := func(t *testing.T, reg *driver.RegistryDefault) *login.Flow {
		return newSeedCapturedTestFlow(t, context.Background(), reg, time.Hour, "some-other-token")
	}

	for _, tc := range []struct {
		name        string
		setup       setup
		wantStatus  int
		wantDeleted bool
	}{
		{
			name:        "uncaptured: 204",
			setup:       seedUncaptured,
			wantStatus:  http.StatusNoContent,
			wantDeleted: true,
		},
		{
			name:       "captured with mismatched CSRF token: 403",
			setup:      seedCapturedMismatch,
			wantStatus: http.StatusForbidden,
		},
		{
			name:        "captured with matching CSRF token: 204",
			setup:       seedCapturedMatching,
			wantStatus:  http.StatusNoContent,
			wantDeleted: true,
		},
	} {
		t.Run("case="+tc.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			reg := newRegistryWithOIDCProvider(t)
			public, _ := testhelpers.NewKratosServer(t, reg)

			f := tc.setup(t, reg)
			u := fmt.Sprintf("%s%s?id=%s", public.URL, login.RouteDeleteTestFlow, f.ID)

			req, err := http.NewRequest(http.MethodDelete, u, nil)
			require.NoError(t, err)
			res, err := public.Client().Do(req)
			require.NoError(t, err)
			t.Cleanup(func() { _ = res.Body.Close() })
			body, err := io.ReadAll(res.Body)
			require.NoError(t, err)
			require.Equal(t, tc.wantStatus, res.StatusCode, "%s", body)

			_, getErr := reg.LoginFlowPersister().GetLoginFlow(ctx, f.ID)
			if tc.wantDeleted {
				assert.Error(t, getErr, "flow should be deleted")
			} else {
				assert.NoError(t, getErr, "flow should still exist")
			}
		})
	}
}

func TestHandler_DeleteTestLoginFlow_NotATestFlow(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	_, reg := pkg.NewFastRegistryWithMocks(t)
	public, _ := testhelpers.NewKratosServer(t, reg)

	// Create a regular (non-test) login flow directly.
	f := &login.Flow{
		ID:              x.NewUUID(),
		Type:            flow.TypeBrowser,
		State:           flow.StateChooseMethod,
		RequestURL:      "https://example.com/self-service/login/browser",
		IssuedAt:        time.Now().UTC(),
		ExpiresAt:       time.Now().UTC().Add(time.Hour),
		CSRFToken:       nosurfx.FakeCSRFToken,
		UI:              container.New(""),
		InternalContext: sqlxx.JSONRawMessage(`{}`),
	}
	require.NoError(t, reg.LoginFlowPersister().CreateLoginFlow(ctx, f))

	u := fmt.Sprintf("%s%s?id=%s", public.URL, login.RouteDeleteTestFlow, f.ID)
	res, body := doDelete(t, public.Client(), u)
	assert.Equal(t, http.StatusNotFound, res.StatusCode, "%s", body)
}

func TestHandler_DeleteTestLoginFlow_InvalidUUID(t *testing.T) {
	t.Parallel()
	_, reg := pkg.NewFastRegistryWithMocks(t)
	public, _ := testhelpers.NewKratosServer(t, reg)

	u := fmt.Sprintf("%s%s?id=not-a-uuid", public.URL, login.RouteDeleteTestFlow)
	res, body := doDelete(t, public.Client(), u)
	assert.Equal(t, http.StatusBadRequest, res.StatusCode, "%s", body)
}

func TestHandler_GetTestLoginFlow(t *testing.T) {
	t.Parallel()

	type setup func(t *testing.T, reg *driver.RegistryDefault) *login.Flow

	seedUncaptured := func(t *testing.T, reg *driver.RegistryDefault) *login.Flow {
		return newSeedTestFlow(t, context.Background(), reg, time.Hour)
	}
	seedCapturedMatching := func(t *testing.T, reg *driver.RegistryDefault) *login.Flow {
		return newSeedCapturedTestFlow(t, context.Background(), reg, time.Hour, nosurfx.FakeCSRFToken)
	}
	seedCapturedMismatch := func(t *testing.T, reg *driver.RegistryDefault) *login.Flow {
		return newSeedCapturedTestFlow(t, context.Background(), reg, time.Hour, "some-other-token")
	}

	for _, tc := range []struct {
		name       string
		setup      setup
		wantStatus int
		assertBody func(t *testing.T, body []byte)
	}{
		{
			name:       "uncaptured: 200",
			setup:      seedUncaptured,
			wantStatus: http.StatusOK,
			assertBody: func(t *testing.T, body []byte) {
				assert.Equal(t, "google", gjson.GetBytes(body, "test_context.provider_id").String())
			},
		},
		{
			name:       "captured with mismatched CSRF token: 403",
			setup:      seedCapturedMismatch,
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "captured with matching CSRF token: 200",
			setup:      seedCapturedMatching,
			wantStatus: http.StatusOK,
			assertBody: func(t *testing.T, body []byte) {
				assert.Equal(t, "alice@example.com",
					gjson.GetBytes(body, "test_context.debug_payload.id_token_claims.sub").String())
			},
		},
	} {
		t.Run("case="+tc.name, func(t *testing.T) {
			t.Parallel()
			reg := newRegistryWithOIDCProvider(t)
			public, _ := testhelpers.NewKratosServer(t, reg)

			f := tc.setup(t, reg)
			u := fmt.Sprintf("%s%s?id=%s", public.URL, login.RouteGetFlow, f.ID)

			req, err := http.NewRequest(http.MethodGet, u, nil)
			require.NoError(t, err)
			res, err := public.Client().Do(req)
			require.NoError(t, err)
			t.Cleanup(func() { _ = res.Body.Close() })
			body, err := io.ReadAll(res.Body)
			require.NoError(t, err)
			require.Equal(t, tc.wantStatus, res.StatusCode, "%s", body)
			if tc.assertBody != nil {
				tc.assertBody(t, body)
			}
		})
	}
}

// newSeedCapturedTestFlow creates a test login flow already advanced to the
// passed_challenge state with a DebugPayload populated, simulating what the
// OIDC callback would have persisted. csrfToken is written to f.CSRFToken so
// tests can simulate either the post-callback "matches the cookie" state or a
// mismatch.
func newSeedCapturedTestFlow(t *testing.T, ctx context.Context, reg *driver.RegistryDefault, exp time.Duration, csrfToken string) *login.Flow {
	t.Helper()
	f := newSeedTestFlow(t, ctx, reg, exp)
	f.CSRFToken = csrfToken
	f.State = flow.StatePassedChallenge
	require.NoError(t, f.SetTestContext(&login.TestContext{
		ProviderID: "google",
		DebugPayload: &login.DebugPayload{
			IDTokenClaims: map[string]any{"sub": "alice@example.com"},
		},
	}))
	require.NoError(t, reg.LoginFlowPersister().UpdateLoginFlow(ctx, f))
	return f
}
