// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package session_test

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/pkg"
	"github.com/ory/kratos/pkg/testhelpers"
	"github.com/ory/kratos/session"
	"github.com/ory/kratos/x"
	"github.com/ory/x/configx"
	"github.com/ory/x/contextx"
)

func TestAdminSessionMarshal(t *testing.T) {
	t.Parallel()

	metadataAdmin := []byte(`{"role":"admin"}`)
	ident := &identity.Identity{
		MetadataAdmin: metadataAdmin,
		State:         identity.StateActive,
	}
	sess := session.Session{
		ID:        x.NewUUID(),
		Active:    true,
		Identity:  ident,
		ExpiresAt: time.Now().Add(time.Hour),
	}

	t.Run("case=regular Session omits metadata_admin", func(t *testing.T) {
		t.Parallel()
		b, err := json.Marshal(sess)
		require.NoError(t, err)
		assert.False(t, gjson.GetBytes(b, "identity.metadata_admin").Exists())
		assert.Equal(t, sess.ID.String(), gjson.GetBytes(b, "id").String())
		assert.True(t, gjson.GetBytes(b, "active").Bool())
	})

	t.Run("case=AdminSession includes metadata_admin", func(t *testing.T) {
		t.Parallel()
		b, err := json.Marshal(session.AdminSession(sess))
		require.NoError(t, err)
		assert.Equal(t, `{"role":"admin"}`, gjson.GetBytes(b, "identity.metadata_admin").Raw)
		assert.Equal(t, sess.ID.String(), gjson.GetBytes(b, "id").String())
		assert.True(t, gjson.GetBytes(b, "active").Bool())
	})

	t.Run("case=AdminSession uses runtime active semantics for expired sessions", func(t *testing.T) {
		t.Parallel()

		expiredSess := session.Session{
			ID:        x.NewUUID(),
			Active:    true,
			Identity:  ident,
			ExpiresAt: time.Now().Add(-time.Minute),
		}

		b, err := json.Marshal(session.AdminSession(expiredSess))
		require.NoError(t, err)
		assert.Equal(t, `{"role":"admin"}`, gjson.GetBytes(b, "identity.metadata_admin").Raw)
		assert.Equal(t, expiredSess.ID.String(), gjson.GetBytes(b, "id").String())
		assert.False(t, gjson.GetBytes(b, "active").Bool())
	})

	t.Run("case=AdminSessions converts slice", func(t *testing.T) {
		t.Parallel()
		sessions := []session.Session{sess, sess}
		adminSessions := session.AdminSessions(sessions)
		require.Len(t, adminSessions, 2)
		for _, as := range adminSessions {
			b, err := json.Marshal(as)
			require.NoError(t, err)
			assert.Equal(t, `{"role":"admin"}`, gjson.GetBytes(b, "identity.metadata_admin").Raw)
			assert.Equal(t, sess.ID.String(), gjson.GetBytes(b, "id").String())
			assert.True(t, gjson.GetBytes(b, "active").Bool())
		}
	})
}

