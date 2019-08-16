package identity

import "encoding/json"

// CredentialsType  represents several different credential types, like password credentials, passwordless credentials,
// and so on.
type CredentialsType string

const (
	CredentialsTypePassword CredentialsType = "password"
	CredentialsTypeOIDC     CredentialsType = "oidc"
)

// Credentials represents a specific credential type
//
// swagger:model identityCredentials
type Credentials struct {
	// RequestID discriminates between different credential types.
	ID CredentialsType `json:"id"`

	// Identifiers represents a list of unique identifiers this credential type matches.
	Identifiers []string `json:"identifiers"`

	// Config contains the concrete credential payload. This might contain the bcrypt-hashed password, or the email
	// for passwordless authentication.
	Config json.RawMessage `json:"config"`
}
