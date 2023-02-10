package identity

// CredentialsPin contains the configuration for credentials of the pin type.
//
// swagger:model identityCredentialsPin
type CredentialsPin struct {
	// HashedPin is a hash-representation of the pin.
	HashedPin string `json:"hashed_pin"`
}
