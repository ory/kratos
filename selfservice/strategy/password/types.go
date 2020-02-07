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
		Password   string `form:"password"`
		Identifier string `form:"identifier"`
	}
)

// RequestMethod contains the configuration for this selfservice strategy.
type RequestMethod struct {
	*form.HTMLForm
}
