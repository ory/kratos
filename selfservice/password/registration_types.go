package password

import (
	"net/http"
	"time"

	"github.com/ory/hive/identity"
	"github.com/ory/hive/selfservice"
)

type RegistrationRequestMethodConfig struct {
	Action string     `json:"action"`
	Error  string     `json:"error,omitempty"`
	Fields FormFields `json:"fields"`
}

func (r *RegistrationRequestMethodConfig) Reset() {
	r.Error = ""
	r.Fields.Reset()
}

func NewBlankRegistrationRequest(id string) *selfservice.RegistrationRequest {
	return &selfservice.RegistrationRequest{
		Request: &selfservice.Request{
			ID:             id,
			IssuedAt:       time.Now().UTC(),
			ExpiresAt:      time.Now().UTC(),
			RequestURL:     "",
			RequestHeaders: http.Header{},
		},
		Methods: map[identity.CredentialsType]*selfservice.RegistrationRequestMethod{
			CredentialsType: {
				Method: CredentialsType,
				Config: &RegistrationRequestMethodConfig{
					Fields: map[string]FormField{},
				},
			},
		},
	}
}
