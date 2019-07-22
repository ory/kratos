package password

import (
	"net/http"
	"time"

	"github.com/ory/hive-cloud/hive/identity"
	"github.com/ory/hive-cloud/hive/selfservice"
)

type LoginRequestMethodConfig struct {
	Action string     `json:"action"`
	Error  string     `json:"error,omitempty"`
	Fields FormFields `json:"fields"`
}

type LoginFormPayload struct {
	Password   string `form:"password"`
	Identifier string `form:"identifier"`
}

func (r *LoginRequestMethodConfig) Reset() {
	r.Error = ""
	r.Fields.Reset()
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
				Config: &LoginRequestMethodConfig{
					Fields: map[string]FormField{},
				},
			},
		},
	}
}
