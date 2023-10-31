// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package node

const (
	TOTPCode      = "totp_code"
	TOTPSecretKey = "totp_secret_key"
	TOTPQR        = "totp_qr"
	TOTPUnlink    = "totp_unlink"
)

const (
	LookupReveal     = "lookup_secret_reveal"
	LookupRegenerate = "lookup_secret_regenerate"
	LookupDisable    = "lookup_secret_disable"
	LookupCodes      = "lookup_secret_codes"
	LookupConfirm    = "lookup_secret_confirm"
	LookupCodeEnter  = "lookup_secret"
)

const (
	WebAuthnRegisterTrigger     = "webauthn_register_trigger"
	WebAuthnRegister            = "webauthn_register"
	WebAuthnLogin               = "webauthn_login"
	WebAuthnLoginTrigger        = "webauthn_login_trigger"
	WebAuthnRegisterDisplayName = "webauthn_register_displayname"
	WebAuthnRemove              = "webauthn_remove"
	WebAuthnScript              = "webauthn_script"
)
