package identity

import "github.com/ory/x/sqlxx"

// CredentialsLookupSecrets is the struct that is being used as part of the identity credentials.
//
// swagger:model identityCredentialsLookupSecrets
type CredentialsLookupSecrets struct {
	// LookupSecrets is a list of recovery codes.
	LookupSecrets []CredentialsLookupSecret `json:"lookup_secrets"`
}

// CredentialsLookupSecret is a container of recovery codes and their usage date.
//
// swagger:model identityCredentialsLookupSecret
type CredentialsLookupSecret struct {
	// Code is a recovery code.
	Code string `json:"code"`

	// UsedAt indicates whether and when a recovery code was used.
	UsedAt sqlxx.NullTime `json:"used_at,omitempty"`
}
