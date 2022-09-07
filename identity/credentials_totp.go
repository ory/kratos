package identity

// CredentialsTOTP is the configuration for credentials of the type totp.
//
// swagger:model identityCredentialsTotp
type CredentialsTOTP struct {
	// TOTPURL is the TOTP URL
	//
	// For more details see: https://github.com/google/google-authenticator/wiki/Key-Uri-Format
	//
	// required: true
	TOTPURL string `json:"totp_url"`
}
