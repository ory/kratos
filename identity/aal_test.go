package identity

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDetermineAAL(t *testing.T) {

	for _, tc := range []struct {
		d        string
		methods  []CredentialsType
		expected AuthenticatorAssuranceLevel
	}{
		{
			d:        "no amr means no assurance",
			expected: NoAuthenticatorAssuranceLevel,
		},
		{
			d:        "password is aal1",
			methods:  []CredentialsType{CredentialsTypePassword},
			expected: AuthenticatorAssuranceLevel1,
		},
		{
			d:        "oidc is aal1",
			methods:  []CredentialsType{CredentialsTypeOIDC},
			expected: AuthenticatorAssuranceLevel1,
		},
		{
			d:        "recovery is aal1",
			methods:  []CredentialsType{CredentialsTypeRecoveryLink},
			expected: AuthenticatorAssuranceLevel1,
		},
		{
			d: "mix of password, oidc, recovery is still aal1",
			methods: []CredentialsType{
				CredentialsTypeRecoveryLink, CredentialsTypeOIDC, CredentialsTypePassword,
			},
			expected: AuthenticatorAssuranceLevel1,
		},
		{
			d: "just totp is aal0",
			methods: []CredentialsType{
				CredentialsTypeTOTP,
			},
			expected: NoAuthenticatorAssuranceLevel,
		},
		{
			d: "password + totp is aal2",
			methods: []CredentialsType{
				CredentialsTypePassword,
				CredentialsTypeTOTP,
			},
			expected: AuthenticatorAssuranceLevel2,
		},
		{
			d: "password + lookup is aal2",
			methods: []CredentialsType{
				CredentialsTypePassword,
				CredentialsTypeLookup,
			},
			expected: AuthenticatorAssuranceLevel2,
		},
		{
			d: "password + webauthn is aal2",
			methods: []CredentialsType{
				CredentialsTypePassword,
				CredentialsTypeWebAuthn,
			},
			expected: AuthenticatorAssuranceLevel2,
		},
		{
			d: "oidc + totp is aal2",
			methods: []CredentialsType{
				CredentialsTypeOIDC,
				CredentialsTypeTOTP,
			},
			expected: AuthenticatorAssuranceLevel2,
		},
		{
			d: "oidc + lookup is aal2",
			methods: []CredentialsType{
				CredentialsTypeOIDC,
				CredentialsTypeLookup,
			},
			expected: AuthenticatorAssuranceLevel2,
		},
		{
			d: "recovery link + totp is aal2",
			methods: []CredentialsType{
				CredentialsTypeRecoveryLink,
				CredentialsTypeTOTP,
			},
			expected: AuthenticatorAssuranceLevel2,
		},
		{
			d: "recovery link + lookup is aal2",
			methods: []CredentialsType{
				CredentialsTypeRecoveryLink,
				CredentialsTypeLookup,
			},
			expected: AuthenticatorAssuranceLevel2,
		},
	} {
		t.Run("case="+tc.d, func(t *testing.T) {
			assert.Equal(t, tc.expected, DetermineAAL(tc.methods))
		})
	}
}
