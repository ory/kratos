// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package password

import (
	"github.com/ory/kratos/ui/container"
)

// Update Login Flow with Password Method
//
// swagger:model updateLoginFlowWithPasswordMethod
type updateLoginFlowWithPasswordMethod struct {
	// Method should be set to "password" when logging in using the identifier and password strategy.
	//
	// required: true
	Method string `json:"method"`

	// Sending the anti-csrf token is only required for browser login flows.
	CSRFToken string `json:"csrf_token"`

	// The user's password.
	//
	// required: true
	Password string `json:"password"`

	// Identifier is the email or username of the user trying to log in.
	// This field is deprecated!
	LegacyIdentifier string `json:"password_identifier"`

	// Identifier is the email or username of the user trying to log in.
	//
	// required: true
	Identifier string `json:"identifier"`
}

// FlowMethod contains the configuration for this selfservice strategy.
type FlowMethod struct {
	*container.Container
}
