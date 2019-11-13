package login

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/pkg/errors"

	"github.com/ory/herodot"
	"github.com/ory/x/urlx"

	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/selfservice/form"
)

// swagger:model loginRequestMethod
type RequestMethod struct {
	// Method contains the request credentials type.
	Method identity.CredentialsType `json:"method"`

	// Config is the credential type's config.
	Config RequestMethodConfig `json:"config"`
}

// swagger:model loginRequestMethodConfig
type RequestMethodConfig interface {
	form.ErrorParser
	form.ValueSetter
	form.Resetter
	form.CSRFSetter
}

// swagger:model loginRequest
type Request struct {
	// ID represents the request's unique ID. When performing the login flow, this
	// represents the id in the login ui's query parameter: http://login-ui/?request=<id>
	ID string `json:"id"`

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

	RequestHeaders http.Header `json:"-" faker:"http_header"`
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
		ID:             uuid.New().String(),
		ExpiresAt:      time.Now().UTC().Add(exp),
		IssuedAt:       time.Now().UTC(),
		RequestURL:     source.String(),
		RequestHeaders: r.Header,
		Methods:        map[identity.CredentialsType]*RequestMethod{},
	}
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

func (r *Request) GetID() string {
	return r.ID
}

// Declassify returns a copy of the Request where all sensitive information
// such as request headers is removed.
func (r *Request) Declassify() *Request {
	rr := *r
	rr.RequestHeaders = http.Header{}
	return &rr
}
