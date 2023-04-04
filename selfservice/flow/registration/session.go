// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package registration

import (
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/session"
)

// The Response for Registration Flows via API
//
// swagger:model successfulNativeRegistration
type APIFlowResponse struct {
	// The Session Token
	//
	// This field is only set when the session hook is configured as a post-registration hook.
	//
	// A session token is equivalent to a session cookie, but it can be sent in the HTTP Authorization
	// Header:
	//
	// 		Authorization: bearer ${session-token}
	//
	// The session token is only issued for API flows, not for Browser flows!
	Token string `json:"session_token,omitempty"`

	// The Session
	//
	// This field is only set when the session hook is configured as a post-registration hook.
	//
	// The session contains information about the user, the session device, and so on.
	// This is only available for API flows, not for Browser flows!
	Session *session.Session `json:"session,omitempty"`

	// The Identity
	//
	// The identity that just signed up.
	//
	// required: true
	Identity *identity.Identity `json:"identity"`

	// Contains a list of actions, that could follow this flow
	//
	// It can, for example, this will contain a reference to the verification flow, created as part of the user's
	// registration or the token of the session.
	//
	// required: false
	ContinueWith []flow.ContinueWith `json:"continue_with"`
}
