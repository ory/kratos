package selfservice

import (
	"net/http"
	"time"

	"github.com/pkg/errors"

	"github.com/ory/herodot"

	"github.com/ory/hive-cloud/hive/identity"
)

var ErrLoginRequestExpired = herodot.ErrBadRequest.
	WithError("login request expired")

type LoginRequestMethod struct {
	Method identity.CredentialsType `json:"method"`
	Config interface{}              `json:"config" faker:"-"`
}

type LoginRequest struct {
	*Request
	Methods map[identity.CredentialsType]*LoginRequestMethod `json:"methods" faker:"-"`
}

func NewLoginRequest(exp time.Duration, r *http.Request) *LoginRequest {
	return &LoginRequest{
		Request: newRequestFromHTTP(exp, r),
		Methods: make(map[identity.CredentialsType]*LoginRequestMethod),
	}
}

func (r *LoginRequest) GetID() string {
	return r.ID
}

func (r *LoginRequest) Valid() error {
	if r.ExpiresAt.Before(time.Now()) {
		return errors.WithStack(ErrLoginRequestExpired.WithReasonf("The login request expired %.2f minutes ago, please try again", time.Now().Sub(r.ExpiresAt).Minutes()))
	}
	return nil
}
