package identity

// CredentialsTOTP is the configuration for credentials of the type totp.
//
// swagger:model identityCredentialsTotp
type CredentialsTOTP struct {
	TOTPURL string `json:"totp_url"`
}
