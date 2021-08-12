package session_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/stretchr/testify/assert"

	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/session"
)

func TestSession(t *testing.T) {
	conf, _ := internal.NewFastRegistryWithMocks(t)
	authAt := time.Now()

	t.Run("case=active session", func(t *testing.T) {
		i := new(identity.Identity)
		i.State = identity.StateActive
		s, _ := session.NewActiveSession(i, conf, authAt, identity.CredentialsTypePassword)
		assert.True(t, s.IsActive())
		require.NotEmpty(t, s.Token)
		require.NotEmpty(t, s.LogoutToken)
		assert.EqualValues(t, identity.CredentialsTypePassword, s.AMR[0].Method)

		i = new(identity.Identity)
		s, err := session.NewActiveSession(i, conf, authAt, identity.CredentialsTypePassword)
		assert.Nil(t, s)
		assert.ErrorIs(t, err, session.ErrIdentityDisabled)
	})

	t.Run("case=expired", func(t *testing.T) {
		assert.False(t, (&session.Session{ExpiresAt: time.Now().Add(time.Hour)}).IsActive())
		assert.False(t, (&session.Session{Active: true}).IsActive())
	})

	t.Run("case=amr", func(t *testing.T) {
		s := session.NewInactiveSession()
		s.CompletedLoginFor(identity.CredentialsTypeOIDC)
		assert.EqualValues(t, identity.CredentialsTypeOIDC, s.AMR[0].Method)
		s.CompletedLoginFor(identity.CredentialsTypeRecoveryLink)
		assert.EqualValues(t, identity.CredentialsTypeOIDC, s.AMR[0].Method)
		assert.EqualValues(t, identity.CredentialsTypeRecoveryLink, s.AMR[1].Method)
	})

	t.Run("case=activate", func(t *testing.T) {
		s := session.NewInactiveSession()
		require.NoError(t, s.Activate(&identity.Identity{State: identity.StateActive}, conf, authAt))
		assert.True(t, s.Active)
		assert.Equal(t, identity.NoAuthenticatorAssuranceLevel, s.AuthenticatorAssuranceLevel)
		assert.Equal(t, authAt, s.AuthenticatedAt)

		s = session.NewInactiveSession()
		require.ErrorIs(t, s.Activate(&identity.Identity{State: identity.StateInactive}, conf, authAt), session.ErrIdentityDisabled)
		assert.False(t, s.Active)
		assert.Equal(t, identity.NoAuthenticatorAssuranceLevel, s.AuthenticatorAssuranceLevel)
		assert.Empty(t, s.AuthenticatedAt)
	})

	t.Run("case=aal", func(t *testing.T) {
		for _, tc := range []struct {
			d        string
			methods  []identity.CredentialsType
			expected identity.AuthenticatorAssuranceLevel
		}{
			{
				d:        "no amr means no assurance",
				expected: identity.NoAuthenticatorAssuranceLevel,
			},
			{
				d:        "password is aal1",
				methods:  []identity.CredentialsType{identity.CredentialsTypePassword},
				expected: identity.AuthenticatorAssuranceLevel1,
			},
			{
				d:        "oidc is aal1",
				methods:  []identity.CredentialsType{identity.CredentialsTypeOIDC},
				expected: identity.AuthenticatorAssuranceLevel1,
			},
			{
				d:        "recovery is aal1",
				methods:  []identity.CredentialsType{identity.CredentialsTypeRecoveryLink},
				expected: identity.AuthenticatorAssuranceLevel1,
			},
			{
				d: "mix of password, oidc, recovery is still aal1",
				methods: []identity.CredentialsType{
					identity.CredentialsTypeRecoveryLink, identity.CredentialsTypeOIDC, identity.CredentialsTypePassword,
				},
				expected: identity.AuthenticatorAssuranceLevel1,
			},
			{
				d: "just totp is aal0",
				methods: []identity.CredentialsType{
					identity.CredentialsTypeTOTP,
				},
				expected: identity.NoAuthenticatorAssuranceLevel,
			},
			{
				d: "password + totp is aal2",
				methods: []identity.CredentialsType{
					identity.CredentialsTypePassword,
					identity.CredentialsTypeTOTP,
				},
				expected: identity.AuthenticatorAssuranceLevel2,
			},
			{
				d: "oidc + totp is aal2",
				methods: []identity.CredentialsType{
					identity.CredentialsTypeOIDC,
					identity.CredentialsTypeTOTP,
				},
				expected: identity.AuthenticatorAssuranceLevel2,
			},
			{
				d: "recovery link + totp is aal2",
				methods: []identity.CredentialsType{
					identity.CredentialsTypeRecoveryLink,
					identity.CredentialsTypeTOTP,
				},
				expected: identity.AuthenticatorAssuranceLevel2,
			},
		} {
			t.Run("case="+tc.d, func(t *testing.T) {
				s := session.NewInactiveSession()
				for _, m := range tc.methods {
					s.CompletedLoginFor(m)
				}

				s.SetAuthenticatorAssuranceLevel()
				assert.Equal(t, tc.expected, s.AuthenticatorAssuranceLevel)
			})
		}
	})
}
