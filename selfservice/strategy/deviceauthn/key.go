// Copyright © 2026 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package deviceauthn

import (
	"time"
)

// The valid values (Android, iOS) are pinned as an enum via the OpenAPI patch
// in .schema/openapi/patches/schema.yaml. (Detached from the doc comment to
// stay out of the generated API description.)

// DeviceType is the platform the key was enrolled on.
//
// "Android" and "iOS" use different attestation and signature encodings.
type DeviceType string

const (
	DeviceTypeAndroid DeviceType = "Android"
	DeviceTypeIOS     DeviceType = "iOS"
)

// The valid values (initial, confirmed, locked) are pinned as an enum via the
// OpenAPI patch in .schema/openapi/patches/schema.yaml. (Detached from the doc
// comment to stay out of the generated API description.)

// KeyState is the lifecycle state of an enrolled key.
//
// Only a "confirmed" key can be used to authenticate. A "locked" key was
// disabled after too many wrong-PIN attempts; rotating its pin_secret unlocks
// it.
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
// "pin" means the holder proves knowledge of a PIN via pin_proof at login;
// "platform" means the device gates key use behind platform biometrics or an
// equivalent lock screen; "none" marks a possession-only key, usable as a
// second factor but never as the sole first factor. The empty value marks a
// legacy key from before user verification existed; such keys are rejected at
// login and must be re-enrolled.
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

// PINConfig is the per-key PIN state.
//
// The pin_secret field holds only the at-rest ciphertext; the plaintext
// exists transiently in server memory during verification and is invalidated
// once the key locks.
type PINConfig struct {
	// pin_secret holds ciphertext of the kratos cipher (hex) and must never be
	// returned to or logged by clients; the OpenAPI patch in
	// .schema/openapi/patches/schema.yaml marks it write-only.
	// (Detached from the doc comment to stay out of the API description.)

	// The at-rest ciphertext of the pin_secret. It never leaves the server and
	// is cleared once the key locks.
	PINSecret string `json:"pin_secret"`
	// A negative stored value must be a decode error instead of silently
	// extending the attempt budget, hence uint. (Detached from the doc comment
	// to stay out of the generated API description.)

	// The number of consecutive wrong-PIN attempts so far; the key locks when
	// it reaches the configured maximum (pin_max_attempts, default 5).
	FailedAttempts uint `json:"failed_attempts"`
	// When the pin_secret was first issued.
	CreatedAt time.Time `json:"created_at"`

	// When the pin_secret was last rotated. Omitted if the secret was never
	// rotated.
	RotatedAt time.Time `json:"rotated_at,omitzero"`
}

// Key is safe to shallow copy as long as the shared reference fields (PublicKey,
// PIN, Attestation, RelaxedAttestationExpiresAt) are treated as read-only:
// in-place mutation goes through a *Key into the backing slice, and the merge
// copies PIN rather than aliasing it. Kept out of the doc comment so it stays
// out of the generated API description.

// An enrolled device authentication (DeviceAuthn) key.
//
// Represents a hardware-backed signing key enrolled from a mobile device.
// The private key resides inside the device and never exists on the server.
//
// To list the identity's enrolled keys, fetch a settings flow: each key's
// remove button (a `ui.nodes` entry named `deviceauthn_remove` in group
// `deviceauthn`) carries the key, with its PIN state redacted, in the node
// label's `context`.
//
// swagger:model DeviceAuthnKey
type Key struct {
	// We intentionally avoid storing the cryptographic algorithm a la JWT/TLS to
	// avoid security issues and algorithm negotiation; a future v2 may use
	// ML-DSA (post-quantum) once Android/iOS support it. (Detached from the doc
	// comment to stay out of the generated API description.)

	// The cryptography version of the key. Version 1 uses ECDSA with SHA-256
	// on an elliptic curve (P-224, P-256, P-384, or P-521); further versions
	// are reserved for future signature suites.
	Version int `json:"version"`

	// A human-readable name for the device, helping the user tell this key
	// apart from others.
	DeviceName string `json:"device_name"`
	// The device's public key (an elliptic-curve key on P-224, P-256, P-384,
	// or P-521 in version 1) in PKIX, ASN.1 DER (SubjectPublicKeyInfo) form,
	// base64-encoded. Signatures are verified against this key.
	//
	// swagger:strfmt byte
	PublicKey []byte `json:"public_key,omitempty"`

	// The key's stable id, unique per identity. Submit it as the
	// `client_key_id` when logging in with the key, deleting it, or rotating
	// its pin_secret.
	//
	// The device can also compute the id without reading it back from the
	// server: it is the lowercase-hex SHA-256 of `public_key` (the key's
	// PKIX, ASN.1 DER encoding). Keys enrolled before the server derived the
	// id keep their original client-chosen value, so prefer reading this
	// field over recomputing it for older keys.
	ClientKeyID string `json:"client_key_id"`
	// When the key was enrolled. Only used for troubleshooting and UI.
	CreatedAt time.Time `json:"created_at"`

	// The platform the key was enrolled on, "Android" or "iOS". The two
	// platforms use different attestation and signature encodings.
	DeviceType DeviceType `json:"device_type"`

	// Auto-confirmation vs. an admin confirmation call (for out-of-band KYC,
	// e.g. a physical letter with a code) is not configurable yet; keys are
	// auto-confirmed at creation time. (Detached from the doc comment to stay
	// out of the generated API description.)

	// The lifecycle state of the key. Only a "confirmed" key can be used to
	// authenticate; a "locked" key was disabled after too many wrong-PIN
	// attempts.
	State KeyState `json:"state"`

	// How the key's holder is verified at use time: "pin", "platform", or
	// "none". Empty for legacy keys enrolled before user verification existed;
	// such keys are rejected at login and must be re-enrolled.
	UserVerification UserVerification `json:"user_verification,omitempty"`
	// The per-key PIN state. Absent unless the key is PIN-protected.
	PIN *PINConfig `json:"pin,omitempty"`

	// The platform-specific attestation captured at enrollment. Useful for
	// telling keys apart in the UI (e.g. OS version, manufacturer) and for
	// forensics.
	Attestation *Attestation `json:"attestation,omitempty"`

	// Set only when the key's attestation chain was accepted under relaxed
	// rules (software roots, expired certificates, software security level)
	// rather than strict hardware attestation. Such keys are refused at login
	// after this time, or immediately once relaxed attestation is turned off.
	// Absent for hardware-attested keys that pass strict validation.
	RelaxedAttestationExpiresAt *time.Time `json:"relaxed_attestation_expires_at,omitempty"`
}

// CredentialsDeviceAuthnConfig is the JSON shape stored in
// `identity_credentials.config` for the `deviceauthn` credential type.
type CredentialsDeviceAuthnConfig struct {
	Credentials []Key `json:"credentials"`
}
