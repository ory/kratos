package webauthn

// swagger:model submitSelfServiceFlowWithWebAuthnRegistrationMethod
// nolint:deadcode,unused
type submitSelfServiceFlowWithWebAuthnRegistrationMethod struct {
	// Register a WebAuthn Security Key
	//
	// It is expected that the JSON returned by the WebAuthn registration process
	// is included here.
	Register string `json:"webauthn_register"`

	// Name of the WebAuthn Security Key to be Added
	//
	// A human-readable name for the security key which will be added.
	RegisterDisplayName string `json:"webauthn_register_displayname"`
}
