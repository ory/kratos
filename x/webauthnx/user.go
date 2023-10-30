// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package webauthnx

import "github.com/go-webauthn/webauthn/webauthn"

var _ webauthn.User = (*User)(nil)

type User struct {
	id          []byte
	credentials []webauthn.Credential
	config      *webauthn.Config
}

func NewUser(id []byte, credentials []webauthn.Credential, config *webauthn.Config) *User {
	return &User{
		id:          id,
		credentials: credentials,
		config:      config,
	}
}

func (u *User) WebAuthnID() []byte {
	return u.id
}

func (u *User) WebAuthnName() string {
	return u.config.RPDisplayName
}

func (u *User) WebAuthnDisplayName() string {
	return u.config.RPDisplayName
}

func (u *User) WebAuthnIcon() string {
	return "" // Icon option has been removed due to security considerations.
}

func (u *User) WebAuthnCredentials() []webauthn.Credential {
	return u.credentials
}
