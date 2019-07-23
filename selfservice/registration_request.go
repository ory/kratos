package selfservice

import (
	"net/http"
	"time"

	"github.com/pkg/errors"

	"github.com/ory/herodot"

	"github.com/ory/hive/identity"
)

var ErrRegistrationRequestExpired = herodot.ErrBadRequest.
	WithError("registration request expired")

type RegistrationRequestMethod struct {
	Method identity.CredentialsType `json:"method"`
	Config interface{}              `json:"config"`
}

type RegistrationRequest struct {
	*Request
	Methods map[identity.CredentialsType]*RegistrationRequestMethod `json:"methods" faker:"-"`
}

func NewRegistrationRequest(exp time.Duration, r *http.Request) *RegistrationRequest {
	return &RegistrationRequest{
		Request: newRequestFromHTTP(exp, r),
		Methods: make(map[identity.CredentialsType]*RegistrationRequestMethod),
	}
}

func (r *RegistrationRequest) GetID() string {
	return r.ID
}

func (r *RegistrationRequest) Valid() error {
	if r.ExpiresAt.Before(time.Now()) {
		return errors.WithStack(ErrRegistrationRequestExpired.WithReasonf("The registration request expired %.2f minutes ago, please try again", time.Now().Sub(r.ExpiresAt).Minutes()))
	}
	return nil
}
