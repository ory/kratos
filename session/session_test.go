package session_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/stretchr/testify/assert"

	"github.com/ory/kratos/driver/config"
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
		s, _ := session.NewActiveSession(i, conf, authAt, identity.CredentialsTypePassword, identity.AuthenticatorAssuranceLevel1)
		assert.True(t, s.IsActive())
		require.NotEmpty(t, s.Token)
		require.NotEmpty(t, s.LogoutToken)
		assert.EqualValues(t, identity.CredentialsTypePassword, s.AMR[0].Method)

		i = new(identity.Identity)
		s, err := session.NewActiveSession(i, conf, authAt, identity.CredentialsTypePassword, identity.AuthenticatorAssuranceLevel1)
		assert.Nil(t, s)
		assert.ErrorIs(t, err, session.ErrIdentityDisabled)
	})

	t.Run("case=expired", func(t *testing.T) {
		assert.False(t, (&session.Session{ExpiresAt: time.Now().Add(time.Hour)}).IsActive())
		assert.False(t, (&session.Session{Active: true}).IsActive())
	})

	t.Run("case=amr", func(t *testing.T) {
		s := session.NewInactiveSession()
		s.CompletedLoginFor(identity.CredentialsTypeOIDC, identity.AuthenticatorAssuranceLevel1)
		assert.EqualValues(t, identity.CredentialsTypeOIDC, s.AMR[0].Method)
		s.CompletedLoginFor(identity.CredentialsTypeRecoveryLink, identity.AuthenticatorAssuranceLevel1)
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

	for k, tc := range []struct {
		d        string
		methods  []session.AuthenticationMethod
		expected identity.AuthenticatorAssuranceLevel
	}{
		{
			d:        "no amr means no assurance",
			expected: identity.NoAuthenticatorAssuranceLevel,
		},
		{
			d:        "password is aal1",
			methods:  []session.AuthenticationMethod{{Method: identity.CredentialsTypePassword}},
			expected: identity.AuthenticatorAssuranceLevel1,
		},
		{
			d:        "oidc is aal1",
			methods:  []session.AuthenticationMethod{{Method: identity.CredentialsTypeOIDC}},
			expected: identity.AuthenticatorAssuranceLevel1,
		},
		{
			d:        "recovery is aal1",
			methods:  []session.AuthenticationMethod{{Method: identity.CredentialsTypeRecoveryLink}},
			expected: identity.AuthenticatorAssuranceLevel1,
		},
		{
			d: "mix of password, oidc, recovery is still aal1",
			methods: []session.AuthenticationMethod{
				{Method: identity.CredentialsTypeRecoveryLink},
				{Method: identity.CredentialsTypeOIDC},
				{Method: identity.CredentialsTypePassword},
			},
			expected: identity.AuthenticatorAssuranceLevel1,
		},
		{
			d:        "just totp is gracefully aal1",
			methods:  []session.AuthenticationMethod{{Method: identity.CredentialsTypeTOTP}},
			expected: identity.AuthenticatorAssuranceLevel1,
		},
		{
			d: "password + totp is aal2",
			methods: []session.AuthenticationMethod{
				{Method: identity.CredentialsTypePassword},
				{Method: identity.CredentialsTypeTOTP},
			},
			expected: identity.AuthenticatorAssuranceLevel2,
		},
		{
			d: "password + lookup is aal2",
			methods: []session.AuthenticationMethod{
				{Method: identity.CredentialsTypePassword},
				{Method: identity.CredentialsTypeLookup},
			},
			expected: identity.AuthenticatorAssuranceLevel2,
		},
		{
			d: "oidc + totp is aal2",
			methods: []session.AuthenticationMethod{
				{Method: identity.CredentialsTypeOIDC},
				{Method: identity.CredentialsTypeTOTP},
			},
			expected: identity.AuthenticatorAssuranceLevel2,
		},
		{
			d: "oidc + lookup is aal2",
			methods: []session.AuthenticationMethod{
				{Method: identity.CredentialsTypeOIDC},
				{Method: identity.CredentialsTypeLookup},
			},
			expected: identity.AuthenticatorAssuranceLevel2,
		},
		{
			d: "recovery link + totp is aal2",
			methods: []session.AuthenticationMethod{
				{Method: identity.CredentialsTypeRecoveryLink},
				{Method: identity.CredentialsTypeTOTP},
			},
			expected: identity.AuthenticatorAssuranceLevel2,
		},
		{
			d: "recovery link + lookup is aal2",
			methods: []session.AuthenticationMethod{
				{Method: identity.CredentialsTypeRecoveryLink},
				{Method: identity.CredentialsTypeLookup},
			},
			expected: identity.AuthenticatorAssuranceLevel2,
		},
		{
			d: "recovery link + passwordless webauth is aal1",
			methods: []session.AuthenticationMethod{
				{Method: identity.CredentialsTypeRecoveryLink},
				{Method: identity.CredentialsTypeWebAuthn, AAL: identity.AuthenticatorAssuranceLevel1},
			},
			expected: identity.AuthenticatorAssuranceLevel1,
		},
		{
			d: "respects AAL on AAL1",
			methods: []session.AuthenticationMethod{
				{Method: identity.CredentialsTypePassword, AAL: identity.AuthenticatorAssuranceLevel1},
				{Method: identity.CredentialsTypeWebAuthn, AAL: identity.AuthenticatorAssuranceLevel1},
			},
			expected: identity.AuthenticatorAssuranceLevel1,
		},
		{
			d: "respects AAL on AAL2 without AAL1",
			methods: []session.AuthenticationMethod{
				{Method: identity.CredentialsTypePassword, AAL: identity.AuthenticatorAssuranceLevel2},
				{Method: identity.CredentialsTypeWebAuthn, AAL: identity.AuthenticatorAssuranceLevel2},
			},
			expected: identity.AuthenticatorAssuranceLevel1,
		},
		{
			d: "respects AAL on AAL2",
			methods: []session.AuthenticationMethod{
				{Method: identity.CredentialsTypePassword, AAL: identity.AuthenticatorAssuranceLevel1},
				{Method: identity.CredentialsTypeWebAuthn, AAL: identity.AuthenticatorAssuranceLevel2},
			},
			expected: identity.AuthenticatorAssuranceLevel2,
		},
	} {
		t.Run(fmt.Sprintf("case=%d/description=%s", k, tc.d), func(t *testing.T) {
			s := session.NewInactiveSession()
			for _, m := range tc.methods {
				s.CompletedLoginFor(m.Method, m.AAL)
			}
		})
	}

	t.Run("case=session refresh", func(t *testing.T) {
		conf.MustSet(config.ViperKeySessionLifespan, "24h")
		conf.MustSet(config.ViperKeySessionRefreshMinTimeLeft, "12h")
		t.Cleanup(func() {
			conf.MustSet(config.ViperKeySessionLifespan, "1m")
			conf.MustSet(config.ViperKeySessionRefreshMinTimeLeft, "1m")
		})
		i := new(identity.Identity)
		i.State = identity.StateActive
		s, _ := session.NewActiveSession(i, conf, authAt, identity.CredentialsTypePassword, identity.AuthenticatorAssuranceLevel1)
		assert.False(t, s.CanBeRefreshed(conf), "fresh session is not refreshable")

		s.ExpiresAt = s.ExpiresAt.Add(-12 * time.Hour)
		assert.True(t, s.CanBeRefreshed(conf), "session is refreshable after 12hrs")
	})
}
