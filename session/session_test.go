// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package session_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/ory/kratos/x"

	"github.com/stretchr/testify/require"

	"github.com/stretchr/testify/assert"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/session"
)

func TestSession(t *testing.T) {
	ctx := context.Background()
	conf, _ := internal.NewFastRegistryWithMocks(t)
	authAt := time.Now()

	t.Run("case=active session", func(t *testing.T) {
		req := x.NewTestHTTPRequest(t, "GET", "/sessions/whoami", nil)

		i := new(identity.Identity)
		i.State = identity.StateActive
		s, _ := session.NewActiveSession(req, i, conf, authAt, identity.CredentialsTypePassword, identity.AuthenticatorAssuranceLevel1)
		assert.True(t, s.IsActive())
		require.NotEmpty(t, s.Token)
		require.NotEmpty(t, s.LogoutToken)
		assert.EqualValues(t, identity.CredentialsTypePassword, s.AMR[0].Method)

		i = new(identity.Identity)
		s, err := session.NewActiveSession(req, i, conf, authAt, identity.CredentialsTypePassword, identity.AuthenticatorAssuranceLevel1)
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
		s.CompletedLoginFor(identity.CredentialsTypeRecoveryCode, identity.AuthenticatorAssuranceLevel1)
		assert.EqualValues(t, identity.CredentialsTypeRecoveryCode, s.AMR[2].Method)
	})

	t.Run("case=activate", func(t *testing.T) {
		req := x.NewTestHTTPRequest(t, "GET", "/sessions/whoami", nil)

		s := session.NewInactiveSession()
		require.NoError(t, s.Activate(req, &identity.Identity{State: identity.StateActive}, conf, authAt))
		assert.True(t, s.Active)
		assert.Equal(t, identity.NoAuthenticatorAssuranceLevel, s.AuthenticatorAssuranceLevel)
		assert.Equal(t, authAt, s.AuthenticatedAt)

		s = session.NewInactiveSession()
		require.ErrorIs(t, s.Activate(req, &identity.Identity{State: identity.StateInactive}, conf, authAt), session.ErrIdentityDisabled)
		assert.False(t, s.Active)
		assert.Equal(t, identity.NoAuthenticatorAssuranceLevel, s.AuthenticatorAssuranceLevel)
		assert.Empty(t, s.AuthenticatedAt)
	})

	t.Run("case=client information reverse proxy forward", func(t *testing.T) {
		for _, tc := range []struct {
			input    string
			expected string
		}{
			{
				input:    "10.10.8.1, 172.19.2.7",
				expected: "",
			},
			{
				input:    "217.73.188.139,162.158.203.149, 172.19.2.7",
				expected: "162.158.203.149",
			},
			{
				input:    "122.122.122.122 , 123.123.123.123",
				expected: "123.123.123.123",
			},
		} {
			t.Run("case=parse "+tc.input, func(t *testing.T) {
				req := x.NewTestHTTPRequest(t, "GET", "/sessions/whoami", nil)
				req.Header["User-Agent"] = []string{"Mozilla/5.0 (X11; Linux x86_64)", "AppleWebKit/537.36 (KHTML, like Gecko)", "Chrome/51.0.2704.103 Safari/537.36"}
				req.Header.Set("X-Forwarded-For", tc.input)

				s := session.NewInactiveSession()
				require.NoError(t, s.Activate(req, &identity.Identity{State: identity.StateActive}, conf, authAt))
				assert.True(t, s.Active)
				assert.Equal(t, identity.NoAuthenticatorAssuranceLevel, s.AuthenticatorAssuranceLevel)
				assert.Equal(t, authAt, s.AuthenticatedAt)
				assert.Equal(t, 1, len(s.Devices))
				assert.Equal(t, s.ID.String(), s.Devices[0].SessionID.String())
				assert.Equal(t, tc.expected, *s.Devices[0].IPAddress)
				assert.Equal(t, "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/51.0.2704.103 Safari/537.36", *s.Devices[0].UserAgent)
				assert.Equal(t, "", *s.Devices[0].Location)
			})
		}
	})

	t.Run("case=client information reverse proxy real IP set", func(t *testing.T) {
		req := x.NewTestHTTPRequest(t, "GET", "/sessions/whoami", nil)
		req.Header["User-Agent"] = []string{"Mozilla/5.0 (X11; Linux x86_64)", "AppleWebKit/537.36 (KHTML, like Gecko)", "Chrome/51.0.2704.103 Safari/537.36"}
		req.Header.Set("X-Real-IP", "54.155.246.155")
		req.Header["X-Forwarded-For"] = []string{"54.155.246.232", "10.145.1.10"}

		s := session.NewInactiveSession()
		require.NoError(t, s.Activate(req, &identity.Identity{State: identity.StateActive}, conf, authAt))
		assert.True(t, s.Active)
		assert.Equal(t, identity.NoAuthenticatorAssuranceLevel, s.AuthenticatorAssuranceLevel)
		assert.Equal(t, authAt, s.AuthenticatedAt)
		assert.Equal(t, 1, len(s.Devices))
		assert.Equal(t, s.ID.String(), s.Devices[0].SessionID.String())
		assert.NotNil(t, s.Devices[0].UpdatedAt)
		assert.NotNil(t, s.Devices[0].CreatedAt)
		assert.Equal(t, "54.155.246.155", *s.Devices[0].IPAddress)
		assert.Equal(t, "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/51.0.2704.103 Safari/537.36", *s.Devices[0].UserAgent)
		assert.Equal(t, "", *s.Devices[0].Location)
	})

	t.Run("case=client information CF true client IP set", func(t *testing.T) {
		req := x.NewTestHTTPRequest(t, "GET", "/sessions/whoami", nil)
		req.Header["User-Agent"] = []string{"Mozilla/5.0 (X11; Linux x86_64)", "AppleWebKit/537.36 (KHTML, like Gecko)", "Chrome/51.0.2704.103 Safari/537.36"}
		req.Header.Set("True-Client-IP", "54.155.246.155")
		req.Header.Set("X-Forwarded-For", "217.73.188.139,162.158.203.149, 172.19.2.7")

		s := session.NewInactiveSession()
		require.NoError(t, s.Activate(req, &identity.Identity{State: identity.StateActive}, conf, authAt))
		assert.True(t, s.Active)
		assert.Equal(t, identity.NoAuthenticatorAssuranceLevel, s.AuthenticatorAssuranceLevel)
		assert.Equal(t, authAt, s.AuthenticatedAt)
		assert.Equal(t, 1, len(s.Devices))
		assert.Equal(t, s.ID.String(), s.Devices[0].SessionID.String())
		assert.NotNil(t, s.Devices[0].UpdatedAt)
		assert.NotNil(t, s.Devices[0].CreatedAt)
		assert.Equal(t, "54.155.246.155", *s.Devices[0].IPAddress)
		assert.Equal(t, "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/51.0.2704.103 Safari/537.36", *s.Devices[0].UserAgent)
		assert.Equal(t, "", *s.Devices[0].Location)
	})

	t.Run("case=client information CF", func(t *testing.T) {
		req := x.NewTestHTTPRequest(t, "GET", "/sessions/whoami", nil)
		req.Header["User-Agent"] = []string{"Mozilla/5.0 (X11; Linux x86_64)", "AppleWebKit/537.36 (KHTML, like Gecko)", "Chrome/51.0.2704.103 Safari/537.36"}
		req.Header.Set("True-Client-IP", "54.155.246.232")
		req.Header.Set("Cf-Ipcity", "Munich")
		req.Header.Set("Cf-Ipcountry", "Germany")

		s := session.NewInactiveSession()
		require.NoError(t, s.Activate(req, &identity.Identity{State: identity.StateActive}, conf, authAt))
		assert.True(t, s.Active)
		assert.Equal(t, identity.NoAuthenticatorAssuranceLevel, s.AuthenticatorAssuranceLevel)
		assert.Equal(t, authAt, s.AuthenticatedAt)
		assert.Equal(t, 1, len(s.Devices))
		assert.Equal(t, s.ID.String(), s.Devices[0].SessionID.String())
		assert.Equal(t, "54.155.246.232", *s.Devices[0].IPAddress)
		assert.Equal(t, "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/51.0.2704.103 Safari/537.36", *s.Devices[0].UserAgent)
		assert.Equal(t, "Munich, Germany", *s.Devices[0].Location)
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
			d: "recovery link is aal1",
			methods: []session.AuthenticationMethod{
				{Method: identity.CredentialsTypeRecoveryLink},
			},
			expected: identity.AuthenticatorAssuranceLevel1,
		},
		{
			d: "recovery code is aal1",
			methods: []session.AuthenticationMethod{
				{Method: identity.CredentialsTypeRecoveryCode},
			},
			expected: identity.AuthenticatorAssuranceLevel1,
		},
		{
			d: "mix of password, oidc, recovery link is still aal1",
			methods: []session.AuthenticationMethod{
				{Method: identity.CredentialsTypeRecoveryLink},
				{Method: identity.CredentialsTypeOIDC},
				{Method: identity.CredentialsTypePassword},
			},
			expected: identity.AuthenticatorAssuranceLevel1,
		},
		{
			d: "mix of password, oidc, recovery code is still aal1",
			methods: []session.AuthenticationMethod{
				{Method: identity.CredentialsTypeRecoveryCode},
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
			d: "recovery code + totp is aal2",
			methods: []session.AuthenticationMethod{
				{Method: identity.CredentialsTypeRecoveryCode},
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
			d: "recovery code + lookup is aal2",
			methods: []session.AuthenticationMethod{
				{Method: identity.CredentialsTypeRecoveryCode},
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
			d: "recovery code + passwordless webauth is aal1",
			methods: []session.AuthenticationMethod{
				{Method: identity.CredentialsTypeRecoveryCode},
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
		req := x.NewTestHTTPRequest(t, "GET", "/sessions/whoami", nil)

		conf.MustSet(ctx, config.ViperKeySessionLifespan, "24h")
		conf.MustSet(ctx, config.ViperKeySessionRefreshMinTimeLeft, "12h")
		t.Cleanup(func() {
			conf.MustSet(ctx, config.ViperKeySessionLifespan, "1m")
			conf.MustSet(ctx, config.ViperKeySessionRefreshMinTimeLeft, "1m")
		})
		i := new(identity.Identity)
		i.State = identity.StateActive
		s, _ := session.NewActiveSession(req, i, conf, authAt, identity.CredentialsTypePassword, identity.AuthenticatorAssuranceLevel1)
		assert.False(t, s.CanBeRefreshed(ctx, conf), "fresh session is not refreshable")

		s.ExpiresAt = s.ExpiresAt.Add(-12 * time.Hour)
		assert.True(t, s.CanBeRefreshed(ctx, conf), "session is refreshable after 12hrs")
	})
}
