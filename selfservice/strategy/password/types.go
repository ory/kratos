package password

import (
	"github.com/ory/kratos/ui/container"
)

// CredentialsConfig is the struct that is being used as part of the identity credentials.
//type CredentialsConfig struct {
//	// HashedPassword is a hash-representation of the password.
//	HashedPassword string `json:"hashed_password"`
//}

// submitSelfServiceLoginFlowWithPasswordMethodBody is used to decode the login form payload.
//
// swagger:model submitSelfServiceLoginFlowWithPasswordMethodBody
type submitSelfServiceLoginFlowWithPasswordMethodBody struct {
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
	//
	// required: true
	Identifier string `json:"password_identifier"`
}

// FlowMethod contains the configuration for this selfservice strategy.
type FlowMethod struct {
	*container.Container
}
