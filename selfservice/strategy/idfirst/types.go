// Copyright © 2024 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package idfirst

import "encoding/json"

// Update Login Flow with Multi-Step Method
//
// swagger:model updateLoginFlowWithIdentifierFirstMethod
type UpdateLoginFlowWithIdentifierFirstMethod struct {
	// Method should be set to "password" when logging in using the identifier and password strategy.
	//
	// required: true
	Method string `json:"method"`

	// Sending the anti-csrf token is only required for browser login flows.
	CSRFToken string `json:"csrf_token"`

	// Identifier is the email or username of the user trying to log in.
	//
	// required: true
	Identifier string `json:"identifier"`

	// Transient data to pass along to any webhooks
	//
	// required: false
	TransientPayload json.RawMessage `json:"transient_payload,omitempty" form:"transient_payload"`
}
