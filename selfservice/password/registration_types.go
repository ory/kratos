package password

import (
	"net/http"
	"time"

	"github.com/ory/hive/identity"
	"github.com/ory/hive/selfservice"
)

type RegistrationRequestMethodConfig struct {
	Action string                 `json:"action"`
	Error  string                 `json:"error,omitempty"`
	Fields selfservice.FormFields `json:"fields"`
}

func NewRegistrationRequestMethodConfig() *RegistrationRequestMethodConfig {
	return &RegistrationRequestMethodConfig{Fields: selfservice.FormFields{}}
}

func (r *RegistrationRequestMethodConfig) Reset() {
	r.Error = ""
	r.Fields.Reset()
}

func (r *RegistrationRequestMethodConfig) SetError(err string) {
	r.Error = err
}

func (r *RegistrationRequestMethodConfig) GetFormFields() selfservice.FormFields {
	return r.Fields
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
				Config: NewRegistrationRequestMethodConfig(),
			},
		},
	}
}
