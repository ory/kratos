// Copyright © 2026 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package oidc_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/herodot"
	"github.com/ory/kratos/driver"
	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/pkg"
	"github.com/ory/kratos/pkg/testhelpers"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/selfservice/flow/login"
	"github.com/ory/kratos/selfservice/strategy/oidc"
	"github.com/ory/kratos/selfservice/strategy/oidc/oidcerr"
	"github.com/ory/kratos/ui/container"
	"github.com/ory/kratos/ui/node"
	"github.com/ory/kratos/x"
	"github.com/ory/kratos/x/nosurfx"
	"github.com/ory/x/configx"
	"github.com/ory/x/snapshotx"
	"github.com/ory/x/sqlxx"
)

func newTestRegistryForTestLogin(t *testing.T, providers ...map[string]any) (*config.Config, *driver.RegistryDefault) {
	t.Helper()
	return pkg.NewFastRegistryWithMocks(t,
		configx.WithValues(testhelpers.DefaultIdentitySchemaConfig("file://stub/registration.schema.json")),
		configx.WithValues(testhelpers.MethodEnableConfig(identity.CredentialsTypeOIDC, true)),
		configx.WithValue(config.ViperKeySelfServiceStrategyConfig+"."+string(identity.CredentialsTypeOIDC)+".config.providers", providers),
		configx.WithValue(config.ViperKeySelfServiceLoginUI, "http://admin.example.com/ui/login"),
	)
}

func newTestStrategyForTestLogin(t *testing.T, providers ...map[string]any) *oidc.Strategy {
	t.Helper()
	_, reg := newTestRegistryForTestLogin(t, providers...)
	s, err := reg.AllLoginStrategies().Strategy(identity.CredentialsTypeOIDC)
	require.NoError(t, err)
	return s.(*oidc.Strategy)
}

// fakeTestFlowProvider is a minimal Provider stub returning a fixed
// Configuration. Tests exercise processTestLogin directly, so no OAuth2 or
// Claims methods are exercised here.
type fakeTestFlowProvider struct {
	config *oidc.Configuration
}

func (p *fakeTestFlowProvider) Config() *oidc.Configuration { return p.config }

// newTestLoginFlow builds an admin-created test login flow with TestContext
// populated and persists it.
func newTestLoginFlow(t *testing.T, ctx context.Context, reg *driver.RegistryDefault, providerID string) *login.Flow {
	t.Helper()
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
	require.NoError(t, f.SetTestContext(&login.TestContext{ProviderID: providerID}))
	require.NoError(t, reg.LoginFlowPersister().CreateLoginFlow(ctx, f))
	return f
}

func providerNode(nodes node.Nodes) *node.Node {
	for i := range nodes {
		if nodes[i].ID() == "provider" {
			return nodes[i]
		}
	}
	return nil
}

func TestStrategy_PopulateTestLoginFlow_AddsProviderNode(t *testing.T) {
	t.Parallel()

	s := newTestStrategyForTestLogin(t, map[string]any{
		"provider":      "generic",
		"id":            "providerID",
		"label":         "My Provider",
		"client_id":     "invalid",
		"client_secret": "invalid",
		"issuer_url":    "https://foobar/",
		"mapper_url":    "file://./stub/oidc.facebook.jsonnet",
	})

	f := &login.Flow{
		ID:        x.NewUUID(),
		CSRFToken: "csrf-token",
		UI:        container.New(""),
	}
	r := httptest.NewRequest(http.MethodGet, "/self-service/login/browser", nil)

	require.NoError(t, s.PopulateTestLoginFlow(r, f, "providerID"))

	csrf := f.UI.GetNodes().Find("csrf_token")
	require.NotNil(t, csrf, "csrf_token node should be present")

	p := providerNode(f.UI.Nodes)
	require.NotNil(t, p, "provider submit node should be present")
	attrs, ok := p.Attributes.(*node.InputAttributes)
	require.True(t, ok)
	assert.Equal(t, "providerID", attrs.FieldValue)
	assert.Equal(t, node.InputAttributeTypeSubmit, attrs.Type)
	assert.Equal(t, node.OpenIDConnectGroup, p.Group)

	// Exactly one provider node was added.
	var providerCount int
	for i := range f.UI.Nodes {
		if f.UI.Nodes[i].ID() == "provider" {
			providerCount++
		}
	}
	assert.Equal(t, 1, providerCount)
}

