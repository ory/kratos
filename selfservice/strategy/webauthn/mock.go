package webauthn

import (
	"github.com/duo-labs/webauthn/webauthn"
	"github.com/gofrs/uuid"
)

type wrappedUser struct {
	id uuid.UUID
	c  []webauthn.Credential
}

var _ webauthn.User = (*wrappedUser)(nil)

func (user *wrappedUser) WebAuthnID() []byte {
	return user.id[:]
}

func (user *wrappedUser) WebAuthnName() string {
	return "placeholder"
}

func (user *wrappedUser) WebAuthnDisplayName() string {
	return "placeholder"
}

func (user *wrappedUser) WebAuthnIcon() string {
	return "https://via.placeholder.com/128"
}

func (user *wrappedUser) WebAuthnCredentials() []webauthn.Credential {
	return user.c
}
