// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package session

import (
	"strings"

	"github.com/ory/x/sqlxx"
)

// Expandable controls what fields to expand for sessions.
type Expandable = sqlxx.Expandable

const (
	// ExpandSessionDevices expands devices related to the session
	ExpandSessionDevices Expandable = "Devices"
	// ExpandSessionIdentity expands Identity related to the session
	ExpandSessionIdentity                  Expandable = "Identity"
	ExpandSessionIdentityRecoveryAddress   Expandable = "Identity.RecoveryAddresses"
	ExpandSessionIdentityVerifiableAddress Expandable = "Identity.VerifiableAddresses"
)

var expandablesMap = map[string]Expandable{
	"devices":  ExpandSessionDevices,
	"identity": ExpandSessionIdentity,
}

// Expandables is a list of Expandable values.
type Expandables = sqlxx.Expandables

func ParseExpandable(in string) (Expandable, bool) {
	e, ok := expandablesMap[strings.ToLower(in)]
	return e, ok
}

// ExpandNothing expands nothing
var ExpandNothing Expandables

// ExpandDefault expands the default fields of a session
// - Associated Identity
var ExpandDefault = Expandables{
	ExpandSessionIdentity,
}

// ExpandEverything expands all the fields of a session.
var ExpandEverything = Expandables{
	ExpandSessionDevices,
	ExpandSessionIdentity,
}