func TestStrategy_PopulateTestLoginFlow_UISnapshot(t *testing.T) {
	t.Parallel()

	s := newTestStrategyForTestLogin(t, map[string]any{
		"provider":      "generic",
		"id":            "providerID",
		"label":         "My Provider",
		"client_id":     "invalid",
		"client_secret": "invalid",
		"issuer_url":    "https://foobar/",
		"mapper_url":    "file://./stub/oidc.facebook.jsonnet",
	})

	f := &login.Flow{
		ID:        x.NewUUID(),
		CSRFToken: nosurfx.FakeCSRFToken,
		UI:        container.New(""),
	}
	r := httptest.NewRequest(http.MethodGet, "/self-service/login/browser", nil)

	require.NoError(t, s.PopulateTestLoginFlow(r, f, "providerID"))

	// CSRF token value is masked per-request by nosurf and varies across
	// runs; the node's presence is asserted in another test.
	f.UI.Nodes.ResetNodes("csrf_token")
	snapshotx.SnapshotT(t, f.UI)
}

func TestStrategy_PopulateTestLoginFlow_UnknownProvider(t *testing.T) {
	t.Parallel()

	s := newTestStrategyForTestLogin(t, map[string]any{
		"provider":      "generic",
		"id":            "providerID",
		"client_id":     "invalid",
		"client_secret": "invalid",
		"issuer_url":    "https://foobar/",
		"mapper_url":    "file://./stub/oidc.facebook.jsonnet",
	})

	f := &login.Flow{
		ID:        x.NewUUID(),
		CSRFToken: "csrf-token",
		UI:        container.New(""),
	}
	r := httptest.NewRequest(http.MethodGet, "/self-service/login/browser", nil)

	err := s.PopulateTestLoginFlow(r, f, "nonesuch")
	require.Error(t, err)
	var herr *herodot.DefaultError
	require.ErrorAs(t, err, &herr)
	assert.Equal(t, http.StatusNotFound, herr.CodeField)
	assert.Contains(t, herr.ReasonField, "nonesuch")
}

func TestStrategy_ProcessTestLogin_HappyPath(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	_, reg := newTestRegistryForTestLogin(t, map[string]any{
		"provider":      "generic",
		"id":            "providerID",
		"client_id":     "invalid",
		"client_secret": "invalid",
		"issuer_url":    "https://foobar/",
		"mapper_url":    "file://./stub/oidc.hydra.jsonnet",
	})
	s, err := reg.AllLoginStrategies().Strategy(identity.CredentialsTypeOIDC)
	require.NoError(t, err)
	strat := s.(*oidc.Strategy)

	f := newTestLoginFlow(t, ctx, reg, "providerID")

	claims := &oidc.Claims{
		Subject: "alice@example.com",
		Email:   "alice@example.com",
		Website: "https://example.com",
		Picture: "https://example.com/alice.png",
		RawClaims: map[string]any{
			"groups": []any{"admin", "user"},
		},
	}
	provider := &fakeTestFlowProvider{config: &oidc.Configuration{
		ID:       "providerID",
		Provider: "generic",
		Mapper:   "file://./stub/oidc.hydra.jsonnet",
	}}

	// Baseline identity count — processTestLogin must not persist anything.
	before, err := reg.Persister().CountIdentities(ctx)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/self-service/methods/oidc/callback?code=x&state=y", nil)
	require.NoError(t, strat.ProcessTestLoginForTest(ctx, w, r, f, claims, provider))

	after, err := reg.Persister().CountIdentities(ctx)
	require.NoError(t, err)
	assert.Equal(t, before, after, "processTestLogin must never create an identity")

	got, err := reg.LoginFlowPersister().GetLoginFlow(ctx, f.ID)
	require.NoError(t, err)
	assert.Equal(t, flow.StatePassedChallenge, got.State)

	require.NoError(t, got.LoadTestContext())
	require.NotNil(t, got.TestContext)
	dp := got.TestContext.DebugPayload
	require.NotNil(t, dp)
	assert.Nil(t, dp.Error, "happy path should have no error: %+v", dp.Error)
	assert.Empty(t, dp.SchemaValidationErrors)
	assert.Equal(t, "file://./stub/oidc.hydra.jsonnet", dp.JsonnetMapperURL)
	assert.Equal(t, "alice@example.com", dp.IDTokenClaims["sub"])
	require.NotNil(t, dp.JsonnetInput)
	assert.Equal(t, "alice@example.com", dp.JsonnetInput["sub"])
	require.NotNil(t, dp.JsonnetOutput)
	ident, ok := dp.JsonnetOutput["identity"].(map[string]any)
	require.True(t, ok)
	traits, ok := ident["traits"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "alice@example.com", traits["subject"])

	// Browser redirected to the admin test UI with the flow ID.
	assert.Equal(t, http.StatusSeeOther, w.Code)
	loc := w.Header().Get("Location")
	assert.Contains(t, loc, "/session/test-oidc")
	assert.Contains(t, loc, "flow="+f.ID.String())
	// The results page is served from the host the callback arrived on, not
	// the login UI host (admin.example.com).
	assert.Contains(t, loc, "//example.com/session/test-oidc")
	assert.NotContains(t, loc, "admin.example.com")

	// CSRF token rotated on the callback and pinned onto the flow, so the
	// post-callback cookie value matches f.CSRFToken. Subsequent GET/DELETE
	// of the captured payload checks the CSRF cookie against this token.
	assert.Equal(t, nosurfx.FakeCSRFToken, got.CSRFToken,
		"finishTestLogin must regenerate the CSRF token onto the flow")
}

