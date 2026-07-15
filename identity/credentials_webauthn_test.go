// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package identity

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
)

func TestCredentialConversion(t *testing.T) {
	expected := &webauthn.Credential{
		ID:              []byte("abcdef"),
		PublicKey:       []byte("foobar"),
		AttestationType: "test",
		Authenticator: webauthn.Authenticator{
			AAGUID:       []byte("baz"),
			SignCount:    1,
			CloneWarning: false,
		},
	}

	actual := CredentialFromWebAuthn(expected, false).ToWebAuthn()
	assert.Equal(t, expected, actual)

	actualList := CredentialsWebAuthn{*CredentialFromWebAuthn(expected, false)}.ToWebAuthnFiltered(AuthenticatorAssuranceLevel2, nil)
	assert.Equal(t, []webauthn.Credential{*expected}, actualList)

	actualList = CredentialsWebAuthn{*CredentialFromWebAuthn(expected, true)}.ToWebAuthnFiltered(AuthenticatorAssuranceLevel1, nil)
	assert.Equal(t, []webauthn.Credential{*expected}, actualList)

	actualList = CredentialsWebAuthn{*CredentialFromWebAuthn(expected, true)}.ToWebAuthnFiltered(AuthenticatorAssuranceLevel2, nil)
	assert.Len(t, actualList, 0)

	actualList = CredentialsWebAuthn{*CredentialFromWebAuthn(expected, false)}.ToWebAuthnFiltered(AuthenticatorAssuranceLevel1, nil)
	assert.Len(t, actualList, 0)

	fromWebAuthn := CredentialFromWebAuthn(expected, true)
	assert.True(t, fromWebAuthn.IsPasswordless)
	fromWebAuthn = CredentialFromWebAuthn(expected, false)
	assert.False(t, fromWebAuthn.IsPasswordless)

	expected.Authenticator.AAGUID = uuid.Must(uuid.FromString("ea9b8d66-4d01-1d21-3ce4-b6b48cb575d4")).Bytes()
	fromWebAuthn = CredentialFromWebAuthn(expected, false)
	assert.Equal(t, "Google Password Manager", fromWebAuthn.DisplayName)
}

