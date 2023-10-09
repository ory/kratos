// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package identity

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/mohae/deepcopy"
	"github.com/stretchr/testify/assert"

	"github.com/ory/x/sqlxx"
)

func TestCredentialsEqual(t *testing.T) {
	original := map[CredentialsType]Credentials{
		"foo": {Type: "foo", Identifiers: []string{"bar"}, Config: sqlxx.JSONRawMessage(`{"foo":"bar"}`)},
	}

	derived := deepcopy.Copy(original).(map[CredentialsType]Credentials)
	assert.EqualValues(t, original, derived)
	derived["foo"].Identifiers[0] = "baz"
	assert.NotEqual(t, original, derived)
}

func TestAALOrder(t *testing.T) {
	assert.True(t, NoAuthenticatorAssuranceLevel < AuthenticatorAssuranceLevel1)
	assert.True(t, AuthenticatorAssuranceLevel1 < AuthenticatorAssuranceLevel2)
}

func TestParseCredentialsType(t *testing.T) {
	for _, tc := range []struct {
		input    string
		expected CredentialsType
	}{
		{"password", CredentialsTypePassword},
		{"oidc", CredentialsTypeOIDC},
		{"totp", CredentialsTypeTOTP},
		{"webauthn", CredentialsTypeWebAuthn},
		{"lookup_secret", CredentialsTypeLookup},
		{"link_recovery", CredentialsTypeRecoveryLink},
		{"code_recovery", CredentialsTypeRecoveryCode},
	} {
		t.Run("case="+tc.input, func(t *testing.T) {
			actual, ok := ParseCredentialsType(tc.input)
			require.True(t, ok)
			assert.Equal(t, tc.expected, actual)
		})
	}

	t.Run("case=unknown", func(t *testing.T) {
		_, ok := ParseCredentialsType("unknown")
		require.False(t, ok)
	})
}
