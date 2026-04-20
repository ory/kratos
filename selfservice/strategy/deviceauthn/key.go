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
	// PublicKey is an EC (in v1) public key, used to verify signatures, stored as uncompressed bytes.
	// The private key resides inside the device and does not exist on the server.
	PublicKey []byte `json:"public_key,omitempty"`

	// ClientKeyID is a client-chosen id for the key and is unique per identity.
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
}
