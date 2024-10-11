// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package identity

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
)

func TestCredentialConversion(t *testing.T) {
	expected := &webauthn.Credential{
		ID:              []byte("abcdef"),
		PublicKey:       []byte("foobar"),
		AttestationType: "test",
		Authenticator: webauthn.Authenticator{
			AAGUID:       []byte("baz"),
			SignCount:    1,
			CloneWarning: false,
		},
	}

	actual := CredentialFromWebAuthn(expected, false).ToWebAuthn()
	assert.Equal(t, expected, actual)

	actualList := CredentialsWebAuthn{*CredentialFromWebAuthn(expected, false)}.ToWebAuthnFiltered(AuthenticatorAssuranceLevel2, nil)
	assert.Equal(t, []webauthn.Credential{*expected}, actualList)

	actualList = CredentialsWebAuthn{*CredentialFromWebAuthn(expected, true)}.ToWebAuthnFiltered(AuthenticatorAssuranceLevel1, nil)
	assert.Equal(t, []webauthn.Credential{*expected}, actualList)

	actualList = CredentialsWebAuthn{*CredentialFromWebAuthn(expected, true)}.ToWebAuthnFiltered(AuthenticatorAssuranceLevel2, nil)
	assert.Len(t, actualList, 0)

	actualList = CredentialsWebAuthn{*CredentialFromWebAuthn(expected, false)}.ToWebAuthnFiltered(AuthenticatorAssuranceLevel1, nil)
	assert.Len(t, actualList, 0)

	fromWebAuthn := CredentialFromWebAuthn(expected, true)
	assert.True(t, fromWebAuthn.IsPasswordless)
	fromWebAuthn = CredentialFromWebAuthn(expected, false)
	assert.False(t, fromWebAuthn.IsPasswordless)

	expected.Authenticator.AAGUID = uuid.Must(uuid.FromString("ea9b8d66-4d01-1d21-3ce4-b6b48cb575d4")).Bytes()
	fromWebAuthn = CredentialFromWebAuthn(expected, false)
	assert.Equal(t, "Google Password Manager", fromWebAuthn.DisplayName)
}

func TestPasswordlessOnly(t *testing.T) {
	a := *CredentialFromWebAuthn(&webauthn.Credential{ID: []byte("a")}, false)
	b := *CredentialFromWebAuthn(&webauthn.Credential{ID: []byte("b")}, false)
	c := *CredentialFromWebAuthn(&webauthn.Credential{ID: []byte("c")}, true)
	d := *CredentialFromWebAuthn(&webauthn.Credential{ID: []byte("d")}, false)
	e := *CredentialFromWebAuthn(&webauthn.Credential{ID: []byte("e")}, true)
	expected := CredentialsWebAuthn{a, b, c, d, e}

	actual := expected.PasswordlessOnly(nil)
	require.Len(t, actual, 2)
	assert.Equal(t, []webauthn.Credential{*c.ToWebAuthn(), *e.ToWebAuthn()}, actual)
}
