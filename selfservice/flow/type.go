// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package flow

// Type is the flow type.
//
// The flow type can either be `api` or `browser`.
//
// swagger:model selfServiceFlowType
type Type string

const (
	TypeAPI     Type = "api"
	TypeBrowser Type = "browser"
)

func (t Type) IsBrowser() bool {
	return t == TypeBrowser
}

func (t Type) IsAPI() bool {
	return t == TypeAPI
}