func TestUpdateFromLogin(t *testing.T) {
	t.Parallel()

	// newCreds builds a single stored credential with the given starting sign count.
	newCreds := func(signCount uint32) CredentialsWebAuthn {
		return CredentialsWebAuthn{{
			ID:            []byte("cred-1"),
			Authenticator: &AuthenticatorWebAuthn{SignCount: signCount},
		}}
	}

	// validated mimics the *webauthn.Credential returned by the library: it has
	// the matching ID and the post-assertion authenticator state.
	validated := func(id []byte, signCount uint32, cloneWarning bool) *webauthn.Credential {
		return &webauthn.Credential{
			ID: id,
			Authenticator: webauthn.Authenticator{
				SignCount:    signCount,
				CloneWarning: cloneWarning,
			},
		}
	}

	t.Run("case=advances the stored sign count", func(t *testing.T) {
		c := newCreds(3)
		assert.True(t, c.UpdateFromLogin(validated([]byte("cred-1"), 10, false)))
		assert.EqualValues(t, 10, c[0].Authenticator.SignCount)
		assert.False(t, c[0].Authenticator.CloneWarning)
	})

	t.Run("case=flags a clone when the counter did not advance", func(t *testing.T) {
		// go-webauthn keeps SignCount unchanged and sets CloneWarning when the
		// counter regresses or stalls (on a non-zero counter). The clone warning
		// is newly set, so the config must be persisted.
		c := newCreds(10)
		assert.True(t, c.UpdateFromLogin(validated([]byte("cred-1"), 10, true)))
		assert.True(t, c[0].Authenticator.CloneWarning)
		assert.EqualValues(t, 10, c[0].Authenticator.SignCount)
	})

	t.Run("case=does not lower a regressing sign count", func(t *testing.T) {
		// The library never returns a lower counter (it flags a clone and leaves
		// the count unchanged), but UpdateFromLogin must never decrement the
		// stored counter, which would weaken clone detection.
		c := newCreds(10)
		assert.False(t, c.UpdateFromLogin(validated([]byte("cred-1"), 5, false)))
		assert.EqualValues(t, 10, c[0].Authenticator.SignCount)
	})

	t.Run("case=synced passkey reporting a static zero counter is a no-op", func(t *testing.T) {
		// Synced passkeys (Apple/Google) always report 0; go-webauthn leaves the
		// counter at 0 and never raises a clone warning, so nothing changes.
		c := newCreds(0)
		assert.False(t, c.UpdateFromLogin(validated([]byte("cred-1"), 0, false)))
		assert.EqualValues(t, 0, c[0].Authenticator.SignCount)
	})

	t.Run("case=no matching credential is a no-op", func(t *testing.T) {
		c := newCreds(3)
		assert.False(t, c.UpdateFromLogin(validated([]byte("other"), 99, false)))
		assert.EqualValues(t, 3, c[0].Authenticator.SignCount)
	})

	t.Run("case=nil validated credential is a no-op", func(t *testing.T) {
		c := newCreds(3)
		assert.False(t, c.UpdateFromLogin(nil))
		assert.EqualValues(t, 3, c[0].Authenticator.SignCount)
	})

	t.Run("case=initialises a missing authenticator before updating", func(t *testing.T) {
		c := CredentialsWebAuthn{{ID: []byte("cred-1")}}
		assert.True(t, c.UpdateFromLogin(validated([]byte("cred-1"), 5, false)))
		require.NotNil(t, c[0].Authenticator)
		assert.EqualValues(t, 5, c[0].Authenticator.SignCount)
	})

	// withBackup builds a stored credential that carries backup flags.
	withBackup := func(signCount uint32, backupState bool) CredentialsWebAuthn {
		return CredentialsWebAuthn{{
			ID:            []byte("cred-1"),
			Authenticator: &AuthenticatorWebAuthn{SignCount: signCount},
			Flags:         &CredentialWebAuthnFlags{BackupEligible: true, BackupState: backupState},
		}}
	}

	// validatedWithBackup is like validated but also reports the post-assertion
	// backup flags the library copies from the authenticator data.
	validatedWithBackup := func(id []byte, signCount uint32, backupState bool) *webauthn.Credential {
		v := validated(id, signCount, false)
		v.Flags = webauthn.CredentialFlags{BackupEligible: true, BackupState: backupState}
		return v
	}

	t.Run("case=refreshes backup state when the credential gets synced", func(t *testing.T) {
		// A passkey registered device-bound (BS=false) and later synced to
		// iCloud/Google reports BS=true on the next login.
		c := withBackup(3, false)
		assert.True(t, c.UpdateFromLogin(validatedWithBackup([]byte("cred-1"), 4, true)))
		assert.True(t, c[0].Flags.BackupState)
	})

	t.Run("case=a backup state change alone triggers a persist", func(t *testing.T) {
		// The counter is unchanged, but the credential stopped being synced, so
		// the flipped backup state must still be written back.
		c := withBackup(3, true)
		assert.True(t, c.UpdateFromLogin(validatedWithBackup([]byte("cred-1"), 3, false)))
		assert.False(t, c[0].Flags.BackupState)
	})

	t.Run("case=an unchanged backup state is a no-op", func(t *testing.T) {
		c := withBackup(3, true)
		assert.False(t, c.UpdateFromLogin(validatedWithBackup([]byte("cred-1"), 3, true)))
		assert.True(t, c[0].Flags.BackupState)
	})
}

func TestPasswordlessOnly(t *testing.T) {
	a := *CredentialFromWebAuthn(&webauthn.Credential{ID: []byte("a")}, false)
	b := *CredentialFromWebAuthn(&webauthn.Credential{ID: []byte("b")}, false)
	c := *CredentialFromWebAuthn(&webauthn.Credential{ID: []byte("c")}, true)
	d := *CredentialFromWebAuthn(&webauthn.Credential{ID: []byte("d")}, false)
	e := *CredentialFromWebAuthn(&webauthn.Credential{ID: []byte("e")}, true)
	expected := CredentialsWebAuthn{a, b, c, d, e}

	actual := expected.PasswordlessOnly(nil)
	require.Len(t, actual, 2)
	assert.Equal(t, []webauthn.Credential{*c.ToWebAuthn(), *e.ToWebAuthn()}, actual)
}
