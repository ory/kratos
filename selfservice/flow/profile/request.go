package profile

import (
	"net/http"
	"time"

	"github.com/gofrs/uuid"
	"github.com/pkg/errors"

	"github.com/ory/herodot"
	"github.com/ory/x/urlx"

	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/selfservice/form"
	"github.com/ory/kratos/session"
	"github.com/ory/kratos/x"
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
	// ID represents the request's unique ID. When performing the profile management flow, this
	// represents the id in the profile ui's query parameter: http://<urls.profile_ui>?request=<id>
	//
	// type: string
	// format: uuid
	ID uuid.UUID `json:"id" db:"id" faker:"uuid" rw:"r"`

	// ExpiresAt is the time (UTC) when the request expires. If the user still wishes to update the profile,
	// a new request has to be initiated.
	ExpiresAt time.Time `json:"expires_at" faker:"time_type" db:"expires_at"`

	// IssuedAt is the time (UTC) when the request occurred.
	IssuedAt time.Time `json:"issued_at" faker:"time_type" db:"issued_at"`

	// RequestURL is the initial URL that was requested from ORY Kratos. It can be used
	// to forward information contained in the URL's path or query for example.
	RequestURL string `json:"request_url" db:"request_url"`

	// Form contains form fields, errors, and so on.
	Form *form.HTMLForm `json:"form" db:"form"`

	// Identity contains all of the identity's data in raw form.
	Identity *identity.Identity `json:"identity" faker:"identity" db:"-" belongs_to:"identities" fk_id:"IdentityID"`

	// UpdateSuccessful, if true, indicates that the profile has been updated successfully with the provided data.
	// Done will stay true when repeatedly checking. If set to true, done will revert back to false only
	// when a request with invalid (e.g. "please use a valid phone number") data was sent.
	UpdateSuccessful bool `json:"update_successful,omitempty" faker:"-" db:"update_successful"`

	// IdentityID is a helper struct field for gobuffalo.pop.
	IdentityID uuid.UUID `json:"-" faker:"-" db:"identity_id"`
	// CreatedAt is a helper struct field for gobuffalo.pop.
	CreatedAt time.Time `json:"-" faker:"-" db:"created_at"`
	// UpdatedAt is a helper struct field for gobuffalo.pop.
	UpdatedAt time.Time `json:"-" faker:"-" db:"updated_at"`
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
		ID:         x.NewUUID(),
		ExpiresAt:  time.Now().UTC().Add(exp),
		IssuedAt:   time.Now().UTC(),
		RequestURL: source.String(),
		IdentityID: s.Identity.ID,
		Identity:   s.Identity,
		Form:       new(form.HTMLForm),
	}
}

func (r *Request) TableName() string {
	return "selfservice_profile_management_requests"
}

func (r *Request) Valid(s *session.Session) error {
	if r.ExpiresAt.Before(time.Now()) {
		return errors.WithStack(ErrRequestExpired.WithReasonf("The profile request expired %.2f minutes ago, please try again.", time.Since(r.ExpiresAt).Minutes()))
	}
	if r.IdentityID != s.Identity.ID {
		return errors.WithStack(herodot.ErrBadRequest.WithReasonf("The profile request expired %.2f minutes ago, please try again", time.Since(r.ExpiresAt).Minutes()))
	}
	return nil
}
