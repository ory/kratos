package profile

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/pkg/errors"

	"github.com/ory/herodot"
	"github.com/ory/x/urlx"

	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/selfservice/form"
	"github.com/ory/kratos/session"
)

// Request presents a profile management request
//
// This request is used when an identity wants to update profile information
// (especially traits) in a selfservice manner.
//
// For more information head over to: https://www.ory.sh/docs/kratos/selfservice/profile
//
// swagger:model profileManagementRequest
type Request struct {
	ID         string             `json:"id"`
	IssuedAt   time.Time          `json:"issued_at"`
	ExpiresAt  time.Time          `json:"expires_at"`
	RequestURL string             `json:"request_url"`
	identityID string             `json:"-"`
	Form       *form.HTMLForm     `json:"form"`
	Identity   *identity.Identity `json:"identity"`
}

func NewRequest(exp time.Duration, r *http.Request, s *session.Session) *Request {
	source := urlx.Copy(r.URL)
	source.Host = r.Host

	if len(source.Scheme) == 0 {
		source.Scheme = "http"
		if r.TLS != nil {
			source.Scheme = "https"
		}
	}

	return &Request{
		ID:         uuid.New().String(),
		ExpiresAt:  time.Now().UTC().Add(exp),
		IssuedAt:   time.Now().UTC(),
		RequestURL: source.String(),
		identityID: s.Identity.ID,
		Form:       new(form.HTMLForm),
	}
}

func (r *Request) Valid(s *session.Session) error {
	if r.ExpiresAt.Before(time.Now()) {
		return errors.WithStack(ErrRequestExpired.WithReasonf("The profile request expired %.2f minutes ago, please try again.", time.Since(r.ExpiresAt).Minutes()))
	}
	if r.identityID != s.Identity.ID {
		return errors.WithStack(herodot.ErrBadRequest.WithReasonf("The profile request expired %.2f minutes ago, please try again", time.Since(r.ExpiresAt).Minutes()))
	}
	return nil
}