// TestStrategy_ProcessTestLogin_ReturnsToRequestHost asserts the test flow's
// results page is served from the public host the browser used to start the
// flow (the OIDC callback host), not from the configured login UI host. A
// project whose login UI lives on a different domain than its public API — for
// example a self-hosted Account Experience — must still land on the domain it
// launched the test from, otherwise the results page renders on a host that
// does not serve the /session/test-oidc route.
func TestStrategy_ProcessTestLogin_ReturnsToRequestHost(t *testing.T) {
	t.Parallel()
	ctx := t.Context()

	// The helper configures the login UI on admin.example.com. The callback
	// below arrives on a different public host, signin.example.com.
	_, reg := newTestRegistryForTestLogin(t, map[string]any{
		"provider":      "generic",
		"id":            "providerID",
		"client_id":     "invalid",
		"client_secret": "invalid",
		"issuer_url":    "https://foobar/",
		"mapper_url":    "file://./stub/oidc.hydra.jsonnet",
	})
	s, err := reg.AllLoginStrategies().Strategy(identity.CredentialsTypeOIDC)
	require.NoError(t, err)
	strat := s.(*oidc.Strategy)

	f := newTestLoginFlow(t, ctx, reg, "providerID")
	claims := &oidc.Claims{Subject: "alice@example.com", Email: "alice@example.com"}
	provider := &fakeTestFlowProvider{config: &oidc.Configuration{
		ID:       "providerID",
		Provider: "generic",
		Mapper:   "file://./stub/oidc.hydra.jsonnet",
	}}

	w := httptest.NewRecorder()
	// The OIDC provider redirects the browser back to the public host it used
	// to start the flow, which differs from the login UI host.
	r := httptest.NewRequest(http.MethodGet, "https://signin.example.com/self-service/methods/oidc/callback?code=x&state=y", nil)
	require.NoError(t, strat.ProcessTestLoginForTest(ctx, w, r, f, claims, provider))

	assert.Equal(t, http.StatusSeeOther, w.Code)
	loc := w.Header().Get("Location")
	assert.Contains(t, loc, "/session/test-oidc")
	assert.Contains(t, loc, "flow="+f.ID.String())
	assert.Contains(t, loc, "signin.example.com",
		"results page must be served from the host the flow started on")
	assert.NotContains(t, loc, "admin.example.com",
		"results page must not jump to the login UI host")
}

