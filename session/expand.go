// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package session

import (
	"fmt"
	"strings"

	"github.com/ory/kratos/identity"
	"github.com/ory/x/sqlxx"
)

// Expandable controls what fields to expand for sessions.
type Expandable = sqlxx.Expandable

const (
	ExpandSessionDevices  Expandable = "Devices"
	ExpandSessionIdentity Expandable = "Identity"
)

// Expandables is a list of Expandable values.
type Expandables = sqlxx.Expandables

func ParseExpandables(in []string) (Expandables, error) {
	es := ExpandNothing
	for _, s := range in {
		switch strings.ToLower(s) {
		case "":
			continue
		case "devices":
			es = append(es, ExpandSessionDevices)
		case "identity":
			es = append(es, ExpandSessionIdentity)
		case "addresses":
			es = append(es, ExpandSessionIdentity, identity.ExpandFieldRecoveryAddresses, identity.ExpandFieldVerifiableAddresses)
		case "credentials":
			es = append(es, ExpandSessionIdentity, identity.ExpandFieldCredentials)
		default:
			return nil, fmt.Errorf("unknown expand value: %q", s)
		}
	}
	return es, nil
}

// ExpandNothing expands nothing
var ExpandNothing Expandables

// ExpandDefault expands the default fields of a session
// - Associated Identity
var ExpandDefault = Expandables{
	ExpandSessionDevices,
	ExpandSessionIdentity,
	identity.ExpandFieldRecoveryAddresses,
	identity.ExpandFieldVerifiableAddresses,
}

// ExpandEverything expands all the fields of a session.
var ExpandEverything = Expandables{
	ExpandSessionDevices,
	ExpandSessionIdentity,
	identity.ExpandFieldRecoveryAddresses,
	identity.ExpandFieldVerifiableAddresses,
	identity.ExpandFieldCredentials,
}
