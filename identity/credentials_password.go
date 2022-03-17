package identity

// CredentialsPassword is contains the configuration for credentials of the type password.
//
// swagger:model identityCredentialsPassword
type CredentialsPassword struct {
	// HashedPassword is a hash-representation of the password.
	HashedPassword string `json:"hashed_password"`
}