// TestStrategy_ProcessTestLogin_PrefersCapturedBaseURL covers the dominant Ory
// Network path: a gateway-validated middleware captures the customer-facing
// base URL onto the request context. That captured value must win over both
// the raw request host and the login UI host, so the results page is always
// served from the trusted customer-facing domain rather than anything an
// untrusted X-Forwarded-Host could inject.
func TestStrategy_ProcessTestLogin_PrefersCapturedBaseURL(t *testing.T) {
	t.Parallel()
	ctx := t.Context()

	_, reg := newTestRegistryForTestLogin(t, map[string]any{
		"provider":      "generic",
		"id":            "providerID",
		"client_id":     "invalid",
		"client_secret": "invalid",
		"issuer_url":    "https://foobar/",
		"mapper_url":    "file://./stub/oidc.hydra.jsonnet",
	})
	s, err := reg.AllLoginStrategies().Strategy(identity.CredentialsTypeOIDC)
	require.NoError(t, err)
	strat := s.(*oidc.Strategy)

	f := newTestLoginFlow(t, ctx, reg, "providerID")
	claims := &oidc.Claims{Subject: "alice@example.com", Email: "alice@example.com"}
	provider := &fakeTestFlowProvider{config: &oidc.Configuration{
		ID:       "providerID",
		Provider: "generic",
		Mapper:   "file://./stub/oidc.hydra.jsonnet",
	}}

	// The gateway captured the customer-facing base URL onto the context. The
	// request itself arrives on an unrelated internal host with a spoofed
	// X-Forwarded-Host that must be ignored in favor of the captured value.
	captured, perr := url.Parse("https://signin.example.com")
	require.NoError(t, perr)
	reqCtx := x.WithBaseURL(ctx, captured)
	r := httptest.NewRequest(http.MethodGet, "https://internal-host.local/self-service/methods/oidc/callback?code=x&state=y", nil)
	r = r.WithContext(reqCtx)
	r.Header.Set("X-Forwarded-Host", "evil.example.com")

	w := httptest.NewRecorder()
	require.NoError(t, strat.ProcessTestLoginForTest(reqCtx, w, r, f, claims, provider))

	assert.Equal(t, http.StatusSeeOther, w.Code)
	loc := w.Header().Get("Location")
	assert.Contains(t, loc, "//signin.example.com/session/test-oidc",
		"results page must use the gateway-captured base URL")
	assert.NotContains(t, loc, "evil.example.com",
		"a spoofed X-Forwarded-Host must not override the captured base URL")
	assert.NotContains(t, loc, "internal-host.local")
	assert.NotContains(t, loc, "admin.example.com")
}

func TestStrategy_ProcessTestLogin_AlreadyCaptured(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	_, reg := newTestRegistryForTestLogin(t, map[string]any{
		"provider":      "generic",
		"id":            "providerID",
		"client_id":     "invalid",
		"client_secret": "invalid",
		"issuer_url":    "https://foobar/",
		"mapper_url":    "file://./stub/oidc.hydra.jsonnet",
	})
	s, err := reg.AllLoginStrategies().Strategy(identity.CredentialsTypeOIDC)
	require.NoError(t, err)
	strat := s.(*oidc.Strategy)

	f := newTestLoginFlow(t, ctx, reg, "providerID")
	f.State = flow.StatePassedChallenge
	require.NoError(t, reg.LoginFlowPersister().UpdateLoginFlow(ctx, f))

	provider := &fakeTestFlowProvider{config: &oidc.Configuration{
		ID:       "providerID",
		Provider: "generic",
		Mapper:   "file://./stub/oidc.hydra.jsonnet",
	}}

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/self-service/methods/oidc/callback?code=x&state=y", nil)
	err = strat.ProcessTestLoginForTest(ctx, w, r, f, &oidc.Claims{Subject: "alice"}, provider)
	require.Error(t, err)
	var herr *herodot.DefaultError
	require.ErrorAs(t, err, &herr)
	assert.Equal(t, http.StatusConflict, herr.CodeField)
}

