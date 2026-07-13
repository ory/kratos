// Copyright © 2026 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package deviceauthn

import (
	"encoding/asn1"
)

// Attestation holds the platform-specific attestation parsed at enrollment.
// Exactly one of `android` or `ios` is set, matching the key's device_type.
//
// swagger:model deviceAuthnAttestation
type Attestation struct {
	Android *AndroidKeyDescription `json:"android,omitempty"`
	IOS     *IOSAttestation        `json:"ios,omitempty"`
}

// AndroidKeyDescription matches the ASN.1 sequence defined by Android.
// It is stored in the leaf certificate's x509 extension.
//
// Schema: https://source.android.com/docs/security/features/keystore/attestation#schema
//
// swagger:model deviceAuthnAndroidKeyDescription
type AndroidKeyDescription struct {
	AttestationVersion       int             `json:"attestation_version,omitempty"`
	AttestationSecurityLevel asn1.Enumerated `json:"attestation_security_level,omitempty"`
	KeymasterVersion         int             `json:"keymaster_version,omitempty"`
	KeymasterSecurityLevel   asn1.Enumerated `json:"keymaster_security_level,omitempty"`
	// AttestationChallenge is verified at enrollment time and not re-used
	// later, so we don't persist it.
	AttestationChallenge []byte `json:"-"`
	// UniqueID is a hardware-tied identifier; we don't persist it for
	// privacy reasons.
	UniqueID         []byte                   `json:"-"`
	SoftwareEnforced AndroidAuthorizationList `json:"software_enforced"`
	TeeEnforced      AndroidAuthorizationList `json:"tee_enforced"`
}

// AndroidAuthorizationList maps the ASN.1 tags to the fields defined in KeyMint 4.
// Schema: https://source.android.com/docs/security/features/keystore/attestation#schema
//
// swagger:model deviceAuthnAndroidAuthorizationList
type AndroidAuthorizationList struct {
	Purpose                     []int              `asn1:"tag:1,explicit,set,optional" json:"purpose,omitempty"`
	Algorithm                   int                `asn1:"tag:2,explicit,optional" json:"algorithm,omitempty"`
	KeySize                     int                `asn1:"tag:3,explicit,optional" json:"key_size,omitempty"`
	Digest                      []int              `asn1:"tag:5,explicit,set,optional" json:"digest,omitempty"`
	EcCurve                     int                `asn1:"tag:10,explicit,optional" json:"ec_curve,omitempty"`
	ActiveDateTime              int                `asn1:"tag:400,explicit,optional" json:"active_date_time,omitempty"`
	OriginationExpireDateTime   int                `asn1:"tag:401,explicit,optional" json:"origination_expire_date_time,omitempty"`
	UsageExpireDateTime         int                `asn1:"tag:402,explicit,optional" json:"usage_expire_date_time,omitempty"`
	NoAuthRequired              asn1.RawValue      `asn1:"tag:503,explicit,optional" json:"-"`
	UserAuthType                int                `asn1:"tag:504,explicit,optional" json:"user_auth_type,omitempty"`
	AuthTimeout                 int                `asn1:"tag:505,explicit,optional" json:"auth_timeout,omitempty"`
	AllowWhileOnBody            asn1.RawValue      `asn1:"tag:506,explicit,optional" json:"-"`
	TrustedUserPresenceRequired asn1.RawValue      `asn1:"tag:507,explicit,optional" json:"-"`
	TrustedConfirmationRequired asn1.RawValue      `asn1:"tag:508,explicit,optional" json:"-"`
	UnlockedDeviceRequired      asn1.RawValue      `asn1:"tag:509,explicit,optional" json:"-"`
	ApplicationID               []byte             `asn1:"tag:601,explicit,optional" json:"application_id,omitempty"`
	CreationDateTime            int                `asn1:"tag:701,explicit,optional" json:"creation_date_time,omitempty"`
	Origin                      int                `asn1:"tag:702,explicit,optional" json:"origin,omitempty"`
	RootOfTrust                 AndroidRootOfTrust `asn1:"tag:704,explicit,optional" json:"root_of_trust"`
	OsVersion                   int                `asn1:"tag:705,explicit,optional" json:"os_version,omitempty"`
	OsPatchLevel                int                `asn1:"tag:706,explicit,optional" json:"os_patch_level,omitempty"`
	AttestationChallenge        []byte             `asn1:"tag:708,explicit,optional" json:"-"`
	AttestationApplicationID    []byte             `asn1:"tag:709,explicit,optional" json:"attestation_application_id,omitempty"`
	AttestationIDBrand          []byte             `asn1:"tag:710,explicit,optional" json:"attestation_id_brand,omitempty"`
	AttestationIDDevice         []byte             `asn1:"tag:711,explicit,optional" json:"attestation_id_device,omitempty"`
	AttestationIDProduct        []byte             `asn1:"tag:712,explicit,optional" json:"attestation_id_product,omitempty"`
	AttestationIDSerial         []byte             `asn1:"tag:713,explicit,optional" json:"attestation_id_serial,omitempty"`
	AttestationIDImei           []byte             `asn1:"tag:714,explicit,optional" json:"attestation_id_imei,omitempty"`
	AttestationIDMeid           []byte             `asn1:"tag:715,explicit,optional" json:"attestation_id_meid,omitempty"`
	AttestationIDManufacturer   []byte             `asn1:"tag:716,explicit,optional" json:"attestation_id_manufacturer,omitempty"`
	AttestationIDModel          []byte             `asn1:"tag:717,explicit,optional" json:"attestation_id_model,omitempty"`
	VendorPatchLevel            int                `asn1:"tag:718,explicit,optional" json:"vendor_patch_level,omitempty"`
	BootPatchLevel              int                `asn1:"tag:719,explicit,optional" json:"boot_patch_level,omitempty"`
	DeviceUniqueAttestation     asn1.RawValue      `asn1:"tag:720,explicit,optional" json:"-"`
}

