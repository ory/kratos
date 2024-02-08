// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package identity

import (
	"time"

	"github.com/go-webauthn/webauthn/webauthn"

	"github.com/ory/kratos/x/webauthnx/aaguid"
)

// CredentialsWebAuthnConfig is the struct that is being used as part of the identity credentials.
type CredentialsWebAuthnConfig struct {
	// List of webauthn credentials.
	Credentials CredentialsWebAuthn `json:"credentials"`
	UserHandle  []byte              `json:"user_handle"`
}

type CredentialsWebAuthn []CredentialWebAuthn

func CredentialFromWebAuthn(credential *webauthn.Credential, isPasswordless bool) *CredentialWebAuthn {
	cred := &CredentialWebAuthn{
		ID:              credential.ID,
		PublicKey:       credential.PublicKey,
		IsPasswordless:  isPasswordless,
		AttestationType: credential.AttestationType,
		AddedAt:         time.Now().UTC().Round(time.Second),
		Authenticator: AuthenticatorWebAuthn{
			AAGUID:       credential.Authenticator.AAGUID,
			SignCount:    credential.Authenticator.SignCount,
			CloneWarning: credential.Authenticator.CloneWarning,
		},
	}
	id := aaguid.Lookup(credential.Authenticator.AAGUID)
	if id != nil {
		cred.DisplayName = id.Name
	}

	return cred
}

func (c CredentialsWebAuthn) ToWebAuthn() (result []webauthn.Credential) {
	for k := range c {
		result = append(result, *c[k].ToWebAuthn())
	}
	return result
}

// PasswordlessOnly returns only passwordless credentials.
func (c CredentialsWebAuthn) PasswordlessOnly() (result []webauthn.Credential) {
	for k, cc := range c {
		if cc.IsPasswordless {
			result = append(result, *c[k].ToWebAuthn())
		}
	}
	return result
}

// ToWebAuthnFiltered returns only the appropriate credentials for the requested
// AAL. For AAL1, only passwordless credentials are returned, for AAL2, only
// non-passwordless credentials are returned.
func (c CredentialsWebAuthn) ToWebAuthnFiltered(aal AuthenticatorAssuranceLevel) (result []webauthn.Credential) {
	for k, cc := range c {
		if (aal == AuthenticatorAssuranceLevel1 && cc.IsPasswordless) ||
			(aal == AuthenticatorAssuranceLevel2 && !cc.IsPasswordless) {
			result = append(result, *c[k].ToWebAuthn())
		}

	}
	return result
}

func (c *CredentialWebAuthn) ToWebAuthn() *webauthn.Credential {
	return &webauthn.Credential{
		ID:              c.ID,
		PublicKey:       c.PublicKey,
		AttestationType: c.AttestationType,
		Authenticator: webauthn.Authenticator{
			AAGUID:       c.Authenticator.AAGUID,
			SignCount:    c.Authenticator.SignCount,
			CloneWarning: c.Authenticator.CloneWarning,
		},
	}
}

type CredentialWebAuthn struct {
	ID              []byte                `json:"id"`
	PublicKey       []byte                `json:"public_key"`
	AttestationType string                `json:"attestation_type"`
	Authenticator   AuthenticatorWebAuthn `json:"authenticator"`
	DisplayName     string                `json:"display_name"`
	AddedAt         time.Time             `json:"added_at"`
	IsPasswordless  bool                  `json:"is_passwordless"`
}

type AuthenticatorWebAuthn struct {
	AAGUID       []byte `json:"aaguid"`
	SignCount    uint32 `json:"sign_count"`
	CloneWarning bool   `json:"clone_warning"`
}