func TestStrategy_ProcessTestLogin_RawTokensNotStored(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	_, reg := newTestRegistryForTestLogin(t, map[string]any{
		"provider":      "generic",
		"id":            "providerID",
		"client_id":     "invalid",
		"client_secret": "invalid",
		"issuer_url":    "https://foobar/",
		"mapper_url":    "file://./stub/oidc.hydra.jsonnet",
	})
	s, err := reg.AllLoginStrategies().Strategy(identity.CredentialsTypeOIDC)
	require.NoError(t, err)
	strat := s.(*oidc.Strategy)

	f := newTestLoginFlow(t, ctx, reg, "providerID")

	claims := &oidc.Claims{
		Subject: "alice@example.com",
		RawClaims: map[string]any{
			"groups": []any{"admin"},
		},
	}
	provider := &fakeTestFlowProvider{config: &oidc.Configuration{
		ID:       "providerID",
		Provider: "generic",
		Mapper:   "file://./stub/oidc.hydra.jsonnet",
	}}

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/self-service/methods/oidc/callback?code=x&state=y", nil)
	require.NoError(t, strat.ProcessTestLoginForTest(ctx, w, r, f, claims, provider))

	got, err := reg.LoginFlowPersister().GetLoginFlow(ctx, f.ID)
	require.NoError(t, err)
	require.NoError(t, got.LoadTestContext())
	require.NotNil(t, got.TestContext)
	require.NotNil(t, got.TestContext.DebugPayload)

	b, err := json.Marshal(got.TestContext.DebugPayload)
	require.NoError(t, err)
	payload := string(b)
	assert.NotContains(t, payload, "raw_id_token")
	assert.NotContains(t, payload, "access_token")
	assert.NotContains(t, payload, "refresh_token")
}

func TestStrategy_ProcessTestLogin_NilClaims(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	_, reg := newTestRegistryForTestLogin(t, map[string]any{
		"provider":      "generic",
		"id":            "providerID",
		"client_id":     "invalid",
		"client_secret": "invalid",
		"issuer_url":    "https://foobar/",
		"mapper_url":    "file://./stub/oidc.hydra.jsonnet",
	})
	s, err := reg.AllLoginStrategies().Strategy(identity.CredentialsTypeOIDC)
	require.NoError(t, err)
	strat := s.(*oidc.Strategy)

	f := newTestLoginFlow(t, ctx, reg, "providerID")
	provider := &fakeTestFlowProvider{config: &oidc.Configuration{
		ID:       "providerID",
		Provider: "generic",
		Mapper:   "file://./stub/oidc.hydra.jsonnet",
	}}

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/self-service/methods/oidc/callback?code=x&state=y", nil)
	require.NoError(t, strat.ProcessTestLoginForTest(ctx, w, r, f, nil, provider))

	got, err := reg.LoginFlowPersister().GetLoginFlow(ctx, f.ID)
	require.NoError(t, err)
	require.NoError(t, got.LoadTestContext())
	require.NotNil(t, got.TestContext)
	dp := got.TestContext.DebugPayload
	require.NotNil(t, dp)
	require.NotNil(t, dp.Error)
	assert.Equal(t, "claims_decode", dp.Error.Step)
	assert.Equal(t, "missing_claims", dp.Error.Reason)
}

func TestStrategy_ProcessTestLogin_MapperEvaluationFailure(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	_, reg := newTestRegistryForTestLogin(t, map[string]any{
		"provider":      "generic",
		"id":            "providerID",
		"client_id":     "invalid",
		"client_secret": "invalid",
		"issuer_url":    "https://foobar/",
		"mapper_url":    "file://./stub/oidc.hydra.jsonnet",
	})
	s, err := reg.AllLoginStrategies().Strategy(identity.CredentialsTypeOIDC)
	require.NoError(t, err)
	strat := s.(*oidc.Strategy)

	f := newTestLoginFlow(t, ctx, reg, "providerID")
	// Point the provider at a mapper file that does not exist so
	// EvaluateClaimsMapper fails on fetch.
	provider := &fakeTestFlowProvider{config: &oidc.Configuration{
		ID:       "providerID",
		Provider: "generic",
		Mapper:   "file://./stub/oidc.does-not-exist.jsonnet",
	}}

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/self-service/methods/oidc/callback?code=x&state=y", nil)
	require.NoError(t, strat.ProcessTestLoginForTest(ctx, w, r, f, &oidc.Claims{Subject: "alice"}, provider))

	got, err := reg.LoginFlowPersister().GetLoginFlow(ctx, f.ID)
	require.NoError(t, err)
	require.NoError(t, got.LoadTestContext())
	require.NotNil(t, got.TestContext)
	dp := got.TestContext.DebugPayload
	require.NotNil(t, dp)
	require.NotNil(t, dp.Error)
	assert.Equal(t, "jsonnet_run", dp.Error.Step)
	assert.Equal(t, "evaluation_failed", dp.Error.Reason)
}