// swagger:model deviceAuthnAndroidRootOfTrust
type AndroidRootOfTrust struct {
	VerifiedBootKey []byte `asn1:"optional" json:"verified_boot_key,omitempty"`
	DeviceLocked    bool   `json:"device_locked,omitempty"`
	// Actually it is `AndroidVerifiedBootState` but due to ASN1 quirks we define it so.
	VerifiedBootState asn1.Enumerated `json:"verified_boot_state,omitempty"`
	VerifiedBootHash  []byte          `asn1:"optional" json:"verified_boot_hash,omitempty"`
}

// IOSAttStmt is the attestation statement of an Apple App Attest attestation.
// Defined in https://developer.apple.com/documentation/devicecheck/validating-apps-that-connect-to-your-server#Verify-the-attestation .
//
// swagger:model deviceAuthnIOSAttStmt
type IOSAttStmt struct {
	// Receipt is several KB of proprietary opaque data; not persisted.
	Receipt []byte `cbor:"receipt" json:"-"`
	// X5c is the leaf-first certificate chain encoded as base64(DER), per
	// the JOSE convention defined in RFC 7515 §4.1.6.
	X5c [][]byte `cbor:"x5c" json:"x5c,omitempty"`
}

// IOSAttestation is an Apple App Attest attestation.
// Defined in https://developer.apple.com/documentation/devicecheck/validating-apps-that-connect-to-your-server#Verify-the-attestation .
//
// swagger:model deviceAuthnIOSAttestation
type IOSAttestation struct {
	Fmt      string     `cbor:"fmt" json:"fmt,omitempty"`
	AttStmt  IOSAttStmt `cbor:"attStmt" json:"att_stmt"`
	AuthData []byte     `cbor:"authData" json:"auth_data,omitempty"`
}
