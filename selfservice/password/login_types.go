package password

import (
	"net/http"
	"time"

	"github.com/ory/hive/identity"
	"github.com/ory/hive/selfservice"
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
