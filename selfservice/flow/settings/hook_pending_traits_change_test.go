// Copyright © 2026 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package settings_test

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/driver"
	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/pkg"
	"github.com/ory/kratos/pkg/testhelpers"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/selfservice/flow/settings"
	"github.com/ory/kratos/selfservice/flow/verification"
	"github.com/ory/kratos/session"
	"github.com/ory/kratos/x"
	"github.com/ory/kratos/x/nosurfx"
)

// applyPTCFixture holds a wired-up registry plus a minimal set of persisted
// objects (identity, session, settings flow, verification flow, PTC) for
// unit-testing HookExecutor.ApplyPendingTraitsChange. Subtests mutate the
// PTC before calling ApplyPendingTraitsChange to exercise specific guard paths.
type applyPTCFixture struct {
	reg      *driver.RegistryDefault
	identity *identity.Identity
	sess     *session.Session
	flow     *settings.Flow
	ptc      *identity.PendingTraitsChange
}

// newApplyPTCFixture constructs a fresh, isolated fixture for each subtest.
// All objects are persisted to the SQLite test database.
func newApplyPTCFixture(t *testing.T) *applyPTCFixture {
	t.Helper()
	ctx := context.Background()
	conf, reg := pkg.NewFastRegistryWithMocks(t)
	testhelpers.SetDefaultIdentitySchema(conf, "file://./stub/identity.schema.json")

	// Create an identity whose traits will be the baseline for the PTC.
	i := &identity.Identity{
		ID:       x.NewUUID(),
		Traits:   identity.Traits(`{"email":"original@example.com"}`),
		SchemaID: "default",
		State:    identity.StateActive,
	}
	require.NoError(t, reg.IdentityManager().Create(ctx, i, identity.ManagerAllowWriteProtectedTraits))

	// Create an active session linked to the identity.
	sess, err := testhelpers.NewActiveSession(
		httptest.NewRequest(http.MethodGet, "/", nil),
		reg, i, time.Now().UTC(),
		identity.CredentialsTypePassword, identity.AuthenticatorAssuranceLevel1,
	)
	require.NoError(t, err)
	require.NoError(t, reg.SessionPersister().UpsertSession(ctx, sess))

	// Create a settings flow that acts as the origin for the PTC.
	f := &settings.Flow{
		ID:              x.NewUUID(),
		ExpiresAt:       time.Now().UTC().Add(time.Hour),
		IssuedAt:        time.Now().UTC(),
		RequestURL:      "http://localhost/settings",
		IdentityID:      i.ID,
		Identity:        i,
		Type:            flow.TypeBrowser,
		State:           flow.StateShowForm,
		InternalContext: []byte("{}"),
	}
	require.NoError(t, reg.SettingsFlowPersister().CreateSettingsFlow(ctx, f))

	// Create a verification flow — the PTC table has a FK on verification_flow_id
	// (ON DELETE CASCADE), so we must persist a real verification flow row.
	vf, err := verification.NewFlow(conf, time.Hour, nosurfx.FakeCSRFToken,
		httptest.NewRequest(http.MethodGet, "/verification", nil), nil, flow.TypeBrowser)
	require.NoError(t, err)
	vf.State = flow.StateEmailSent
	require.NoError(t, reg.VerificationFlowPersister().CreateVerificationFlow(ctx, vf))

	// Build a PTC that is valid in every field; subtests nil out individual
	// fields to trigger the specific guard under test.
	sessID := sess.ID
	originFlowID := f.ID
	ptc := &identity.PendingTraitsChange{
		ID:                   x.NewUUID(),
		IdentityID:           i.ID,
		NewAddressValue:      "new@example.com",
		NewAddressVia:        string(identity.AddressTypeEmail),
		OriginalTraitsHash:   identity.HashTraits(json.RawMessage(i.Traits)),
		ProposedTraits:       json.RawMessage(`{"email":"new@example.com"}`),
		VerificationFlowID:   vf.ID,
		Status:               identity.PendingTraitsChangeStatusPending,
		SessionID:            uuid.NullUUID{UUID: sessID, Valid: true},
		OriginSettingsFlowID: uuid.NullUUID{UUID: originFlowID, Valid: true},
	}
	require.NoError(t, reg.PendingTraitsChangePersister().CreatePendingTraitsChange(ctx, ptc))

	return &applyPTCFixture{
		reg:      reg,
		identity: i,
		sess:     sess,
		flow:     f,
		ptc:      ptc,
	}
}

