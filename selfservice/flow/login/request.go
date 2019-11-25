package login

import (
	"net/http"
	"time"

	"github.com/gofrs/uuid"
	"github.com/pkg/errors"

	"github.com/ory/herodot"
	"github.com/ory/x/urlx"

	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/x"
)

// swagger:model loginRequest
type Request struct {
	// ID represents the request's unique ID. When performing the login flow, this
	// represents the id in the login ui's query parameter: http://<urls.login_ui>/?request=<id>
	ID uuid.UUID `json:"id" faker:"uuid"`

	// ExpiresAt is the time (UTC) when the request expires. If the user still wishes to log in,
	// a new request has to be initiated.
	ExpiresAt time.Time `json:"expires_at" faker:"time_type"`

	// IssuedAt is the time (UTC) when the request occurred.
	IssuedAt time.Time `json:"issued_at" faker:"time_type"`

	// RequestURL is the initial URL that was requested from ORY Kratos. It can be used
	// to forward information contained in the URL's path or query for example.
	RequestURL string `json:"request_url"`

	// Active, if set, contains the login method that is being used. It is initially
	// not set.
	Active identity.CredentialsType `json:"active,omitempty"`

	// Methods contains context for all enabled login methods. If a login request has been
	// processed, but for example the password is incorrect, this will contain error messages.
	Methods map[identity.CredentialsType]*RequestMethod `json:"methods" faker:"login_request_methods"`

	// MethodsRaw is a helper struct field for gobuffalo.pop.
	MethodsRaw []RequestMethod `json:"-" faker:"-"`
}

func NewLoginRequest(exp time.Duration, r *http.Request) *Request {
	source := urlx.Copy(r.URL)
	source.Host = r.Host

	if len(source.Scheme) == 0 {
		source.Scheme = "http"
		if r.TLS != nil {
			source.Scheme = "https"
		}
	}

	return &Request{
		ID:         x.NewUUID(),
		ExpiresAt:  time.Now().UTC().Add(exp),
		IssuedAt:   time.Now().UTC(),
		RequestURL: source.String(),
		Methods:    map[identity.CredentialsType]*RequestMethod{},
	}
}

func (r *Request) TableName() string {
	return "login_requests"
}

func (r *Request) Valid() error {
	if r.ExpiresAt.Before(time.Now()) {
		return errors.WithStack(ErrRequestExpired.WithReasonf("The login request expired %.2f minutes ago, please try again.", time.Since(r.ExpiresAt).Minutes()))
	}

	if r.IssuedAt.After(time.Now()) {
		return errors.WithStack(herodot.ErrBadRequest.WithReason("The login request was issued in the future."))
	}
	return nil
}

func (r *Request) GetID() uuid.UUID {
	return r.ID
}
