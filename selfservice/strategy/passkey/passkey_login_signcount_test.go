// Copyright © 2025 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package passkey_test

import (
	"encoding/json"
	"net/url"
	"testing"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/pkg/testhelpers"
	"github.com/ory/kratos/ui/node"
)

// TestLoginUpdatesSignCount is a regression test for HackerOne #3780149 on the
// passkey strategy (which uses ValidateDiscoverableLogin).
//
// The stored passkey credential starts at sign_count = 3 and the assertion
// response reports counter = 10. After a successful login, the stored
// SignCount must be updated to 10 so that W3C clone detection works.
func TestLoginUpdatesSignCount(t *testing.T) {
	t.Parallel()

	fix := newLoginFixture(t)
	fix.conf.MustSet(t.Context(), config.ViperKeySessionWhoAmIAAL, "aal1")

	storedCredential := func(t *testing.T, idID uuid.UUID) (*identity.Credentials, *identity.AuthenticatorWebAuthn) {
		stored, err := fix.reg.PrivilegedIdentityPool().GetIdentityConfidential(t.Context(), idID)
		require.NoError(t, err)
		c, ok := stored.GetCredentials(identity.CredentialsTypePasskey)
		require.True(t, ok)
		var conf identity.CredentialsWebAuthnConfig
		require.NoError(t, json.Unmarshal(c.Config, &conf))
		require.NotEmpty(t, conf.Credentials)
		require.NotNil(t, conf.Credentials[0].Authenticator)
		return c, conf.Credentials[0].Authenticator
	}

	id := fix.createIdentityWithPasskey(t, identity.Credentials{
		Config:  loginPasswordlessCredentials,
		Version: 1,
	})

	// Sanity: the registered credential starts at the fixture's sign_count = 3.
	credBefore, authBefore := storedCredential(t, id.ID)
	require.EqualValues(t, 3, authBefore.SignCount)

	browserClient := testhelpers.NewClientWithCookies(t)
	body, _, _ := fix.submitWebAuthnLoginWithClient(t, true, loginPasswordlessContext, browserClient, func(values url.Values) {
		values.Set(node.PasskeyLogin, string(loginPasswordlessResponse))
	}, testhelpers.InitFlowWithAAL(identity.AuthenticatorAssuranceLevel1))

	require.True(t, gjson.Get(body, "session.active").Bool(), "%s", body)

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
