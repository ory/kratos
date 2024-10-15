// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package identity

import (
	"time"

	"github.com/go-webauthn/webauthn/protocol"

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
		Authenticator: &AuthenticatorWebAuthn{
			AAGUID:       credential.Authenticator.AAGUID,
			SignCount:    credential.Authenticator.SignCount,
			CloneWarning: credential.Authenticator.CloneWarning,
		},
		Flags: &CredentialWebAuthnFlags{
			UserPresent:    credential.Flags.UserPresent,
			UserVerified:   credential.Flags.UserVerified,
			BackupEligible: credential.Flags.BackupEligible,
			BackupState:    credential.Flags.BackupState,
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
func (c CredentialsWebAuthn) PasswordlessOnly(authenticatorResponseFlags *protocol.AuthenticatorFlags) (result []webauthn.Credential) {
	for k, cc := range c {
		// Upgrade path for legacy webauthn credentials. Only possible if we are handling a response from an authenticator.
		if c[k].Flags == nil && authenticatorResponseFlags != nil {
			c[k].Flags = &CredentialWebAuthnFlags{
				BackupEligible: authenticatorResponseFlags.HasBackupEligible(),
				BackupState:    authenticatorResponseFlags.HasBackupState(),
			}
		}

		if cc.IsPasswordless {
			result = append(result, *c[k].ToWebAuthn())
		}
	}
	return result
}

// ToWebAuthnFiltered returns only the appropriate credentials for the requested
// AAL. For AAL1, only passwordless credentials are returned, for AAL2, only
// non-passwordless credentials are returned.
//
// authenticatorResponseFlags should be passed  if the response is from an authenticator. It will be used to
// upgrade legacy webauthn credentials' BackupEligible and BackupState flags.
func (c CredentialsWebAuthn) ToWebAuthnFiltered(aal AuthenticatorAssuranceLevel, authenticatorResponseFlags *protocol.AuthenticatorFlags) (result []webauthn.Credential) {
	for k, cc := range c {
		// Upgrade path for legacy webauthn credentials. Only possible if we are handling a response from an authenticator.
		if c[k].Flags == nil && authenticatorResponseFlags != nil {
			c[k].Flags = &CredentialWebAuthnFlags{
				BackupEligible: authenticatorResponseFlags.HasBackupEligible(),
				BackupState:    authenticatorResponseFlags.HasBackupState(),
			}
		}

		if (aal == AuthenticatorAssuranceLevel1 && cc.IsPasswordless) ||
			(aal == AuthenticatorAssuranceLevel2 && !cc.IsPasswordless) {
			result = append(result, *c[k].ToWebAuthn())
		}
	}
	return result
}

func (c *CredentialWebAuthn) ToWebAuthn() *webauthn.Credential {
	wc := &webauthn.Credential{
		ID:              c.ID,
		PublicKey:       c.PublicKey,
		AttestationType: c.AttestationType,
		Transport:       c.Transport,
	}

	if c.Authenticator != nil {
		wc.Authenticator = webauthn.Authenticator{
			AAGUID:       c.Authenticator.AAGUID,
			SignCount:    c.Authenticator.SignCount,
			CloneWarning: c.Authenticator.CloneWarning,
		}
	}

	if c.Flags != nil {
		wc.Flags = webauthn.CredentialFlags{
			UserPresent:    c.Flags.UserPresent,
			UserVerified:   c.Flags.UserVerified,
			BackupEligible: c.Flags.BackupEligible,
			BackupState:    c.Flags.BackupState,
		}
	}

	if c.Attestation != nil {
		wc.Attestation = webauthn.CredentialAttestation{
			ClientDataJSON:     c.Attestation.ClientDataJSON,
			ClientDataHash:     c.Attestation.ClientDataHash,
			AuthenticatorData:  c.Attestation.AuthenticatorData,
			PublicKeyAlgorithm: c.Attestation.PublicKeyAlgorithm,
			Object:             c.Attestation.Object,
		}
	}

	return wc
}

type CredentialWebAuthn struct {
	ID              []byte                            `json:"id"`
	PublicKey       []byte                            `json:"public_key"`
	AttestationType string                            `json:"attestation_type"`
	Authenticator   *AuthenticatorWebAuthn            `json:"authenticator,omitempty"`
	DisplayName     string                            `json:"display_name"`
	AddedAt         time.Time                         `json:"added_at"`
	IsPasswordless  bool                              `json:"is_passwordless"`
	Flags           *CredentialWebAuthnFlags          `json:"flags,omitempty"`
	Transport       []protocol.AuthenticatorTransport `json:"transport,omitempty"`
	Attestation     *CredentialWebAuthnAttestation    `json:"attestation,omitempty"`
}

type CredentialWebAuthnFlags struct {
	UserPresent    bool `json:"user_present"`
	UserVerified   bool `json:"user_verified"`
	BackupEligible bool `json:"backup_eligible"`
	BackupState    bool `json:"backup_state"`
}

type CredentialWebAuthnAttestation struct {
	ClientDataJSON     []byte `json:"client_dataJSON"`
	ClientDataHash     []byte `json:"client_data_hash"`
	AuthenticatorData  []byte `json:"authenticator_data"`
	PublicKeyAlgorithm int64  `json:"public_key_algorithm"`
	Object             []byte `json:"object"`
}

type AuthenticatorWebAuthn struct {
	AAGUID       []byte `json:"aaguid"`
	SignCount    uint32 `json:"sign_count"`
	CloneWarning bool   `json:"clone_warning"`
}
