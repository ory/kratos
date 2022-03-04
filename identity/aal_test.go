package identity

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type fakeConfig struct {
	less bool
}

func (f *fakeConfig) WebAuthnForPasswordless() bool {
	return f.less
}

func TestDetermineAAL(t *testing.T) {
	for _, tc := range []struct {
		d                    string
		methods              map[CredentialsType]Credentials
		expected             AuthenticatorAssuranceLevel
		webAuthnPasswordless bool
	}{
		{
			d:        "no amr means no assurance",
			expected: NoAuthenticatorAssuranceLevel,
		},
		{
			d:        "password is aal1",
			methods:  map[CredentialsType]Credentials{CredentialsTypePassword: {Type: CredentialsTypePassword}},
			expected: AuthenticatorAssuranceLevel1,
		},
		{
			d:        "oidc is aal1",
			methods:  map[CredentialsType]Credentials{CredentialsTypeOIDC: {Type: CredentialsTypeOIDC}},
			expected: AuthenticatorAssuranceLevel1,
		},
		{
			d:        "recovery is aal1",
			methods:  map[CredentialsType]Credentials{CredentialsTypeRecoveryLink: {Type: CredentialsTypeRecoveryLink}},
			expected: AuthenticatorAssuranceLevel1,
		},
		{
			d:        "legacy is aal1",
			methods:  map[CredentialsType]Credentials{"v0.6_legacy_session": {Type: "v0.6_legacy_session"}},
			expected: AuthenticatorAssuranceLevel1,
		},
		{
			d: "mix of password, oidc, recovery is still aal1",
			methods: map[CredentialsType]Credentials{
				CredentialsTypeRecoveryLink: {Type: CredentialsTypeRecoveryLink},
				CredentialsTypeOIDC:         {Type: CredentialsTypeOIDC},
				CredentialsTypePassword:     {Type: CredentialsTypePassword},
			},
			expected: AuthenticatorAssuranceLevel1,
		},
		{
			d:        "just totp is aal0",
			methods:  map[CredentialsType]Credentials{CredentialsTypeTOTP: {Type: CredentialsTypeTOTP}},
			expected: NoAuthenticatorAssuranceLevel,
		},
		{
			d: "password + totp is aal2",
			methods: map[CredentialsType]Credentials{
				CredentialsTypeTOTP:     {Type: CredentialsTypeTOTP},
				CredentialsTypePassword: {Type: CredentialsTypePassword},
			},
			expected: AuthenticatorAssuranceLevel2,
		},
		{
			d: "password + lookup is aal2",
			methods: map[CredentialsType]Credentials{
				CredentialsTypeLookup:   {Type: CredentialsTypeLookup},
				CredentialsTypePassword: {Type: CredentialsTypePassword},
			},
			expected: AuthenticatorAssuranceLevel2,
		},
		{
			d: "password + webauthn is aal2",
			methods: map[CredentialsType]Credentials{
				CredentialsTypeWebAuthn: {Type: CredentialsTypeWebAuthn},
				CredentialsTypePassword: {Type: CredentialsTypePassword},
			},
			expected: AuthenticatorAssuranceLevel2,
		},
		{
			d: "oidc + totp is aal2",
			methods: map[CredentialsType]Credentials{
				CredentialsTypeOIDC: {Type: CredentialsTypeOIDC},
				CredentialsTypeTOTP: {Type: CredentialsTypeTOTP},
			},
			expected: AuthenticatorAssuranceLevel2,
		},
		{
			d: "oidc + lookup is aal2",
			methods: map[CredentialsType]Credentials{
				CredentialsTypeOIDC:   {Type: CredentialsTypeOIDC},
				CredentialsTypeLookup: {Type: CredentialsTypeLookup},
			},
			expected: AuthenticatorAssuranceLevel2,
		},
		{
			d: "recovery link + totp is aal2",
			methods: map[CredentialsType]Credentials{
				CredentialsTypeRecoveryLink: {Type: CredentialsTypeRecoveryLink},
				CredentialsTypeTOTP:         {Type: CredentialsTypeTOTP},
			},
			expected: AuthenticatorAssuranceLevel2,
		},
		{
			d: "recovery link + lookup is aal2",
			methods: map[CredentialsType]Credentials{
				CredentialsTypeRecoveryLink: {Type: CredentialsTypeRecoveryLink},
				CredentialsTypeLookup:       {Type: CredentialsTypeLookup},
			},
			expected: AuthenticatorAssuranceLevel2,
		},
		{
			d: "webauthn only is aal1 if passwordless is set",
			methods: map[CredentialsType]Credentials{
				CredentialsTypeWebAuthn: {Type: CredentialsTypeWebAuthn, Config: []byte(`{"credentials":[{"is_passwordless":true}]}`)},
			},
			expected:             AuthenticatorAssuranceLevel1,
			webAuthnPasswordless: true,
		},
		{
			d: "webauthn and another method is aal1 if passwordless is set",
			methods: map[CredentialsType]Credentials{
				CredentialsTypePassword: {Type: CredentialsTypePassword},
				CredentialsTypeWebAuthn: {Type: CredentialsTypeWebAuthn, Config: []byte(`{"credentials":[{"is_passwordless":true}]}`)},
			},
			expected: AuthenticatorAssuranceLevel1,
		},
		{
			d: "webauthn only is unknown if passwordless is not set",
			methods: map[CredentialsType]Credentials{
				CredentialsTypeWebAuthn: {Type: CredentialsTypeWebAuthn, Config: []byte(`{"credentials":[{"is_passwordless":false}]}`)},
			},
			expected: NoAuthenticatorAssuranceLevel,
		},
		{
			d: "webauthn with password is aal1 if passwordless is set",
			methods: map[CredentialsType]Credentials{
				CredentialsTypePassword: {Type: CredentialsTypePassword},
				CredentialsTypeWebAuthn: {Type: CredentialsTypeWebAuthn, Config: []byte(`{"credentials":[{"is_passwordless":true}]}`)},
			},
			expected: AuthenticatorAssuranceLevel1,
		},
		{
			d: "webauthn with password and two credentials is aal2 if no passwordless",
			methods: map[CredentialsType]Credentials{
				CredentialsTypePassword: {Type: CredentialsTypePassword},
				CredentialsTypeWebAuthn: {Type: CredentialsTypeWebAuthn, Config: []byte(`{"credentials":[{"is_passwordless":false},{"is_passwordless":true}]}`)},
			},
			expected: AuthenticatorAssuranceLevel2,
		},
		{
			d: "webauthn with password and two credentials is aal1 if passwordless",
			methods: map[CredentialsType]Credentials{
				CredentialsTypePassword: {Type: CredentialsTypePassword},
				CredentialsTypeWebAuthn: {Type: CredentialsTypeWebAuthn, Config: []byte(`{"credentials":[{"is_passwordless":false},{"is_passwordless":true}]}`)},
			},
			expected:             AuthenticatorAssuranceLevel1,
			webAuthnPasswordless: true,
		},
		{
			d: "webauthn with password and two credentials without passwordless key is aal2",
			methods: map[CredentialsType]Credentials{
				CredentialsTypePassword: {Type: CredentialsTypePassword},
				CredentialsTypeWebAuthn: {Type: CredentialsTypeWebAuthn, Config: []byte(`{"credentials":[{},{}]}`)},
			},
			expected: AuthenticatorAssuranceLevel2,
		},
		{
			d: "webauthn with two credentials is still aal1 if passwordless",
			methods: map[CredentialsType]Credentials{
				CredentialsTypeWebAuthn: {Type: CredentialsTypeWebAuthn, Config: []byte(`{"credentials":[{"is_passwordless":false},{"is_passwordless":true}]}`)},
			},
			expected:             AuthenticatorAssuranceLevel1,
			webAuthnPasswordless: true,
		},
		{
			d: "webauthn with two credentials is no aal",
			methods: map[CredentialsType]Credentials{
				CredentialsTypeWebAuthn: {Type: CredentialsTypeWebAuthn, Config: []byte(`{"credentials":[{"is_passwordless":false},{"is_passwordless":true}]}`)},
			},
			expected: NoAuthenticatorAssuranceLevel,
		},
	} {
		t.Run("case="+tc.d, func(t *testing.T) {
			assert.Equal(t, tc.expected, MaximumAAL(tc.methods, &fakeConfig{less: tc.webAuthnPasswordless}))
		})
	}
}
