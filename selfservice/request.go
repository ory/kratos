package selfservice

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/pkg/errors"

	"github.com/ory/x/urlx"

	"github.com/ory/hive/identity"
)

type RequestMethodConfig interface {
	Reset()
	SetError(err string)
	GetFormFields() FormFields
}

type RequestMethod interface {
	GetConfig() RequestMethodConfig
}

type DefaultRequestMethod struct {
	Method identity.CredentialsType `json:"method"`
	Config RequestMethodConfig      `json:"config" faker:"-"`
}

type RegistrationRequest struct{ *Request }

func NewRegistrationRequest(exp time.Duration, r *http.Request) *RegistrationRequest {
	return &RegistrationRequest{Request: newRequestFromHTTP(exp, r)}
}

func (r *RegistrationRequest) Valid() error {
	if r.ExpiresAt.Before(time.Now()) {
		return errors.WithStack(ErrRegistrationRequestExpired.WithReasonf("The registration request expired %.2f minutes ago, please try again", time.Now().Sub(r.ExpiresAt).Minutes()))
	}
	return nil
}

type LoginRequest struct{ *Request }

func NewLoginRequest(exp time.Duration, r *http.Request) *LoginRequest {
	return &LoginRequest{Request: newRequestFromHTTP(exp, r)}
}

func (r *LoginRequest) Valid() error {
	if r.ExpiresAt.Before(time.Now()) {
		return errors.WithStack(ErrLoginRequestExpired.WithReasonf("The login request expired %.2f minutes ago, please try again", time.Now().Sub(r.ExpiresAt).Minutes()))
	}
	return nil
}

type Request struct {
	ID             string                                             `json:"id"`
	IssuedAt       time.Time                                          `json:"issued_at"`
	ExpiresAt      time.Time                                          `json:"expires_at"`
	RequestURL     string                                             `json:"request_url"`
	RequestHeaders http.Header                                        `json:"headers"`
	Active         identity.CredentialsType                           `json:"active,omitempty"`
	Methods        map[identity.CredentialsType]*DefaultRequestMethod `json:"methods" faker:"-"`
}

func (r *Request) GetID() string {
	return r.ID
}

func newRequestFromHTTP(exp time.Duration, r *http.Request) *Request {
	source := urlx.Copy(r.URL)
	source.Host = r.Host

	if len(source.Scheme) == 0 {
		source.Scheme = "http"
		if r.TLS != nil {
			source.Scheme = "https"
		}
	}

	return &Request{
		ID:             uuid.New().String(),
		IssuedAt:       time.Now().UTC(),
		ExpiresAt:      time.Now().UTC().Add(exp),
		RequestURL:     source.String(),
		RequestHeaders: r.Header,
		Methods:        map[identity.CredentialsType]*DefaultRequestMethod{},
	}
}
