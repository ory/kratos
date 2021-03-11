package password

import (
	"github.com/ory/kratos/ui/container"
)

type (
	// CredentialsConfig is the struct that is being used as part of the identity credentials.
	CredentialsConfig struct {
		// HashedPassword is a hash-representation of the password.
		HashedPassword string `json:"hashed_password"`
	}

	// CompleteSelfServiceLoginFlowWithPasswordMethod is used to decode the login form payload.
	CompleteSelfServiceLoginFlowWithPasswordMethod struct {
		// Method should be set to "password" when logging in using the identifier and password strategy.
		Method string `json:"method"`

		// Sending the anti-csrf token is only required for browser login flows.
		CSRFToken string `json:"csrf_token"`

		Password CompleteSelfServiceLoginFlowWithPasswordMethodPayload `json:"password"`
	}

	CompleteSelfServiceLoginFlowWithPasswordMethodPayload struct {
		// The user's password.
		Password string `json:"password,omitempty"`

		// Identifier is the email or username of the user trying to log in.
		Identifier string `json:"identifier,omitempty"`
	}
)

// FlowMethod contains the configuration for this selfservice strategy.
type FlowMethod struct {
	*container.Container
}
