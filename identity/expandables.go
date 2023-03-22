// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package identity

import "github.com/ory/x/sqlxx"

type Expandable = sqlxx.Expandable
type Expandables = sqlxx.Expandables

const (
	ExpandFieldVerifiableAddresses Expandable = "VerifiableAddresses"
	ExpandFieldRecoveryAddresses   Expandable = "RecoveryAddresses"
	ExpandFieldCredentials         Expandable = "Credentials"
)

// ExpandNothing expands nothing
var ExpandNothing = Expandables{}

// ExpandDefault expands the default fields:
//
// - Verifiable addresses
// - Recovery addresses
var ExpandDefault = Expandables{
	ExpandFieldVerifiableAddresses,
	ExpandFieldRecoveryAddresses,
}

// ExpandCredentials expands the identity's credentials.
var ExpandCredentials = Expandables{
	ExpandFieldCredentials,
}

// ExpandEverything expands all the fields of an identity.
var ExpandEverything = Expandables{
	ExpandFieldVerifiableAddresses,
	ExpandFieldRecoveryAddresses,
	ExpandFieldCredentials,
}
