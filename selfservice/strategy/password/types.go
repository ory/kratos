package password

import "github.com/ory/kratos/selfservice/form"

type (
	// CredentialsConfig is the struct that is being used as part of the identity credentials.
	CredentialsConfig struct {
		// HashedPassword is a hash-representation of the password.
		HashedPassword string `json:"hashed_password"`
	}

	// LoginFormPayload is used to decode the login form payload.
	LoginFormPayload struct {
		// The user's password.
		Password string `form:"password" json:"password,omitempty"`

		// Identifier is the email or username of the user trying to log in.
		Identifier string `form:"identifier" json:"identifier,omitempty"`

		// Sending the anti-csrf token is only required for browser login flows.
		CSRFToken string `form:"csrf_token" json:"csrf_token"`
	}
)

// FlowMethod contains the configuration for this selfservice strategy.
type FlowMethod struct {
	*form.HTMLForm
}
