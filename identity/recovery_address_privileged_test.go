// Copyright © 2025 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package identity_test

import (
	"fmt"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/pkg"
	"github.com/ory/kratos/pkg/testhelpers"
	"github.com/ory/kratos/x"
	"github.com/ory/x/configx"
)

// TestRecoveryAddressChangeRequiresPrivilegedSession reproduces HackerOne
// #3786482: the Settings-flow privileged-session gate compares only credentials
// and verifiable addresses, never recovery addresses. A non-privileged (stale)
// session could therefore change a victim's recovery address without re-auth,
// then drive account recovery to take over the account.
//
// The fix adds RecoveryAddressesEqual to identity.Manager.requiresPrivilegedAccess
// so that changing only a recovery address on a non-privileged session is
// rejected with identity.ErrProtectedFieldModified — the same protected-field
// path that already guards credentials and verifiable addresses.
//
// The "default" schema here maps email_recovery to a recovery address ONLY (no
// verification extension), so changing email_recovery changes the recovery
// address without touching the verifiable address. This isolates the recovery
// address as the single changed protected field.
func TestRecoveryAddressChangeRequiresPrivilegedSession(t *testing.T) {
	t.Parallel()

	_, reg := pkg.NewFastRegistryWithMocks(t,
		configx.WithValues(map[string]any{
			config.ViperKeyDefaultIdentitySchemaID: "default",
		}),
		configx.WithValues(testhelpers.IdentitySchemasConfig(map[string]string{
			"default": "file://./stub/manager.schema.json",
		})),
	)

	// traits builds a trait set where the recovery address (email_recovery) can
	// be varied independently of the verifiable address (email_verify) and the
	// login identifiers (email, email_creds).
	traits := func(identifier, recovery string) identity.Traits {
		return identity.Traits(fmt.Sprintf(
			`{"email":%[1]q,"email_creds":%[1]q,"email_verify":%[1]q,"email_recovery":%[2]q}`,
			identifier, recovery,
		))
	}

	recoveryAddressValues := func(t *testing.T, i *identity.Identity) []string {
		t.Helper()
		fromStore, err := reg.PrivilegedIdentityPool().GetIdentityConfidential(t.Context(), i.ID)
		require.NoError(t, err)
		values := make([]string, 0, len(fromStore.RecoveryAddresses))
		for _, a := range fromStore.RecoveryAddresses {
			values = append(values, a.Value)
		}
		return values
	}

	t.Run("case=non-privileged session must not change a recovery address", func(t *testing.T) {
		identifier := "victim-" + x.NewUUID().String() + "@ory.sh"
		original := identity.NewIdentity(config.DefaultIdentityTraitsSchemaID)
		original.Traits = traits(identifier, "original-recovery@ory.sh")
		require.NoError(t, reg.IdentityManager().Create(t.Context(), original))
		require.Equal(t, []string{"original-recovery@ory.sh"}, recoveryAddressValues(t, original))

		// Attacker (non-privileged session) changes ONLY the recovery address.
		// The login identifier and verifiable address are unchanged.
		original.Traits = traits(identifier, "attacker-recovery@ory.sh")
		err := reg.IdentityManager().Update(t.Context(), original)
		require.Error(t, err, "changing a recovery address without a privileged session must be rejected")
		assert.Equal(t, identity.ErrProtectedFieldModified(), errors.Cause(err))

		// The stored recovery address must be unchanged.
		assert.Equal(t, []string{"original-recovery@ory.sh"}, recoveryAddressValues(t, original),
			"the recovery address must not have been modified by a non-privileged session")
	})

	t.Run("case=privileged session may change a recovery address", func(t *testing.T) {
		identifier := "owner-" + x.NewUUID().String() + "@ory.sh"
		original := identity.NewIdentity(config.DefaultIdentityTraitsSchemaID)
		original.Traits = traits(identifier, "owner-recovery@ory.sh")
		require.NoError(t, reg.IdentityManager().Create(t.Context(), original))

		original.Traits = traits(identifier, "new-owner-recovery@ory.sh")
		require.NoError(t, reg.IdentityManager().Update(t.Context(), original, identity.ManagerAllowWriteProtectedTraits))

		assert.Equal(t, []string{"new-owner-recovery@ory.sh"}, recoveryAddressValues(t, original))
	})

	t.Run("case=non-privileged session may still update unprotected traits", func(t *testing.T) {
		// Guard against the fix over-blocking: an update that does not touch the
		// recovery address (or any protected field) must continue to succeed on a
		// non-privileged session.
		identifier := "stable-" + x.NewUUID().String() + "@ory.sh"
		original := identity.NewIdentity(config.DefaultIdentityTraitsSchemaID)
		original.Traits = identity.Traits(fmt.Sprintf(
			`{"email":%[1]q,"email_recovery":%[1]q,"unprotected":"before"}`, identifier))
		require.NoError(t, reg.IdentityManager().Create(t.Context(), original))

		original.Traits = identity.Traits(fmt.Sprintf(
			`{"email":%[1]q,"email_recovery":%[1]q,"unprotected":"after"}`, identifier))
		require.NoError(t, reg.IdentityManager().Update(t.Context(), original),
			"updating an unprotected trait must not require a privileged session")
	})
}

// TestRecoveryAddressesEqual covers the new order-independent comparison used by
// the privileged-access gate, mirroring VerifiableAddressesEqual.
func TestRecoveryAddressesEqual(t *testing.T) {
	t.Parallel()

	id := x.NewUUID()
	a := identity.RecoveryAddress{Value: "a@ory.sh", Via: identity.AddressTypeEmail, IdentityID: id}
	b := identity.RecoveryAddress{Value: "b@ory.sh", Via: identity.AddressTypeEmail, IdentityID: id}

	for _, tc := range []struct {
		name             string
		original, update []identity.RecoveryAddress
		equal            bool
	}{
		{name: "both empty", original: nil, update: nil, equal: true},
		{name: "identical single", original: []identity.RecoveryAddress{a}, update: []identity.RecoveryAddress{a}, equal: true},
		{name: "different value", original: []identity.RecoveryAddress{a}, update: []identity.RecoveryAddress{b}, equal: false},
		{name: "added address", original: []identity.RecoveryAddress{a}, update: []identity.RecoveryAddress{a, b}, equal: false},
		{name: "removed address", original: []identity.RecoveryAddress{a, b}, update: []identity.RecoveryAddress{a}, equal: false},
		{name: "different via", original: []identity.RecoveryAddress{{Value: "a@ory.sh", Via: identity.AddressTypeEmail, IdentityID: id}}, update: []identity.RecoveryAddress{{Value: "a@ory.sh", Via: identity.AddressTypeSMS, IdentityID: id}}, equal: false},
	} {
		t.Run("case="+tc.name, func(t *testing.T) {
			assert.Equal(t, tc.equal, identity.RecoveryAddressesEqual(tc.original, tc.update))
		})
	}
}
