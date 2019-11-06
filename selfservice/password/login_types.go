package password

import (
	"net/http"
	"time"

	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/selfservice"
)

func NewBlankLoginRequest(id string) *selfservice.LoginRequest {
	return &selfservice.LoginRequest{
		Request: &selfservice.Request{
			ID:             id,
			IssuedAt:       time.Now().UTC(),
			ExpiresAt:      time.Now().UTC(),
			RequestURL:     "",
			RequestHeaders: http.Header{},
			Methods: map[identity.CredentialsType]*selfservice.DefaultRequestMethod{
				identity.CredentialsTypePassword: {
					Method: identity.CredentialsTypePassword,
					Config: NewRequestMethodConfig(),
				},
			},
		},
	}
}
