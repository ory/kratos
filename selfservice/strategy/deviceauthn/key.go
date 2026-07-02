// Copyright © 2026 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package deviceauthn

import (
	"time"
)

type DeviceType string

const (
	DeviceTypeAndroid DeviceType = "Android"
	DeviceTypeIOS     DeviceType = "iOS"
)

type KeyState string

const (
	KeyStateInitial   KeyState = "initial"
	KeyStateConfirmed KeyState = "confirmed"
)

// Key represents an enrolled hardware security key used for device authentication (DeviceAuthn).
//
// swagger:model DeviceAuthnKey
type Key struct {
	// v1 uses SHA256 + EC256.
	// v2 (in the future) may use ML-DSA which is post-quantum resistant.
	// This requires Android/iOS support so we have to wait.
	// We intentionally avoid storing the cryptographic algorithm here a la JWT/TLS
	// to avoid security issues and algorithm negotiation.
	Version int `json:"version"`

	// DeviceName is a human readable name for the device, helping the user to distinguish it from others.
	DeviceName string `json:"device_name"`
	// PublicKey is the device's public key (EC P-256 in v1), used to verify
	// signatures. It is stored in PKIX, ASN.1 DER form (the SubjectPublicKeyInfo
	// encoding produced by x509.MarshalPKIXPublicKey). The private key resides
	// inside the device and does not exist on the server.
	PublicKey []byte `json:"public_key,omitempty"`

	// ClientKeyID is the key's stable, unique-per-identity id, computed as
	// hex(SHA-256(PublicKey)) — the lowercase-hex SHA-256 digest of the public
	// key's PKIX, ASN.1 DER (SubjectPublicKeyInfo) encoding, exactly the bytes
	// stored in PublicKey. It is a deterministic fingerprint of the enrolled
	// signing key, so the device recomputes the identical value locally instead
	// of receiving it from the server. Keys enrolled before the server derived
	// the id keep their original client-chosen value.
	ClientKeyID string `json:"client_key_id"`
	// CreatedAt is the timestamp of when the key was created.
	// Only used for troubleshooting/UI.
	CreatedAt time.Time `json:"created_at"`

	// Android/iOS. We have to track that because these two platforms have different encoding formats.
	DeviceType DeviceType `json:"device_type"`

	// Only a confirmed key can be use to step-up.
	// Depending on the configuration (TODO), the key is auto-confirmed at creation time,
	// or it must be set to 'confirmed' with an admin API call.
	// This is to support out-of-band KYC processes e.g. sending a physical letter with a code to the device owner.
	State KeyState `json:"state"`

	// Attestation holds the platform-specific attestation captured at
	// enrollment. Useful for distinguishing keys in the UI (e.g. OS
	// version, manufacturer) and for forensics.
	Attestation *Attestation `json:"attestation,omitempty"`

	// RelaxedAttestationExpiresAt is set only when the key's attestation
	// chain validated because relaxed attestation was allowed (software
	// roots, expired certs, software security level) rather than under
	// strict rules. Such keys are second-class: they are refused at login
	// after this time, or immediately if relaxed attestation is turned off.
	// It is nil for hardware-attested keys that pass strict validation.
	RelaxedAttestationExpiresAt *time.Time `json:"relaxed_attestation_expires_at,omitempty"`
}
