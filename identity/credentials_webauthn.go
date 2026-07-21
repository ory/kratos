// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package identity

import (
	"bytes"
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

// swagger:model identityCredentialsWebAuthnList
type CredentialsWebAuthn []CredentialWebAuthn

func CredentialFromWebAuthn(credential *webauthn.Credential, isPasswordless bool) *CredentialWebAuthn {
	cred := &CredentialWebAuthn{
		ID:              credential.ID,
		PublicKey:       credential.PublicKey,
		IsPasswordless:  isPasswordless,
		AttestationType: credential.AttestationType,
		Transport:       credential.Transport,
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

// UpdateFromLogin persists the mutable authenticator state returned by the
// go-webauthn library after a successful login assertion onto the matching
// stored credential. It implements the relying-party side of the W3C WebAuthn
// storage contract for the fields that change across assertions
// (https://www.w3.org/TR/webauthn-2/#sctn-sign-counter):
//
//   - SignCount: incremented by the library when the assertion reports a higher
//     counter; left unchanged (and CloneWarning set) when it did not advance, a
//     possible cloned authenticator.
//   - CloneWarning: persisted so the flag survives on the credential.
//   - BackupState: refreshed because it flips when the user starts or stops
//     syncing the credential (for example a device-bound passkey later added to
//     iCloud or Google). BackupEligible is immutable by spec and is left alone.
//
// It returns whether the stored config changed so the caller can skip a needless
// write.
//
// Synced passkeys (for example Apple iCloud or Google) always report
// SignCount = 0. For them go-webauthn neither advances the counter nor raises a
// clone warning, so this method is a no-op for the counter and never blocks the
// login.
func (c CredentialsWebAuthn) UpdateFromLogin(updated *webauthn.Credential) (changed bool) {
	if updated == nil {
		return false
	}

	for k := range c {
		if !bytes.Equal(c[k].ID, updated.ID) {
			continue
		}

		if c[k].Authenticator == nil {
			c[k].Authenticator = &AuthenticatorWebAuthn{}
		}

		// Only ever advance the counter, never lower it. go-webauthn already
		// leaves SignCount unchanged when the assertion does not advance it, so
		// in practice updated >= stored; the strict < guards against a regressing
		// counter ever weakening clone detection by writing a lower value.
		if c[k].Authenticator.SignCount < updated.Authenticator.SignCount {
			c[k].Authenticator.SignCount = updated.Authenticator.SignCount
			changed = true
		}
		if c[k].Authenticator.CloneWarning != updated.Authenticator.CloneWarning {
			c[k].Authenticator.CloneWarning = updated.Authenticator.CloneWarning
			changed = true
		}

		// BackupState is the only credential flag that legitimately changes over
		// a credential's lifetime; the others are fixed at registration. Sync it
		// so the stored state stays current for later risk decisions. Legacy
		// credentials with no Flags are skipped on purpose: BackupEligible is
		// unknown for them, so we do not synthesize a Flags struct here; the
		// login filter upgrades those separately before the assertion.
		if c[k].Flags != nil && c[k].Flags.BackupState != updated.Flags.BackupState {
			c[k].Flags.BackupState = updated.Flags.BackupState
			changed = true
		}

		return changed
	}

	return false
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

// CredentialWebAuthn contains information about a webauthn credential.
//
// swagger:model identityCredentialsWebAuthn
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

// CredentialWebAuthnFlags contains information about the flags of a webauthn credential.
//
// swagger:model identityCredentialsWebAuthnFlags
type CredentialWebAuthnFlags struct {
	UserPresent    bool `json:"user_present"`
	UserVerified   bool `json:"user_verified"`
	BackupEligible bool `json:"backup_eligible"`
	BackupState    bool `json:"backup_state"`
}

// CredentialWebAuthnAttestation contains information about attestation of a webauthn credential.
//
// swagger:model identityCredentialsWebAuthnAttestation
type CredentialWebAuthnAttestation struct {
	ClientDataJSON     []byte `json:"client_dataJSON"`
	ClientDataHash     []byte `json:"client_data_hash"`
	AuthenticatorData  []byte `json:"authenticator_data"`
	PublicKeyAlgorithm int64  `json:"public_key_algorithm"`
	Object             []byte `json:"object"`
}

// AuthenticatorWebAuthn contains information about the authenticator of a webauthn credential.
//
// swagger:model identityCredentialsWebAuthnAuthenticator
type AuthenticatorWebAuthn struct {
	AAGUID       []byte `json:"aaguid"`
	SignCount    uint32 `json:"sign_count"`
	CloneWarning bool   `json:"clone_warning"`
}
