// Copyright © 2025 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package webauthn_test

import (
	"encoding/json"
	"net/http"
	"net/url"
	"testing"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	"github.com/ory/x/configx"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/pkg"
	kratos "github.com/ory/kratos/pkg/httpclient"
	"github.com/ory/kratos/pkg/testhelpers"
	"github.com/ory/kratos/ui/node"
)

// TestLoginUpdatesSignCount is a regression test for HackerOne #3780149.
//
// The go-webauthn library returns a *webauthn.Credential from ValidateLogin
// whose Authenticator.SignCount reflects the signature counter reported by the
// authenticator during the assertion. W3C WebAuthn clone detection relies on
// the relying party persisting this incremented counter so that a replayed or
// cloned authenticator (which reports a stale counter) can be detected.
//
// Before the fix, Kratos discarded the returned credential, so the stored
// SignCount stayed at its registration value forever and clone detection was
// inert.
//
// The MFA v1 fixtures store sign_count = 3 and the assertion response reports
// counter = 10, so a correct implementation persists SignCount = 10.
func TestLoginUpdatesSignCount(t *testing.T) {
	t.Parallel()

	conf, reg := pkg.NewFastRegistryWithMocks(t,
		configx.WithValue(config.ViperKeySelfServiceStrategyConfig+"."+string(identity.CredentialsTypePassword)+".enabled", false),
		enabledWebauthn,
		configx.WithValues(testhelpers.DefaultIdentitySchemaConfig("file://./stub/login.schema.json")),
	)

	publicTS, _ := testhelpers.NewKratosServer(t, reg)
	_ = testhelpers.NewErrorTestServer(t, reg)
	_ = testhelpers.NewLoginUIFlowEchoServer(t, reg)
	_ = testhelpers.NewRedirSessionEchoTS(t, reg)

	createIdentityWithWebAuthn := func(t *testing.T, c identity.Credentials) *identity.Identity {
		var id identity.Identity
		require.NoError(t, json.Unmarshal(loginFixtureSuccessIdentity, &id))

		id.SetCredentials(identity.CredentialsTypeWebAuthn, identity.Credentials{
			Identifiers: []string{loginFixtureSuccessEmail},
			Config:      c.Config,
			Type:        identity.CredentialsTypeWebAuthn,
			Version:     c.Version,
		})

		_ = reg.PrivilegedIdentityPool().DeleteIdentity(t.Context(), id.ID)
		require.NoError(t, reg.PrivilegedIdentityPool().CreateIdentity(t.Context(), &id))
		return &id
	}

	submitWebAuthnLogin := func(t *testing.T, isSPA bool, id *identity.Identity, contextFixture []byte, cb func(values url.Values), opts ...testhelpers.InitFlowWithOption) (string, *http.Response, *kratos.LoginFlow) {
		client := testhelpers.NewHTTPClientWithIdentitySessionCookie(t.Context(), t, reg, id)
		f := testhelpers.InitializeLoginFlowViaBrowser(t, client, publicTS, false, isSPA, false, false, opts...)

		interim, err := reg.LoginFlowPersister().GetLoginFlow(t.Context(), uuid.FromStringOrNil(f.Id))
		require.NoError(t, err)
		interim.InternalContext = contextFixture
		require.NoError(t, reg.LoginFlowPersister().UpdateLoginFlow(t.Context(), interim))

		values := testhelpers.SDKFormFieldsToURLValues(f.Ui.Nodes)
		cb(values)
		body, res := testhelpers.LoginMakeRequest(t, false, isSPA, f, client, values.Encode())
		return body, res, f
	}

	// Reads the stored first webauthn credential: the outer row (for id and
	// created_at) and the decoded authenticator state.
	storedCredential := func(t *testing.T, idID uuid.UUID) (*identity.Credentials, *identity.AuthenticatorWebAuthn) {
		stored, err := reg.PrivilegedIdentityPool().GetIdentityConfidential(t.Context(), idID)
		require.NoError(t, err)
		c, ok := stored.GetCredentials(identity.CredentialsTypeWebAuthn)
		require.True(t, ok)
		var conf identity.CredentialsWebAuthnConfig
		require.NoError(t, json.Unmarshal(c.Config, &conf))
		require.NotEmpty(t, conf.Credentials)
		require.NotNil(t, conf.Credentials[0].Authenticator)
		return c, conf.Credentials[0].Authenticator
	}

	conf.MustSet(t.Context(), config.ViperKeySessionWhoAmIAAL, "aal1")

	id := createIdentityWithWebAuthn(t, identity.Credentials{
		Config:  loginFixtureSuccessV1Credentials,
		Version: 1,
	})

	// Sanity: the registered credential starts at the fixture's sign_count = 3.
	credBefore, authBefore := storedCredential(t, id.ID)
	require.EqualValues(t, 3, authBefore.SignCount)

	body, _, _ := submitWebAuthnLogin(t, true, id, loginFixtureSuccessV1Context, func(values url.Values) {
		values.Set("identifier", loginFixtureSuccessEmail)
		values.Set(node.WebAuthnLogin, string(loginFixtureSuccessV1Response))
	}, testhelpers.InitFlowWithAAL(identity.AuthenticatorAssuranceLevel2))

	// The login must succeed.
	require.True(t, gjson.Get(body, "session.active").Bool(), "%s", body)

	// The stored SignCount must be updated to the value reported by the
	// authenticator assertion (10), not left at the registration value (3).
	credAfter, auth := storedCredential(t, id.ID)
	assert.EqualValues(t, 10, auth.SignCount,
		"stored SignCount must be updated from the login assertion for clone detection to work")

	// The assertion counter advanced (3 -> 10), so this is not a clone: the
	// persisted clone warning must round-trip as false rather than being set
	// spuriously on a normal login.
	assert.False(t, auth.CloneWarning,
		"a normally advancing counter must not set the clone warning")

	// The credential row must be updated in place, not deleted and recreated:
	// its id and created_at must be stable across the login.
	assert.Equal(t, credBefore.ID, credAfter.ID,
		"the credential row must be updated in place (stable id), not recreated")
	assert.Equal(t, credBefore.CreatedAt, credAfter.CreatedAt,
		"persisting the sign counter must not reset the credential's created_at")
}
