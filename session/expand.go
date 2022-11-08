// Copyright © 2022 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package session

import "strings"

// Expandable controls what fields to expand for sessions.
// swagger:enum Expandable
type Expandable string

const (
	// ExpandSessionDevices expands devices related to the session
	ExpandSessionDevices Expandable = "Devices"
	// ExpandSessionIdentity expands Identity related to the session
	ExpandSessionIdentity Expandable = "Identity"
)

var expandablesMap = map[string]Expandable{
	"devices":  ExpandSessionDevices,
	"identity": ExpandSessionIdentity,
}

// Expandables is a list of Expandable values.
type Expandables []Expandable

// String returns a string representation of the Expandable.
func (e Expandable) String() string {
	return string(e)
}

// ToEager returns the fields used by pop's Eager command.
func (e Expandables) ToEager() []string {
	var s []string
	for _, e := range e {
		if e == ExpandSessionIdentity {
			continue
		}
		s = append(s, e.String())
	}
	return s
}

// Has returns true if the Expandable is in the list.
func (e Expandables) Has(search Expandable) bool {
	for _, e := range e {
		if e == search {
			return true
		}
	}
	return false
}

func ParseExpandable(in string) (Expandable, bool) {
	e, ok := expandablesMap[strings.ToLower(in)]
	return e, ok
}

// ExpandNothing expands nothing
var ExpandNothing []Expandable

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
