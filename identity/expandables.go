// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package identity

import "github.com/ory/x/sqlxx"

type Expandable = sqlxx.Expandable
type Expandables = sqlxx.Expandables

// Each value must be the exact name of an Identity struct field: the values are passed to pop's
// Eager/EagerPreload for reflection-based loading, and the persister runs one query per value.
// That is why the two address expandables cannot be merged into one: they name distinct fields
// backed by distinct tables. To load both, compose them in an Expandables set (see ExpandDefault).
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

// ExpandEverythingButCredentials expands all the fields of an identity except its credentials.
// Expanding credentials is by far the most expensive part of fetching an identity, so prefer
// this expansion on hot paths that never read the credentials.
var ExpandEverythingButCredentials = Expandables{
	ExpandFieldVerifiableAddresses,
	ExpandFieldRecoveryAddresses,
}
