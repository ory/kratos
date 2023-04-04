// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package identity

import (
	"time"

	"github.com/duo-labs/webauthn/webauthn"
)

// CredentialsConfig is the struct that is being used as part of the identity credentials.
type CredentialsWebAuthnConfig struct {
	// List of webauthn credentials.
	Credentials CredentialsWebAuthn `json:"credentials"`
	UserHandle  []byte              `json:"user_handle"`
}

type CredentialsWebAuthn []CredentialWebAuthn

func CredentialFromWebAuthn(credential *webauthn.Credential, isPasswordless bool) *CredentialWebAuthn {
	return &CredentialWebAuthn{
		ID:              credential.ID,
		PublicKey:       credential.PublicKey,
		IsPasswordless:  isPasswordless,
		AttestationType: credential.AttestationType,
		Authenticator: AuthenticatorWebAuthn{
			AAGUID:       credential.Authenticator.AAGUID,
			SignCount:    credential.Authenticator.SignCount,
			CloneWarning: credential.Authenticator.CloneWarning,
		},
	}
}

func (c CredentialsWebAuthn) ToWebAuthn() (result []webauthn.Credential) {
	for k := range c {
		result = append(result, *c[k].ToWebAuthn())
	}
	return result
}

func (c CredentialsWebAuthn) ToWebAuthnFiltered(aal AuthenticatorAssuranceLevel) (result []webauthn.Credential) {
	for k, cc := range c {
		if aal == AuthenticatorAssuranceLevel1 && !cc.IsPasswordless {
			continue
		} else if aal == AuthenticatorAssuranceLevel2 && cc.IsPasswordless {
			continue
		}

		result = append(result, *c[k].ToWebAuthn())
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
