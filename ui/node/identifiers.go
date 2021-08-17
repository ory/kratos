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
	LookupCodes      = "lookup_secret_codes"
	LookupConfirm    = "lookup_secret_confirm"
	LookupCodeEnter  = "lookup_secret"
)

const (
	WebAuthnRegister            = "webauthn_register"
	WebAuthnIdentifier          = "webauthn_identifier"
	WebAuthnLogin               = "webauthn_login"
	WebAuthnRegisterDisplayName = "webauthn_register_displayname"
	WebAuthnRemove              = "webauthn_remove"
)
