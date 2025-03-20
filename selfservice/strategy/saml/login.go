// Copyright © 2025 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package saml

import "encoding/json"

// Update login flow using SAML
//
// swagger:model updateLoginFlowWithSamlMethod
type _ struct {
	// The provider to register with
	//
	// required: true
	Provider string `json:"provider"`

	// The CSRF Token
	CSRFToken string `json:"csrf_token"`

	// Method to use
	//
	// This field must be set to `saml` when using the saml method.
	//
	// required: true
	Method string `json:"method"`

	// Transient data to pass along to any webhooks
	//
	// required: false
	TransientPayload json.RawMessage `json:"transient_payload,omitempty" form:"transient_payload"`
}
