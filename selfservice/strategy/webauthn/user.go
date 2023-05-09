// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package webauthn

import "github.com/duo-labs/webauthn/webauthn"

var _ webauthn.User = (*User)(nil)

type User struct {
	id  []byte
	c   []webauthn.Credential
	cfg *webauthn.Config
}

func NewUser(id []byte, c []webauthn.Credential, cfg *webauthn.Config) *User {
	return &User{
		id:  id,
		c:   c,
		cfg: cfg,
	}
}

func (u *User) WebAuthnID() []byte {
	return u.id
}

func (u *User) WebAuthnName() string {
	return u.cfg.RPDisplayName
}

func (u *User) WebAuthnDisplayName() string {
	return u.cfg.RPDisplayName
}

func (u *User) WebAuthnIcon() string {
	return u.cfg.RPIcon
}

func (u *User) WebAuthnCredentials() []webauthn.Credential {
	return u.c
}