func TestSession(t *testing.T) {
	_, reg := pkg.NewFastRegistryWithMocks(t,
		configx.WithValues(testhelpers.DefaultIdentitySchemaConfig("file://./stub/identity.schema.json")),
	)
	authAt := time.Now()

	t.Run("case=active session", func(t *testing.T) {
		req := testhelpers.NewTestHTTPRequest(t, "GET", "/sessions/whoami", nil)

		i := new(identity.Identity)
		i.State = identity.StateActive
		i.NID = x.NewUUID()
		s, err := testhelpers.NewActiveSession(req, reg, i, authAt, identity.CredentialsTypePassword, identity.AuthenticatorAssuranceLevel1)
		require.NoError(t, err)
		assert.True(t, s.IsActive())
		require.NotEmpty(t, s.Token)
		require.NotEmpty(t, s.LogoutToken)
		assert.EqualValues(t, identity.CredentialsTypePassword, s.AMR[0].Method)

		i = new(identity.Identity)
		i.NID = x.NewUUID()
		s, err = testhelpers.NewActiveSession(req, reg, i, authAt, identity.CredentialsTypePassword, identity.AuthenticatorAssuranceLevel1)
		assert.Nil(t, s)
		assert.ErrorIs(t, err, session.ErrIdentityDisabled())
	})

	t.Run("case=expired", func(t *testing.T) {
		assert.False(t, (&session.Session{ExpiresAt: time.Now().Add(time.Hour)}).IsActive())
		assert.False(t, (&session.Session{Active: true}).IsActive())
	})

	t.Run("case=amr add", func(t *testing.T) {
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
		req := testhelpers.NewTestHTTPRequest(t, "GET", "/sessions/whoami", nil)

		s := session.NewInactiveSession()
		require.NoError(t, reg.SessionManager().ActivateSession(req, s, &identity.Identity{NID: x.NewUUID(), State: identity.StateActive}, authAt))
		assert.True(t, s.Active)
		assert.Equal(t, identity.NoAuthenticatorAssuranceLevel, s.AuthenticatorAssuranceLevel)
		assert.Equal(t, authAt, s.AuthenticatedAt)

		s = session.NewInactiveSession()
		require.ErrorIs(t, reg.SessionManager().ActivateSession(req, s, &identity.Identity{NID: x.NewUUID(), State: identity.StateInactive}, authAt), session.ErrIdentityDisabled())
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
				req := testhelpers.NewTestHTTPRequest(t, "GET", "/sessions/whoami", nil)
				req.Header["User-Agent"] = []string{"Mozilla/5.0 (X11; Linux x86_64)", "AppleWebKit/537.36 (KHTML, like Gecko)", "Chrome/51.0.2704.103 Safari/537.36"}
				req.Header.Set("X-Forwarded-For", tc.input)

				s := session.NewInactiveSession()
				require.NoError(t, reg.SessionManager().ActivateSession(req, s, &identity.Identity{NID: x.NewUUID(), State: identity.StateActive}, authAt))
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
		req := testhelpers.NewTestHTTPRequest(t, "GET", "/sessions/whoami", nil)
		req.Header["User-Agent"] = []string{"Mozilla/5.0 (X11; Linux x86_64)", "AppleWebKit/537.36 (KHTML, like Gecko)", "Chrome/51.0.2704.103 Safari/537.36"}
		req.Header.Set("X-Real-IP", "54.155.246.155")
		req.Header["X-Forwarded-For"] = []string{"54.155.246.232", "10.145.1.10"}

		s := session.NewInactiveSession()
		require.NoError(t, reg.SessionManager().ActivateSession(req, s, &identity.Identity{NID: x.NewUUID(), State: identity.StateActive}, authAt))
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
		req := testhelpers.NewTestHTTPRequest(t, "GET", "/sessions/whoami", nil)
		req.Header["User-Agent"] = []string{"Mozilla/5.0 (X11; Linux x86_64)", "AppleWebKit/537.36 (KHTML, like Gecko)", "Chrome/51.0.2704.103 Safari/537.36"}
		req.Header.Set("True-Client-IP", "54.155.246.155")
		req.Header.Set("X-Forwarded-For", "217.73.188.139,162.158.203.149, 172.19.2.7")

		s := session.NewInactiveSession()
		require.NoError(t, reg.SessionManager().ActivateSession(req, s, &identity.Identity{State: identity.StateActive, NID: x.NewUUID()}, authAt))
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
		req := testhelpers.NewTestHTTPRequest(t, "GET", "/sessions/whoami", nil)
		req.Header["User-Agent"] = []string{"Mozilla/5.0 (X11; Linux x86_64)", "AppleWebKit/537.36 (KHTML, like Gecko)", "Chrome/51.0.2704.103 Safari/537.36"}
		req.Header.Set("True-Client-IP", "54.155.246.232")
		req.Header.Set("Cf-Ipcity", "Munich")
		req.Header.Set("Cf-Ipcountry", "Germany")

		s := session.NewInactiveSession()
		require.NoError(t, reg.SessionManager().ActivateSession(req, s, &identity.Identity{NID: x.NewUUID(), State: identity.StateActive}, authAt))
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

	t.Run("case=federated AAL2 carry-over", func(t *testing.T) {
		// A single OIDC or SAML method tagged AAL2 elevates the
		// session to AAL2 on its own. This is the upstream-MFA
		// carry-over path: the provider config matched the upstream
		// `acr` or `amr` claims, so Ory trusts that the upstream IdP
		// already performed a second factor. Local methods (password,
		// TOTP, WebAuthn, lookup, recovery) still need to be paired
		// with an AAL1 entry to reach AAL2.
		for _, tc := range []struct {
			d        string
			methods  []session.AuthenticationMethod
			expected identity.AuthenticatorAssuranceLevel
		}{
			{
				d: "single OIDC method with explicit AAL2 is AAL2",
				methods: []session.AuthenticationMethod{
					{Method: identity.CredentialsTypeOIDC, AAL: identity.AuthenticatorAssuranceLevel2},
				},
				expected: identity.AuthenticatorAssuranceLevel2,
			},
			{
				d: "single SAML method with explicit AAL2 is AAL2",
				methods: []session.AuthenticationMethod{
					{Method: identity.CredentialsTypeSAML, AAL: identity.AuthenticatorAssuranceLevel2},
				},
				expected: identity.AuthenticatorAssuranceLevel2,
			},
			{
				d: "single password method tagged AAL2 stays AAL1",
				methods: []session.AuthenticationMethod{
					{Method: identity.CredentialsTypePassword, AAL: identity.AuthenticatorAssuranceLevel2},
				},
				expected: identity.AuthenticatorAssuranceLevel1,
			},
			{
				d: "single WebAuthn method tagged AAL2 stays AAL1",
				methods: []session.AuthenticationMethod{
					{Method: identity.CredentialsTypeWebAuthn, AAL: identity.AuthenticatorAssuranceLevel2},
				},
				expected: identity.AuthenticatorAssuranceLevel1,
			},
			{
				d: "OIDC AAL2 carry-over plus extra password AAL1 stays AAL2",
				methods: []session.AuthenticationMethod{
					{Method: identity.CredentialsTypePassword, AAL: identity.AuthenticatorAssuranceLevel1},
					{Method: identity.CredentialsTypeOIDC, AAL: identity.AuthenticatorAssuranceLevel2},
				},
				expected: identity.AuthenticatorAssuranceLevel2,
			},
		} {
			t.Run("case="+tc.d, func(t *testing.T) {
				s := session.NewInactiveSession()
				for _, m := range tc.methods {
					s.CompletedLoginFor(m.Method, m.AAL)
				}
				s.SetAuthenticatorAssuranceLevel()
				assert.Equal(t, tc.expected, s.AuthenticatorAssuranceLevel)
			})
		}
	})

	t.Run("case=authentication method upstream acr/amr round-trip", func(t *testing.T) {
		t.Parallel()
		in := session.AuthenticationMethod{
			Method:       identity.CredentialsTypeOIDC,
			AAL:          identity.AuthenticatorAssuranceLevel2,
			CompletedAt:  time.Now().UTC().Round(time.Second),
			Provider:     "hydra",
			Organization: "acme",
			UpstreamACR:  "urn:mfa",
			UpstreamAMR:  []string{"pwd", "mfa"},
		}

		value, err := in.Value()
		require.NoError(t, err)

		var out session.AuthenticationMethod
		require.NoError(t, out.Scan(value))
		assert.Equal(t, in, out)

		// Also round-trip via a list, to exercise the AuthenticationMethods
		// Scan/Value path used by the sessions table column.
		methods := session.AuthenticationMethods{in}
		listValue, err := methods.Value()
		require.NoError(t, err)

		var outList session.AuthenticationMethods
		require.NoError(t, outList.Scan(listValue))
		require.Len(t, outList, 1)
		assert.Equal(t, in, outList[0])
	})

	t.Run("case=CompletedLoginForOIDC populates upstream claims", func(t *testing.T) {
		t.Parallel()
		s := session.NewInactiveSession()
		s.CompletedLoginForOIDC(
			identity.CredentialsTypeOIDC,
			identity.AuthenticatorAssuranceLevel2,
			"hydra",
			"acme",
			"urn:mfa",
			[]string{"pwd", "mfa"},
		)
		require.Len(t, s.AMR, 1)
		assert.Equal(t, identity.CredentialsTypeOIDC, s.AMR[0].Method)
		assert.Equal(t, identity.AuthenticatorAssuranceLevel2, s.AMR[0].AAL)
		assert.Equal(t, "hydra", s.AMR[0].Provider)
		assert.Equal(t, "acme", s.AMR[0].Organization)
		assert.Equal(t, "urn:mfa", s.AMR[0].UpstreamACR)
		assert.Equal(t, []string{"pwd", "mfa"}, s.AMR[0].UpstreamAMR)
		assert.False(t, s.AMR[0].CompletedAt.IsZero())
	})

	t.Run("case=session refresh", func(t *testing.T) {
		req := testhelpers.NewTestHTTPRequest(t, "GET", "/sessions/whoami", nil)

		ctx := contextx.WithConfigValues(t.Context(), map[string]any{
			config.ViperKeySessionLifespan:           "24h",
			config.ViperKeySessionRefreshMinTimeLeft: "12h",
		})
		i := new(identity.Identity)
		i.State = identity.StateActive
		i.NID = x.NewUUID()
		s, err := testhelpers.NewActiveSession(req, reg, i, authAt, identity.CredentialsTypePassword, identity.AuthenticatorAssuranceLevel1)
		require.NoError(t, err)
		assert.False(t, s.CanBeRefreshed(ctx, reg.Config()), "fresh session is not refreshable")

		s.ExpiresAt = s.ExpiresAt.Add(-12 * time.Hour)
		assert.True(t, s.CanBeRefreshed(ctx, reg.Config()), "session is refreshable after 12hrs")
	})
}
