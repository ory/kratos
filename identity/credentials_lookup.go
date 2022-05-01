package identity

import "github.com/ory/x/sqlxx"

// CredentialsLookup is the struct that is being used as part of the identity credentials.
//
// swagger:model identityCredentialsLookup
type CredentialsLookup struct {
	// RecoveryCodes is a list of recovery codes.
	RecoveryCodes []CredentialsLookupRecoveryCode `json:"recovery_codes"`
}

// CredentialsLookupRecoveryCode is a container of recovery codes and their usage date.
//
// swagger:model identityCredentialsLookupRecoveryCode
type CredentialsLookupRecoveryCode struct {
	// Code is a recovery code.
	Code string `json:"code"`

	// UsedAt indicates whether and when a recovery code was used.
	UsedAt sqlxx.NullTime `json:"used_at,omitempty"`
}