func TestApplyPendingTraitsChange_NilOriginSettingsFlowID(t *testing.T) {
	t.Parallel()
	fx := newApplyPTCFixture(t)

	// Simulate a PTC whose OriginSettingsFlowID was never set (malformed writer)
	// or was cleared before the apply call.
	fx.ptc.OriginSettingsFlowID = uuid.NullUUID{}

	req := httptest.NewRequest(http.MethodGet, "/verify", nil)
	rw := httptest.NewRecorder()

	_, err := fx.reg.SettingsHookExecutor().ApplyPendingTraitsChange(req.Context(), rw, req, fx.ptc)
	require.Error(t, err)
	assert.ErrorIs(t, err, settings.ErrPendingTraitsChangeSessionInvalid)
}

// Ensure the fixture itself is valid: a fully-populated PTC must not be
// rejected. This guards against regressions in the fixture setup.
func TestApplyPendingTraitsChange_ValidPTC(t *testing.T) {
	t.Parallel()
	fx := newApplyPTCFixture(t)

	req := httptest.NewRequest(http.MethodGet, "/verify", nil)
	rw := httptest.NewRecorder()

	updated, err := fx.reg.SettingsHookExecutor().ApplyPendingTraitsChange(req.Context(), rw, req, fx.ptc)
	require.NoError(t, err)
	require.NotNil(t, updated)
}

func TestApplyPendingTraitsChange_NilSessionID(t *testing.T) {
	t.Parallel()
	fx := newApplyPTCFixture(t)

	fx.ptc.SessionID = uuid.NullUUID{}

	req := httptest.NewRequest(http.MethodGet, "/verify", nil)
	rw := httptest.NewRecorder()

	_, err := fx.reg.SettingsHookExecutor().ApplyPendingTraitsChange(req.Context(), rw, req, fx.ptc)
	require.Error(t, err)
	assert.ErrorIs(t, err, settings.ErrPendingTraitsChangeSessionInvalid)
}

func TestApplyPendingTraitsChange_NIDMismatch(t *testing.T) {
	t.Parallel()
	fx := newApplyPTCFixture(t)

	// Simulate a cross-tenant PTC: the PTC's NID differs from the session's NID.
	fx.ptc.NID = uuid.Must(uuid.NewV4())

	req := httptest.NewRequest(http.MethodGet, "/verify", nil)
	rw := httptest.NewRecorder()

	_, err := fx.reg.SettingsHookExecutor().ApplyPendingTraitsChange(req.Context(), rw, req, fx.ptc)
	require.Error(t, err)
	assert.ErrorIs(t, err, settings.ErrPendingTraitsChangeSessionInvalid)
}

