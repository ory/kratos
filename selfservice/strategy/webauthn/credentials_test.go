package webauthn

import (
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

	actual := CredentialFromWebAuthn(expected).ToWebAuthn()
	assert.Equal(t, expected, actual)

	actualList := Credentials{*CredentialFromWebAuthn(expected)}.ToWebAuthn()
	assert.Equal(t, []webauthn.Credential{*expected}, actualList)
}
