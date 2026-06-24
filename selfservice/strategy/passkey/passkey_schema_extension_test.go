// Copyright © 2026 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package passkey_test

import (
	"context"
	"net/url"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/pkg"
	"github.com/ory/kratos/pkg/testhelpers"
	"github.com/ory/kratos/selfservice/strategy/passkey"
)

func TestPasskeyDisplayNameFromSchema(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	_, reg := pkg.NewFastRegistryWithMocks(t)

	strategy := passkey.NewStrategy(reg)

	abs := func(t *testing.T, p string) string {
		t.Helper()
		a, err := filepath.Abs(p)
		require.NoError(t, err)
		return (&url.URL{Scheme: "file", Path: a}).String()
	}

	t.Run("case=single trait flagged returns single path", func(t *testing.T) {
		t.Parallel()
		paths, err := strategy.PasskeyDisplayNameFromSchema(ctx, abs(t, "stub/registration-single-identifier.schema.json"))
		require.NoError(t, err)
		assert.Equal(t, []string{"traits.username"}, paths)
	})

	t.Run("case=multiple traits flagged returns sorted paths", func(t *testing.T) {
		t.Parallel()
		paths, err := strategy.PasskeyDisplayNameFromSchema(ctx, abs(t, "stub/registration-multi-identifier.schema.json"))
		require.NoError(t, err)
		assert.Equal(t, []string{"traits.email", "traits.phone"}, paths)
	})

	t.Run("case=webauthn.identifier-only flagged trait returns its path", func(t *testing.T) {
		t.Parallel()
		paths, err := strategy.PasskeyDisplayNameFromSchema(ctx, abs(t, "stub/registration-webauthn-only.schema.json"))
		require.NoError(t, err)
		assert.Equal(t, []string{"traits.username"}, paths)
	})

	t.Run("case=schema without flagged traits falls back to the untitled trait", func(t *testing.T) {
		t.Parallel()
		// Preserves the legacy behavior: with no flagged trait, the first
		// untitled trait is used as the display-name source.
		paths, err := strategy.PasskeyDisplayNameFromSchema(ctx, abs(t, "stub/noid.schema.json"))
		require.NoError(t, err)
		assert.Equal(t, []string{"traits.email"}, paths)
	})

	t.Run("case=schema with only titled, unflagged traits returns an error", func(t *testing.T) {
		t.Parallel()
		paths, err := strategy.PasskeyDisplayNameFromSchema(ctx, abs(t, "stub/registration-no-passkey-titled.schema.json"))
		require.ErrorContains(t, err, "no identifier found")
		assert.Nil(t, paths)
	})
}

// TestPasskeyDisplayNameFromTraits_MultipleFlaggedTraits guards the resolution
// order when an identity schema flags more than one trait as a passkey
// display-name source. A flagged trait that the user left empty must not clobber
// a flagged trait that has a value, regardless of the order the validator visits
// the traits. This mirrors the client-side behavior, which picks the first
// non-empty candidate field.
func TestPasskeyDisplayNameFromTraits_MultipleFlaggedTraits(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	conf, reg := pkg.NewFastRegistryWithMocks(t)
	testhelpers.SetDefaultIdentitySchema(conf, "file://./stub/registration-multi-identifier.schema.json")

	strategy := passkey.NewStrategy(reg)

	t.Run("case=keeps the non-empty candidate when a later flagged trait is empty", func(t *testing.T) {
		t.Parallel()
		name := strategy.PasskeyDisplayNameFromTraits(ctx, identity.Traits(`{"email":"alice@example.com","phone":""}`))
		assert.Equal(t, "alice@example.com", name)
	})

	t.Run("case=keeps the non-empty candidate when an earlier flagged trait is empty", func(t *testing.T) {
		t.Parallel()
		name := strategy.PasskeyDisplayNameFromTraits(ctx, identity.Traits(`{"email":"","phone":"+15555550123"}`))
		assert.Equal(t, "+15555550123", name)
	})
}