// TestApplyPendingTraitsChange_WebhookReceivesOriginFlow asserts that the
// post-persist hook chain is invoked with the originating settings flow
// (not a synthetic one), so webhook payloads carry the real flow ID,
// type, and request URL.
func TestApplyPendingTraitsChange_WebhookReceivesOriginFlow(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	conf, reg := pkg.NewFastRegistryWithMocks(t)
	testhelpers.SetDefaultIdentitySchema(conf, "file://./stub/identity.schema.json")

	// Set up a recording webhook target that captures the raw body.
	receivedCh := make(chan []byte, 1)
	webhookServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		select {
		case receivedCh <- body:
		default:
		}
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(webhookServer.Close)

	// Register a web_hook as a settings-profile post-persist hook.
	// The jsonnet body template `function(ctx) ctx` forwards the entire context,
	// which includes the flow object with its ID, type, and request_url.
	conf.MustSet(ctx, config.HookStrategyKey(config.ViperKeySelfServiceSettingsAfter, "profile"), []map[string]any{
		{
			"hook": "web_hook",
			"config": map[string]any{
				"url":    webhookServer.URL,
				"method": "POST",
				"body":   "base64://" + base64.StdEncoding.EncodeToString([]byte(`function(ctx) ctx`)),
			},
		},
	})
	t.Cleanup(func() {
		conf.MustSet(ctx, config.HookStrategyKey(config.ViperKeySelfServiceSettingsAfter, "profile"), []map[string]any{})
	})

	i := &identity.Identity{
		ID:       x.NewUUID(),
		Traits:   identity.Traits(`{"email":"webhooktest@example.com"}`),
		SchemaID: "default",
		State:    identity.StateActive,
	}
	require.NoError(t, reg.IdentityManager().Create(ctx, i, identity.ManagerAllowWriteProtectedTraits))

	sess, err := testhelpers.NewActiveSession(
		httptest.NewRequest(http.MethodGet, "/", nil),
		reg, i, time.Now().UTC(),
		identity.CredentialsTypePassword, identity.AuthenticatorAssuranceLevel1,
	)
	require.NoError(t, err)
	require.NoError(t, reg.SessionPersister().UpsertSession(ctx, sess))

	originFlow := &settings.Flow{
		ID:              x.NewUUID(),
		ExpiresAt:       time.Now().UTC().Add(time.Hour),
		IssuedAt:        time.Now().UTC(),
		RequestURL:      "http://localhost/settings?origin=real",
		IdentityID:      i.ID,
		Identity:        i,
		Type:            flow.TypeAPI,
		State:           flow.StateShowForm,
		InternalContext: []byte("{}"),
	}
	require.NoError(t, reg.SettingsFlowPersister().CreateSettingsFlow(ctx, originFlow))

	vf, err := verification.NewFlow(conf, time.Hour, nosurfx.FakeCSRFToken,
		httptest.NewRequest(http.MethodGet, "/verification", nil), nil, flow.TypeBrowser)
	require.NoError(t, err)
	vf.State = flow.StateEmailSent
	require.NoError(t, reg.VerificationFlowPersister().CreateVerificationFlow(ctx, vf))

	sessID := sess.ID
	originFlowID := originFlow.ID
	ptc := &identity.PendingTraitsChange{
		ID:                   x.NewUUID(),
		IdentityID:           i.ID,
		NewAddressValue:      "webhooktest-new@example.com",
		NewAddressVia:        string(identity.AddressTypeEmail),
		OriginalTraitsHash:   identity.HashTraits(json.RawMessage(i.Traits)),
		ProposedTraits:       json.RawMessage(`{"email":"webhooktest-new@example.com"}`),
		VerificationFlowID:   vf.ID,
		Status:               identity.PendingTraitsChangeStatusPending,
		SessionID:            uuid.NullUUID{UUID: sessID, Valid: true},
		OriginSettingsFlowID: uuid.NullUUID{UUID: originFlowID, Valid: true},
	}
	require.NoError(t, reg.PendingTraitsChangePersister().CreatePendingTraitsChange(ctx, ptc))

	req := httptest.NewRequest(http.MethodGet, "/verify", nil)
	rw := httptest.NewRecorder()
	_, err = reg.SettingsHookExecutor().ApplyPendingTraitsChange(req.Context(), rw, req, ptc)
	require.NoError(t, err)

	var receivedBody []byte
	select {
	case receivedBody = <-receivedCh:
	case <-time.After(5 * time.Second):
		t.Fatal("settings post-persist webhook was not invoked after pending traits change applied")
	}

	// The webhook payload should carry the real origin settings flow's ID and
	// request URL — not values from a synthetic flow.
	assert.Contains(t, string(receivedBody), originFlow.ID.String(),
		"webhook payload should contain the real origin settings flow ID")
	assert.Contains(t, string(receivedBody), "http://localhost/settings?origin=real",
		"webhook payload should contain the real origin settings flow RequestURL")
}
