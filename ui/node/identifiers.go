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
	ProfileChooseCredentials = "profile_choose_credentials"
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

const (
	PasskeyRegisterTrigger  = "passkey_register_trigger"
	PasskeyRegister         = "passkey_register"
	PasskeySettingsRegister = "passkey_settings_register"
	PasskeyCreateData       = "passkey_create_data"
	PasskeyLogin            = "passkey_login"
	PasskeyChallenge        = "passkey_challenge"
	PasskeyLoginTrigger     = "passkey_login_trigger" //#nosec G101 -- Not a credential
	PasskeyRemove           = "passkey_remove"
)
