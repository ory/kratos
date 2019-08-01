package password

import (
	"net/http"

	"github.com/ory/hive/identity"
	"github.com/ory/hive/selfservice"
)

const (
	CredentialsType identity.CredentialsType = "password"
	csrfTokenName                            = "csrf_token"
)

type (
	csrfGenerator func(r *http.Request) string

	// CredentialsConfig is the struct that is being used as part of the identity credentials.
	CredentialsConfig struct {
		// HashedPassword is a hash-representation of the password.
		HashedPassword string `json:"hashed_password"`
	}

	// RequestMethodConfig contains the configuration for this selfservice strategy.
	RequestMethodConfig struct {
		// Action should be used as the form action URL (<form action="{{ .Action }}" method="post">).
		Action string `json:"action"`

		// Error contains any form errors.
		Error string `json:"error,omitempty"`

		// Fields contains the form fields.
		Fields selfservice.FormFields `json:"fields"`
	}

	// LoginFormPayload is used to decode the login form payload.
	LoginFormPayload struct {
		Password   string `form:"password"`
		Identifier string `form:"identifier"`
	}
)

func NewRequestMethodConfig() *RequestMethodConfig {
	return &RequestMethodConfig{Fields: selfservice.FormFields{}}
}

func (r *RequestMethodConfig) Reset() {
	r.Error = ""
	r.Fields.Reset()
}

func (r *RequestMethodConfig) SetError(err string) {
	r.Error = err
}

func (r *RequestMethodConfig) GetFormFields() selfservice.FormFields {
	return r.Fields
}
