package identity

import "github.com/ory/x/sqlxx"

const (
	ExpandFieldVerifiableAddresses   sqlxx.Expandable = "VerifiableAddresses"
	ExpandFieldRecoveryAddresses     sqlxx.Expandable = "RecoveryAddresses"
	ExpandFieldCredentials           sqlxx.Expandable = "InternalCredentials"
	ExpandFieldCredentialType        sqlxx.Expandable = "InternalCredentials.IdentityCredentialType"
	ExpandFieldCredentialIdentifiers sqlxx.Expandable = "InternalCredentials.CredentialIdentifiers"
)

// ExpandNothing expands nothing
var ExpandNothing = sqlxx.Expandables{}

// ExpandDefault expands the default fields:
//
// - Verifiable addresses
// - Recovery addresses
var ExpandDefault = sqlxx.Expandables{
	ExpandFieldVerifiableAddresses,
	ExpandFieldRecoveryAddresses,
}

// ExpandCredentials expands the identity's credentials.
var ExpandCredentials = sqlxx.Expandables{
	ExpandFieldCredentials,
	ExpandFieldCredentialType,
	ExpandFieldCredentialIdentifiers,
}

// ExpandEverything expands all the fields of an identity.
var ExpandEverything = sqlxx.Expandables{
	ExpandFieldVerifiableAddresses,
	ExpandFieldRecoveryAddresses,
	ExpandFieldCredentials,
	ExpandFieldCredentialType,
	ExpandFieldCredentialIdentifiers,
}
