// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package hook_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/driver"
	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/pkg"
	"github.com/ory/kratos/pkg/testhelpers"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/selfservice/flow/settings"
	"github.com/ory/kratos/selfservice/hook"
	"github.com/ory/kratos/session"
	"github.com/ory/kratos/ui/container"
	"github.com/ory/kratos/x"
	"github.com/ory/x/sqlxx"
	"github.com/ory/x/urlx"
)

func TestVerifyNewAddress(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	setup := func(t *testing.T) (*config.Config, *hook.VerifyNewAddress, *driver.RegistryDefault) {
		conf, reg := pkg.NewFastRegistryWithMocks(t)
		testhelpers.SetDefaultIdentitySchema(conf, "file://./stub/verify_single_email.schema.json")
		conf.MustSet(ctx, config.ViperKeyPublicBaseURL, "https://www.ory.sh/")
		conf.MustSet(ctx, config.ViperKeyCourierSMTPURL, "smtp://foo@bar@dev.null/")
		h := hook.NewVerifyNewAddress(reg)
		return conf, h, reg
	}

	newSettingsFlow := func(t *testing.T, i identity.Identity) *settings.Flow {
		t.Helper()
		now := time.Now().UTC()
		return &settings.Flow{
			ID:              x.NewUUID(),
			ExpiresAt:       now.Add(time.Hour),
			IssuedAt:        now,
			RequestURL:      "http://foo.com/settings",
			IdentityID:      i.ID,
			Identity:        &i,
			Type:            flow.TypeBrowser,
			State:           flow.StateShowForm,
			UI:              &container.Container{Method: "POST"},
			InternalContext: []byte("{}"),
		}
	}

	verifyAllAddresses := func(t *testing.T, ctx context.Context, reg *driver.RegistryDefault, i *identity.Identity) *identity.Identity {
		t.Helper()
		for idx := range i.VerifiableAddresses {
			i.VerifiableAddresses[idx].Verified = true
			verifiedAt := sqlxx.NullTime(time.Now())
			i.VerifiableAddresses[idx].VerifiedAt = &verifiedAt
			i.VerifiableAddresses[idx].Status = identity.VerifiableAddressStatusCompleted
			require.NoError(t, reg.PrivilegedIdentityPool().UpdateVerifiableAddress(ctx, &i.VerifiableAddresses[idx]))
		}
		var err error
		i, err = reg.PrivilegedIdentityPool().GetIdentity(ctx, i.ID, identity.ExpandDefault)
		require.NoError(t, err)
		return i
	}

	createIdentity := func(t *testing.T, ctx context.Context, reg *driver.RegistryDefault, email string, verified bool) *identity.Identity {
		t.Helper()
		i := identity.NewIdentity(config.DefaultIdentityTraitsSchemaID)
		i.Traits = identity.Traits(`{"email":"` + email + `","name":"Test"}`)
		require.NoError(t, reg.IdentityManager().Create(ctx, i))

		if verified {
			i = verifyAllAddresses(t, ctx, reg, i)
		}
		return i
	}

	t.Run("case=no-op when address does not change", func(t *testing.T) {
		t.Parallel()
		_, h, reg := setup(t)

		original := createIdentity(t, ctx, reg, "old@example.com", true)
		f := newSettingsFlow(t, *original)
		require.NoError(t, reg.SettingsFlowPersister().CreateSettingsFlow(ctx, f))

		proposed := &identity.Identity{
			ID:     original.ID,
			Traits: identity.Traits(`{"email":"old@example.com","name":"Test"}`),
			VerifiableAddresses: []identity.VerifiableAddress{
				{Value: "old@example.com", Via: identity.AddressTypeEmail, IdentityID: original.ID},
			},
		}

		sess := &session.Session{ID: x.NewUUID(), Identity: original, AuthenticatedAt: time.Now()}
		r := &http.Request{URL: urlx.ParseOrPanic("https://www.ory.sh/")}
		r = r.WithContext(ctx)
		w := httptest.NewRecorder()

		err := h.ExecuteSettingsPrePersistHook(w, r, f, proposed, sess)
		require.NoError(t, err, "hook should be a no-op when address did not change")
	})

	t.Run("case=aborts flow when address changes", func(t *testing.T) {
		t.Parallel()
		_, h, reg := setup(t)

		original := createIdentity(t, ctx, reg, "old@example.com", true)
		f := newSettingsFlow(t, *original)
		require.NoError(t, reg.SettingsFlowPersister().CreateSettingsFlow(ctx, f))

		proposed := &identity.Identity{
			ID:     original.ID,
			Traits: identity.Traits(`{"email":"new@example.com","name":"Test"}`),
			VerifiableAddresses: []identity.VerifiableAddress{
				{Value: "new@example.com", Via: identity.AddressTypeEmail, IdentityID: original.ID},
			},
		}

		sess := &session.Session{ID: x.NewUUID(), Identity: original, AuthenticatedAt: time.Now()}
		r := &http.Request{URL: urlx.ParseOrPanic("https://www.ory.sh/")}
		r = r.WithContext(ctx)
		w := httptest.NewRecorder()

		err := h.ExecuteSettingsPrePersistHook(w, r, f, proposed, sess)
		require.Error(t, err)
		require.True(t, errors.Is(err, settings.ErrHookAbortFlow), "expected ErrHookAbortFlow, got: %v", err)

		// Settings flow should have a ContinueWith entry pointing to the verification flow.
		require.NotEmpty(t, f.ContinueWith(), "settings flow should have ContinueWith items")
		cw := f.ContinueWith()[0]
		assert.IsType(t, &flow.ContinueWithVerificationUI{}, cw)
		vfID := cw.(*flow.ContinueWithVerificationUI).Flow.ID

		// A PendingTraitsChange record should exist.
		ptc, err := reg.PendingTraitsChangePersister().GetPendingTraitsChangeByVerificationFlow(ctx, vfID)
		require.NoError(t, err)
		assert.Equal(t, identity.PendingTraitsChangeStatusPending, ptc.Status)
		assert.Equal(t, "new@example.com", ptc.NewAddressValue)
		assert.Equal(t, original.ID, ptc.IdentityID)
		assert.Equal(t, identity.HashTraits(json.RawMessage(original.Traits)), ptc.OriginalTraitsHash)

		// Verification flow should exist and be in StateEmailSent.
		vf, err := reg.VerificationFlowPersister().GetVerificationFlow(ctx, vfID)
		require.NoError(t, err)
		assert.Equal(t, flow.StateEmailSent, vf.State)

		// A verification code message should have been sent.
		messages, err := reg.CourierPersister().NextMessages(ctx, 12)
		require.NoError(t, err)
		require.Len(t, messages, 1)
		assert.Equal(t, "new@example.com", messages[0].Recipient)
	})

	t.Run("case=previous pending changes are deleted on new submission", func(t *testing.T) {
		t.Parallel()
		_, h, reg := setup(t)

		original := createIdentity(t, ctx, reg, "old@example.com", true)

		// First change: old@example.com -> new1@example.com
		f1 := newSettingsFlow(t, *original)
		require.NoError(t, reg.SettingsFlowPersister().CreateSettingsFlow(ctx, f1))

		proposed1 := &identity.Identity{
			ID:     original.ID,
			Traits: identity.Traits(`{"email":"new1@example.com","name":"Test"}`),
			VerifiableAddresses: []identity.VerifiableAddress{
				{Value: "new1@example.com", Via: identity.AddressTypeEmail, IdentityID: original.ID},
			},
		}

		sess := &session.Session{ID: x.NewUUID(), Identity: original, AuthenticatedAt: time.Now()}
		r := &http.Request{URL: urlx.ParseOrPanic("https://www.ory.sh/")}
		r = r.WithContext(ctx)
		w := httptest.NewRecorder()

		err := h.ExecuteSettingsPrePersistHook(w, r, f1, proposed1, sess)
		require.True(t, errors.Is(err, settings.ErrHookAbortFlow))

		vfID1 := f1.ContinueWith()[0].(*flow.ContinueWithVerificationUI).Flow.ID
		ptc1, err := reg.PendingTraitsChangePersister().GetPendingTraitsChangeByVerificationFlow(ctx, vfID1)
		require.NoError(t, err)
		assert.Equal(t, "new1@example.com", ptc1.NewAddressValue)

		// Drain courier messages from first change.
		_, _ = reg.CourierPersister().NextMessages(ctx, 12)

		// Second change: old@example.com -> new2@example.com
		f2 := newSettingsFlow(t, *original)
		require.NoError(t, reg.SettingsFlowPersister().CreateSettingsFlow(ctx, f2))

		proposed2 := &identity.Identity{
			ID:     original.ID,
			Traits: identity.Traits(`{"email":"new2@example.com","name":"Test"}`),
			VerifiableAddresses: []identity.VerifiableAddress{
				{Value: "new2@example.com", Via: identity.AddressTypeEmail, IdentityID: original.ID},
			},
		}

		w2 := httptest.NewRecorder()
		err = h.ExecuteSettingsPrePersistHook(w2, r, f2, proposed2, sess)
		require.True(t, errors.Is(err, settings.ErrHookAbortFlow))

		vfID2 := f2.ContinueWith()[0].(*flow.ContinueWithVerificationUI).Flow.ID
		ptc2, err := reg.PendingTraitsChangePersister().GetPendingTraitsChangeByVerificationFlow(ctx, vfID2)
		require.NoError(t, err)
		assert.Equal(t, "new2@example.com", ptc2.NewAddressValue)

		// The first pending change should have been deleted.
		_, err = reg.PendingTraitsChangePersister().GetPendingTraitsChangeByVerificationFlow(ctx, vfID1)
		require.Error(t, err, "first pending change should have been deleted")
	})

	t.Run("case=correctly detects change with two email addresses", func(t *testing.T) {
		t.Parallel()
		conf, h, reg := setup(t)
		testhelpers.SetDefaultIdentitySchema(conf, "file://./stub/verify_two_emails.schema.json")

		i := identity.NewIdentity(config.DefaultIdentityTraitsSchemaID)
		i.Traits = identity.Traits(`{"email":"primary@example.com","recovery_email":"recovery@example.com","name":"Test"}`)
		require.NoError(t, reg.IdentityManager().Create(ctx, i))

		// Verify both addresses.
		i = verifyAllAddresses(t, ctx, reg, i)
		require.Len(t, i.VerifiableAddresses, 2, "identity should have two verifiable addresses")

		// Change only the primary email; recovery email stays the same.
		f := newSettingsFlow(t, *i)
		require.NoError(t, reg.SettingsFlowPersister().CreateSettingsFlow(ctx, f))

		proposed := &identity.Identity{
			ID:     i.ID,
			Traits: identity.Traits(`{"email":"new-primary@example.com","recovery_email":"recovery@example.com","name":"Test"}`),
			VerifiableAddresses: []identity.VerifiableAddress{
				{Value: "new-primary@example.com", Via: identity.AddressTypeEmail, IdentityID: i.ID},
				{Value: "recovery@example.com", Via: identity.AddressTypeEmail, IdentityID: i.ID},
			},
		}

		sess := &session.Session{ID: x.NewUUID(), Identity: i, AuthenticatedAt: time.Now()}
		r := &http.Request{URL: urlx.ParseOrPanic("https://www.ory.sh/")}
		r = r.WithContext(ctx)
		w := httptest.NewRecorder()

		err := h.ExecuteSettingsPrePersistHook(w, r, f, proposed, sess)
		// Should abort with exactly one changed address (not two, not zero).
		require.Error(t, err)
		require.True(t, errors.Is(err, settings.ErrHookAbortFlow), "expected ErrHookAbortFlow, got: %v", err)

		// Verify only one verification was triggered (for the changed address).
		assert.Len(t, f.ContinueWith(), 1)
		cw := f.ContinueWith()[0]
		assert.IsType(t, &flow.ContinueWithVerificationUI{}, cw)
		vfID := cw.(*flow.ContinueWithVerificationUI).Flow.ID

		ptc, err := reg.PendingTraitsChangePersister().GetPendingTraitsChangeByVerificationFlow(ctx, vfID)
		require.NoError(t, err)
		assert.Equal(t, "new-primary@example.com", ptc.NewAddressValue, "pending change should be for the changed address only")
	})

	t.Run("case=no-op with two unchanged email addresses", func(t *testing.T) {
		t.Parallel()
		conf, h, reg := setup(t)
		testhelpers.SetDefaultIdentitySchema(conf, "file://./stub/verify_two_emails.schema.json")

		i := identity.NewIdentity(config.DefaultIdentityTraitsSchemaID)
		i.Traits = identity.Traits(`{"email":"primary@example.com","recovery_email":"recovery@example.com","name":"Test"}`)
		require.NoError(t, reg.IdentityManager().Create(ctx, i))

		i = verifyAllAddresses(t, ctx, reg, i)

		f := newSettingsFlow(t, *i)
		require.NoError(t, reg.SettingsFlowPersister().CreateSettingsFlow(ctx, f))

		proposed := &identity.Identity{
			ID:     i.ID,
			Traits: identity.Traits(`{"email":"primary@example.com","recovery_email":"recovery@example.com","name":"Test"}`),
			VerifiableAddresses: []identity.VerifiableAddress{
				{Value: "primary@example.com", Via: identity.AddressTypeEmail, IdentityID: i.ID},
				{Value: "recovery@example.com", Via: identity.AddressTypeEmail, IdentityID: i.ID},
			},
		}

		sess := &session.Session{ID: x.NewUUID(), Identity: i, AuthenticatedAt: time.Now()}
		r := &http.Request{URL: urlx.ParseOrPanic("https://www.ory.sh/")}
		r = r.WithContext(ctx)
		w := httptest.NewRecorder()

		err := h.ExecuteSettingsPrePersistHook(w, r, f, proposed, sess)
		require.NoError(t, err, "hook should be a no-op when no addresses changed")
		assert.Empty(t, f.ContinueWith())
	})

	t.Run("case=returns error when multiple addresses change at once", func(t *testing.T) {
		t.Parallel()
		conf, h, reg := setup(t)
		testhelpers.SetDefaultIdentitySchema(conf, "file://./stub/verify_email_and_phone.schema.json")

		i := identity.NewIdentity(config.DefaultIdentityTraitsSchemaID)
		i.Traits = identity.Traits(`{"email":"old@example.com","phone":"+12345678901","name":"Test"}`)
		require.NoError(t, reg.IdentityManager().Create(ctx, i))

		i = verifyAllAddresses(t, ctx, reg, i)

		f := newSettingsFlow(t, *i)
		require.NoError(t, reg.SettingsFlowPersister().CreateSettingsFlow(ctx, f))

		proposed := &identity.Identity{
			ID:     i.ID,
			Traits: identity.Traits(`{"email":"new@example.com","phone":"+19876543210","name":"Test"}`),
			VerifiableAddresses: []identity.VerifiableAddress{
				{Value: "new@example.com", Via: identity.AddressTypeEmail, IdentityID: i.ID},
				{Value: "+19876543210", Via: "sms", IdentityID: i.ID},
			},
		}

		sess := &session.Session{ID: x.NewUUID(), Identity: i, AuthenticatedAt: time.Now()}
		r := &http.Request{URL: urlx.ParseOrPanic("https://www.ory.sh/")}
		r = r.WithContext(ctx)
		w := httptest.NewRecorder()

		err := h.ExecuteSettingsPrePersistHook(w, r, f, proposed, sess)
		require.Error(t, err)
		require.True(t, errors.Is(err, settings.ErrHookAbortFlow), "expected ErrHookAbortFlow, got: %v", err)

		// The flow should contain the too-many-address-changes error message.
		require.NotEmpty(t, f.UI.Messages)
	})
}
