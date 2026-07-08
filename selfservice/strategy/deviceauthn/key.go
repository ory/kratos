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
	// KeyStateLocked marks a PIN key disabled after too many wrong-PIN attempts.
	KeyStateLocked KeyState = "locked"
)

// The schema name stays UpperCamelCase on purpose: it is part of the published
// DeviceAuthn schema family (DeviceAuthnKey, DeviceType, KeyState, PINConfig),
// which all use UpperCamelCase. Renaming it to the lowerCamelCase the core
// naming convention prefers would be a breaking change to the public SDK and
// would split the family's casing, so the deliberate exception is kept.
// The valid values (pin, platform, none) are pinned as an enum via the OpenAPI
// patch in .schema/openapi/patches/schema.yaml. (Detached from the doc comment
// to stay out of the generated API description.)

// UserVerification records how the key's holder is verified at use time.
//
// The empty value marks a legacy key from before user verification existed;
// such keys are rejected at login and must be re-enrolled.
type UserVerification string

const (
	UserVerificationPIN      UserVerification = "pin"
	UserVerificationPlatform UserVerification = "platform"
	// UserVerificationNone marks a possession-only key: usable as a second
	// factor (step-up) but never as a first factor, which login enforces.
	// Unlike JWT's alg=none the value strictly reduces what the key may do —
	// it bypasses no verification. It covers keys deliberately enrolled
	// without PIN or platform verification.
	UserVerificationNone UserVerification = "none"
)

// PINConfig holds only scalar fields, so it is safe to shallow copy; the pointer
// on Key.PIN encodes optionality (nil means no PIN), not copy safety.
//
// The schema name stays UpperCamelCase: it belongs to the published DeviceAuthn
// schema family (DeviceAuthnKey, DeviceType, KeyState, UserVerification), and
// renaming it would break the public SDK.
// (Detached from the doc comment to stay out of the generated API description.)

// PINConfig is the per-key PIN state. The pin_secret field holds the at-rest
// ciphertext; the plaintext exists only transiently in memory during
// verification and is cleared once the key locks.
type PINConfig struct {
	// pin_secret holds ciphertext of the kratos cipher (hex) and must never be
	// returned to or logged by clients; the OpenAPI patch in
	// .schema/openapi/patches/schema.yaml marks it write-only.
	// (Detached from the doc comment to stay out of the API description.)

	// PINSecret is the at-rest pin_secret ciphertext. Server-internal: never
	// logged or transmitted. Empty once the key locks.
	PINSecret string `json:"pin_secret"`
	// A negative stored value must be a decode error instead of silently
	// extending the attempt budget, hence uint. (Detached from the doc comment
	// to stay out of the generated API description.)

	// FailedAttempts counts consecutive wrong-PIN attempts; the key locks when it
	// reaches the configured maximum.
	FailedAttempts uint `json:"failed_attempts"`
	// CreatedAt is when the pin_secret was first issued.
	CreatedAt time.Time `json:"created_at"`
	// RotatedAt is when the pin_secret was last rotated; the zero value means
	// never rotated. omitzero (not omitempty) drops the zero timestamp from the
	// JSON, since omitempty never treats a time.Time value as empty.
	RotatedAt time.Time `json:"rotated_at,omitzero"`
}

// Key is safe to shallow copy as long as the shared reference fields (PublicKey,
// PIN, Attestation, RelaxedAttestationExpiresAt) are treated as read-only:
// in-place mutation goes through a *Key into the backing slice, and the merge
// copies PIN rather than aliasing it. Kept out of the doc comment so it stays
// out of the generated API description.

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

	// UserVerification records how the key's holder is verified at use time.
	// Empty for v1 (legacy step-up-only) keys.
	UserVerification UserVerification `json:"user_verification,omitempty"`
	// PIN holds the per-key PIN state. Nil for v1 (legacy step-up-only) keys.
	PIN *PINConfig `json:"pin,omitempty"`

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
