// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package identity

import (
	"testing"

	"github.com/gofrs/uuid"
	"github.com/mohae/deepcopy"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

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

func TestCredentials_Hash(t *testing.T) {
	baseID := uuid.Must(uuid.NewV4())
	baseNID := uuid.Must(uuid.NewV4())

	for _, tc := range []struct {
		name        string
		cred1       Credentials
		cred2       Credentials
		expectEqual bool
		description string
	}{
		{
			name: "same json with different whitespace",
			cred1: Credentials{
				Type:        CredentialsTypeOIDC,
				Identifiers: []string{"test@example.com"},
				Config:      sqlxx.JSONRawMessage(`{"foo":"bar","baz":"qux"}`),
				Version:     1,
			},
			cred2: Credentials{
				Type:        CredentialsTypeOIDC,
				Identifiers: []string{"test@example.com"},
				Config: sqlxx.JSONRawMessage(`{
				"foo": "bar",
				"baz": "qux"
			}`),
				Version: 1,
			},
			expectEqual: true,
			description: "hashes should be equal for same JSON with different whitespace",
		},
		{
			name: "same json with different key order",
			cred1: Credentials{
				Type:        CredentialsTypeCodeAuth,
				Identifiers: []string{"test@example.com"},
				Config:      sqlxx.JSONRawMessage(`{"addresses":[{"address":"test@example.com","channel":"email"}]}`),
				Version:     1,
			},
			cred2: Credentials{
				Type:        CredentialsTypeCodeAuth,
				Identifiers: []string{"test@example.com"},
				Config:      sqlxx.JSONRawMessage(`{"addresses":[{"channel":"email","address":"test@example.com"}]}`),
				Version:     1,
			},
			expectEqual: true,
			description: "hashes should be equal for same JSON with different key order",
		},
		{
			name: "nested json with different key order",
			cred1: Credentials{
				Type:        CredentialsTypeWebAuthn,
				Identifiers: []string{"test@example.com"},
				Config:      sqlxx.JSONRawMessage(`{"credentials":[{"id":"abc","public_key":"xyz","type":"webauthn"}]}`),
				Version:     1,
			},
			cred2: Credentials{
				Type:        CredentialsTypeWebAuthn,
				Identifiers: []string{"test@example.com"},
				Config:      sqlxx.JSONRawMessage(`{"credentials":[{"type":"webauthn","public_key":"xyz","id":"abc"}]}`),
				Version:     1,
			},
			expectEqual: true,
			description: "hashes should be equal for nested JSON with different key order",
		},
		{
			name: "same identifiers in different order",
			cred1: Credentials{
				Type:        CredentialsTypeOIDC,
				Identifiers: []string{"a@example.com", "b@example.com", "c@example.com"},
				Config:      sqlxx.JSONRawMessage(`{}`),
				Version:     1,
			},
			cred2: Credentials{
				Type:        CredentialsTypeOIDC,
				Identifiers: []string{"c@example.com", "a@example.com", "b@example.com"},
				Config:      sqlxx.JSONRawMessage(`{}`),
				Version:     1,
			},
			expectEqual: true,
			description: "hashes should be equal for same identifiers in different order",
		},
		{
			name: "different json config",
			cred1: Credentials{
				Type:        CredentialsTypeOIDC,
				Identifiers: []string{"test@example.com"},
				Config:      sqlxx.JSONRawMessage(`{"foo":"bar"}`),
				Version:     1,
			},
			cred2: Credentials{
				Type:        CredentialsTypeOIDC,
				Identifiers: []string{"test@example.com"},
				Config:      sqlxx.JSONRawMessage(`{"foo":"different"}`),
				Version:     1,
			},
			expectEqual: false,
			description: "hashes should be different for different JSON content",
		},
		{
			name: "different types",
			cred1: Credentials{
				Type:        CredentialsTypePassword,
				Identifiers: []string{"test@example.com"},
				Config:      sqlxx.JSONRawMessage(`{"foo":"bar"}`),
				Version:     1,
			},
			cred2: Credentials{
				Type:        CredentialsTypeOIDC,
				Identifiers: []string{"test@example.com"},
				Config:      sqlxx.JSONRawMessage(`{"foo":"bar"}`),
				Version:     1,
			},
			expectEqual: false,
			description: "hashes should be different for different types",
		},
		{
			name: "different identifiers",
			cred1: Credentials{
				Type:        CredentialsTypeOIDC,
				Identifiers: []string{"test1@example.com"},
				Config:      sqlxx.JSONRawMessage(`{"foo":"bar"}`),
				Version:     1,
			},
			cred2: Credentials{
				Type:        CredentialsTypeOIDC,
				Identifiers: []string{"test2@example.com"},
				Config:      sqlxx.JSONRawMessage(`{"foo":"bar"}`),
				Version:     1,
			},
			expectEqual: false,
			description: "hashes should be different for different identifiers",
		},
		{
			name: "different versions",
			cred1: Credentials{
				Type:        CredentialsTypeOIDC,
				Identifiers: []string{"test@example.com"},
				Config:      sqlxx.JSONRawMessage(`{"foo":"bar"}`),
				Version:     1,
			},
			cred2: Credentials{
				Type:        CredentialsTypeOIDC,
				Identifiers: []string{"test@example.com"},
				Config:      sqlxx.JSONRawMessage(`{"foo":"bar"}`),
				Version:     2,
			},
			expectEqual: false,
			description: "hashes should be different for different versions",
		},
		{
			name: "different identity IDs",
			cred1: Credentials{
				Type:        CredentialsTypeOIDC,
				Identifiers: []string{"test@example.com"},
				Config:      sqlxx.JSONRawMessage(`{"foo":"bar"}`),
				Version:     1,
				IdentityID:  baseID,
			},
			cred2: Credentials{
				Type:        CredentialsTypeOIDC,
				Identifiers: []string{"test@example.com"},
				Config:      sqlxx.JSONRawMessage(`{"foo":"bar"}`),
				Version:     1,
				IdentityID:  uuid.Must(uuid.NewV4()),
			},
			expectEqual: false,
			description: "hashes should be different for different identity IDs",
		},
		{
			name: "different NIDs",
			cred1: Credentials{
				Type:        CredentialsTypeOIDC,
				Identifiers: []string{"test@example.com"},
				Config:      sqlxx.JSONRawMessage(`{"foo":"bar"}`),
				Version:     1,
				NID:         baseNID,
			},
			cred2: Credentials{
				Type:        CredentialsTypeOIDC,
				Identifiers: []string{"test@example.com"},
				Config:      sqlxx.JSONRawMessage(`{"foo":"bar"}`),
				Version:     1,
				NID:         uuid.Must(uuid.NewV4()),
			},
			expectEqual: false,
			description: "hashes should be different for different NIDs",
		},
	} {
		t.Run("case="+tc.name, func(t *testing.T) {
			hash1 := tc.cred1.Signature()
			hash2 := tc.cred2.Signature()

			if tc.expectEqual {
				assert.Equal(t, hash1, hash2, tc.description)
			} else {
				assert.NotEqual(t, hash1, hash2, tc.description)
			}
		})
	}
}
