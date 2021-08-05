package totp

// CredentialsConfig is the struct that is being used as part of the identity credentials.
type CredentialsConfig struct {
	// TOTPURL is the TOTP URL
	//
	// For more details see: https://github.com/google/google-authenticator/wiki/Key-Uri-Format
	TOTPURL string `json:"totp_url"`
}
