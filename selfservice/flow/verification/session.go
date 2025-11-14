// Copyright © 2025 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package verification

import (
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/session"
)

// The Response for Verification Flows via API
//
// swagger:model successfulNativeVerification
type APIFlowResponse struct {
	// The Session Token
	//
	// This field is only set when the session hook is configured as a post-verification hook.
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
	// This field is only set when the session hook is configured as a post-verification hook.
	//
	// The session contains information about the user, the session device, and so on.
	// This is only available for API flows, not for Browser flows!
	Session *session.Session `json:"session,omitempty"`

	// The Identity
	//
	// The identity that was verified.
	//
	// required: true
	Identity *identity.Identity `json:"identity"`

	// Contains a list of actions, that could follow this flow
	//
	// It can, for example, contain a reference to another flow or the session token.
	//
	// required: false
	ContinueWith []flow.ContinueWith `json:"continue_with"`
}