func TestStrategy_FinishTestLogin_Idempotent(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	_, reg := newTestRegistryForTestLogin(t, map[string]any{
		"provider":      "generic",
		"id":            "providerID",
		"client_id":     "invalid",
		"client_secret": "invalid",
		"issuer_url":    "https://foobar/",
		"mapper_url":    "file://./stub/oidc.hydra.jsonnet",
	})
	s, err := reg.AllLoginStrategies().Strategy(identity.CredentialsTypeOIDC)
	require.NoError(t, err)
	strat := s.(*oidc.Strategy)

	// Seed a captured test flow directly so a subsequent finishTestLogin
	// call observes the idempotency guard.
	f := newTestLoginFlow(t, ctx, reg, "providerID")
	originalDP := &login.DebugPayload{
		IDTokenClaims: map[string]any{"sub": "first"},
	}
	require.NoError(t, f.SetTestContext(&login.TestContext{
		ProviderID:   "providerID",
		DebugPayload: originalDP,
	}))
	f.State = flow.StatePassedChallenge
	require.NoError(t, reg.LoginFlowPersister().UpdateLoginFlow(ctx, f))

	// Call finishTestLogin with a different DebugPayload. The guard must
	// refuse to overwrite the captured payload.
	replacementDP := &login.DebugPayload{
		IDTokenClaims: map[string]any{"sub": "second"},
		Error:         &login.StepError{Step: "callback", Reason: "callback_failed", Message: "should not win"},
	}
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/self-service/methods/oidc/callback?code=x&state=y", nil)
	require.NoError(t, strat.FinishTestLoginForTest(ctx, w, r, f, replacementDP))

	// Browser is still redirected to the admin test UI.
	assert.Equal(t, http.StatusSeeOther, w.Code)

	got, err := reg.LoginFlowPersister().GetLoginFlow(ctx, f.ID)
	require.NoError(t, err)
	require.NoError(t, got.LoadTestContext())
	require.NotNil(t, got.TestContext)
	require.NotNil(t, got.TestContext.DebugPayload)
	assert.Equal(t, "first", got.TestContext.DebugPayload.IDTokenClaims["sub"])
	assert.Nil(t, got.TestContext.DebugPayload.Error)
	assert.Equal(t, flow.StatePassedChallenge, got.State)
}

func TestStrategy_ForwardError_TestLoginFlow(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	_, reg := newTestRegistryForTestLogin(t, map[string]any{
		"provider":      "generic",
		"id":            "providerID",
		"client_id":     "invalid",
		"client_secret": "invalid",
		"issuer_url":    "https://foobar/",
		"mapper_url":    "file://./stub/oidc.hydra.jsonnet",
	})
	s, err := reg.AllLoginStrategies().Strategy(identity.CredentialsTypeOIDC)
	require.NoError(t, err)
	strat := s.(*oidc.Strategy)

	f := newTestLoginFlow(t, ctx, reg, "providerID")

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/self-service/methods/oidc/callback?code=x&state=y", nil)

	// Feed a step-tagged error to exercise the classifier and the
	// test-login branch of forwardError.
	strat.ForwardErrorForTest(ctx, w, r, f, oidcerr.Wrap(oidcerr.StepProviderDenied, errors.New("access_denied: user refused")))

	// The browser must be redirected to the admin test UI, not the
	// standard login error handler.
	assert.Equal(t, http.StatusSeeOther, w.Code)
	loc := w.Header().Get("Location")
	assert.Contains(t, loc, "/session/test-oidc")
	assert.Contains(t, loc, "flow="+f.ID.String())
	// The results page is served from the host the callback arrived on, not
	// the login UI host (admin.example.com).
	assert.Contains(t, loc, "//example.com/session/test-oidc")
	assert.NotContains(t, loc, "admin.example.com")

	got, err := reg.LoginFlowPersister().GetLoginFlow(ctx, f.ID)
	require.NoError(t, err)
	assert.Equal(t, flow.StatePassedChallenge, got.State)

	require.NoError(t, got.LoadTestContext())
	require.NotNil(t, got.TestContext)
	dp := got.TestContext.DebugPayload
	require.NotNil(t, dp)
	require.NotNil(t, dp.Error)
	assert.Equal(t, "provider_denied", dp.Error.Step)
	assert.Equal(t, "access_denied", dp.Error.Reason)
}
