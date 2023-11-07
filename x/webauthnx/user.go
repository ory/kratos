// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package webauthnx

import (
	"github.com/go-webauthn/webauthn/webauthn"

	"github.com/ory/x/stringsx"
)

var _ webauthn.User = (*User)(nil)

type User struct {
	Name        string
	ID          []byte
	Credentials []webauthn.Credential
	Config      *webauthn.Config
}

func NewUser(id []byte, credentials []webauthn.Credential, config *webauthn.Config) *User {
	return &User{
		ID:          id,
		Credentials: credentials,
		Config:      config,
	}
}

func (u *User) WebAuthnID() []byte {
	return u.ID
}

func (u *User) WebAuthnName() string {
	return stringsx.Coalesce(u.Name, u.Config.RPDisplayName)
}

func (u *User) WebAuthnDisplayName() string {
	return stringsx.Coalesce(u.Name, u.Config.RPDisplayName)
}

func (u *User) WebAuthnIcon() string {
	return "" // Icon option has been removed due to security considerations.
}

func (u *User) WebAuthnCredentials() []webauthn.Credential {
	return u.Credentials
}
