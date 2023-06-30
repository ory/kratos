package identity

// CredentialsOTP represents an OTP code
//
// swagger:model identityCredentialsOTP
type CredentialsOTP struct {
	// CodeHMAC represents the HMACed value of the login/registration code
	CodeHMAC string `json:"code_hmac"`
}
