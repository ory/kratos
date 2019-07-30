package password

import (
	"net/http"
	"time"

	"github.com/ory/hive/identity"
	"github.com/ory/hive/selfservice"
)

type LoginFormPayload struct {
	Password   string `form:"password"`
	Identifier string `form:"identifier"`
}

type LoginRequestMethodConfig struct {
	Action string                 `json:"action"`
	Error  string                 `json:"error,omitempty"`
	Fields selfservice.FormFields `json:"fields"`
}

func NewLoginRequestMethodConfig() *LoginRequestMethodConfig {
	return &LoginRequestMethodConfig{Fields: selfservice.FormFields{}}
}

func (r *LoginRequestMethodConfig) Reset() {
	r.Error = ""
	r.Fields.Reset()
}

func (r *LoginRequestMethodConfig) SetError(err string) {
	r.Error = err
}

func (r *LoginRequestMethodConfig) GetFormFields() selfservice.FormFields {
	return r.Fields
}

func NewBlankLoginRequest(id string) *selfservice.LoginRequest {
	return &selfservice.LoginRequest{
		Request: &selfservice.Request{
			ID:             id,
			IssuedAt:       time.Now().UTC(),
			ExpiresAt:      time.Now().UTC(),
			RequestURL:     "",
			RequestHeaders: http.Header{},
		},
		Methods: map[identity.CredentialsType]*selfservice.LoginRequestMethod{
			CredentialsType: {
				Method: CredentialsType,
				Config: NewLoginRequestMethodConfig(),
			},
		},
	}
}
