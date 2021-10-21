package webauthn

import (
	"time"

	"github.com/duo-labs/webauthn/webauthn"
)

// CredentialsConfig is the struct that is being used as part of the identity credentials.
type CredentialsConfig struct {
	// List of webauthn credentials.
	Credentials Credentials `json:"credentials"`
}

type Credentials []Credential

func CredentialFromWebAuthn(credential *webauthn.Credential) *Credential {
	return &Credential{
		ID:              credential.ID,
		PublicKey:       credential.PublicKey,
		AttestationType: credential.AttestationType,
		Authenticator: Authenticator{
			AAGUID:       credential.Authenticator.AAGUID,
			SignCount:    credential.Authenticator.SignCount,
			CloneWarning: credential.Authenticator.CloneWarning,
		},
	}
}

func (c Credentials) ToWebAuthn() []webauthn.Credential {
	result := make([]webauthn.Credential, len(c))
	for k := range c {
		result[k] = *c[k].ToWebAuthn()
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
}

type Authenticator struct {
	AAGUID       []byte `json:"aaguid"`
	SignCount    uint32 `json:"sign_count"`
	CloneWarning bool   `json:"clone_warning"`
}
