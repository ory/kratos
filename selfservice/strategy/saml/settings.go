// Copyright © 2025 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package saml

import "encoding/json"

// Update settings flow using SAML
//
// swagger:model updateSettingsFlowWithSamlMethod
type _ struct {
	// Method
	//
	// Should be set to saml when trying to update a profile.
	//
	// required: true
	Method string `json:"method"`

	// Link this provider
	//
	// Either this or `unlink` must be set.
	//
	// type: string
	// in: body
	Link string `json:"link"`

	// Unlink this provider
	//
	// Either this or `link` must be set.
	//
	// type: string
	// in: body
	Unlink string `json:"unlink"`

	// Flow ID is the flow's ID.
	//
	// in: query
	FlowID string `json:"flow"`

	// The CSRF Token
	CSRFToken string `json:"csrf_token"`

	// The identity's traits
	//
	// in: body
	Traits json.RawMessage `json:"traits"`

	// Transient data to pass along to any webhooks
	//
	// required: false
	TransientPayload json.RawMessage `json:"transient_payload,omitempty" form:"transient_payload"`
}
