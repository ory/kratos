package webauthn

import (
	"time"

	"github.com/ory/kratos/identity"

	"github.com/duo-labs/webauthn/webauthn"
)

// CredentialsConfig is the struct that is being used as part of the identity credentials.
type CredentialsConfig struct {
	// List of webauthn credentials.
	Credentials Credentials `json:"credentials"`
	UserHandle  []byte      `json:"user_handle"`
}

type Credentials []Credential

func CredentialFromWebAuthn(credential *webauthn.Credential, isPasswordless bool) *Credential {
	return &Credential{
		ID:              credential.ID,
		PublicKey:       credential.PublicKey,
		IsPasswordless:  isPasswordless,
		AttestationType: credential.AttestationType,
		Authenticator: Authenticator{
			AAGUID:       credential.Authenticator.AAGUID,
			SignCount:    credential.Authenticator.SignCount,
			CloneWarning: credential.Authenticator.CloneWarning,
		},
	}
}

func (c Credentials) ToWebAuthn() (result []webauthn.Credential) {
	for k := range c {
		result = append(result, *c[k].ToWebAuthn())
	}
	return result
}

func (c Credentials) ToWebAuthnFiltered(aal identity.AuthenticatorAssuranceLevel) (result []webauthn.Credential) {
	for k, cc := range c {
		if aal == identity.AuthenticatorAssuranceLevel1 && !cc.IsPasswordless {
			continue
		} else if aal == identity.AuthenticatorAssuranceLevel2 && cc.IsPasswordless {
			continue
		}

		result = append(result, *c[k].ToWebAuthn())
	}
	return result
}

func (c *Credential) ToWebAuthn() *webauthn.Credential {
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

type Credential struct {
	ID              []byte        `json:"id"`
	PublicKey       []byte        `json:"public_key"`
	AttestationType string        `json:"attestation_type"`
	Authenticator   Authenticator `json:"authenticator"`
	DisplayName     string        `json:"display_name"`
	AddedAt         time.Time     `json:"added_at"`
	IsPasswordless  bool          `json:"is_passwordless"`
}

type Authenticator struct {
	AAGUID       []byte `json:"aaguid"`
	SignCount    uint32 `json:"sign_count"`
	CloneWarning bool   `json:"clone_warning"`
}
