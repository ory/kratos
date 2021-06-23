package login

import "github.com/ory/kratos/session"

// The Response for Login Flows via API
//
// swagger:model loginViaApiResponse
type APIFlowResponse struct {
	// The Session Token
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
	// The session contains information about the user, the session device, and so on.
	// This is only available for API flows, not for Browser flows!
	//
	// required: true
	Session *session.Session `json:"session"`
}
