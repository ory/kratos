package webauthn

import (
	"github.com/ory/kratos/identity"
	"testing"

	"github.com/duo-labs/webauthn/webauthn"
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

	actualList := Credentials{*CredentialFromWebAuthn(expected, false)}.ToWebAuthn(identity.AuthenticatorAssuranceLevel2)
	assert.Equal(t, []webauthn.Credential{*expected}, actualList)

	actualList = Credentials{*CredentialFromWebAuthn(expected, true)}.ToWebAuthn(identity.AuthenticatorAssuranceLevel1)
	assert.Equal(t, []webauthn.Credential{*expected}, actualList)

	actualList = Credentials{*CredentialFromWebAuthn(expected, true)}.ToWebAuthn(identity.AuthenticatorAssuranceLevel2)
	assert.Len(t, actualList, 0)

	actualList = Credentials{*CredentialFromWebAuthn(expected, false)}.ToWebAuthn(identity.AuthenticatorAssuranceLevel1)
	assert.Len(t, actualList, 0)

	fromWebAuthn := CredentialFromWebAuthn(expected, true)
	assert.True(t, fromWebAuthn.IsPasswordless)
	fromWebAuthn = CredentialFromWebAuthn(expected, false)
	assert.False(t, fromWebAuthn.IsPasswordless)
}
