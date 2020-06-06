package settings

import (
	"net/http"
	"net/url"
	"time"

	"github.com/gobuffalo/pop/v5"
	"github.com/gofrs/uuid"
	"github.com/pkg/errors"

	"github.com/ory/x/urlx"

	"github.com/ory/x/sqlxx"

	"github.com/ory/herodot"

	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/session"
	"github.com/ory/kratos/text"
	"github.com/ory/kratos/x"
)

// Request presents a settings request
//
// This request is used when an identity wants to update settings
// (e.g. profile data, passwords, ...) in a selfservice manner.
//
// We recommend reading the [User Settings Documentation](../self-service/flows/user-settings)
//
// swagger:model settingsRequest
type Request struct {
	// ID represents the request's unique ID. When performing the settings flow, this
	// represents the id in the settings ui's query parameter: http://<urls.settings_ui>?request=<id>
	//
	// required: true
	// type: string
	// format: uuid
	ID uuid.UUID `json:"id" db:"id" faker:"-"`

	// ExpiresAt is the time (UTC) when the request expires. If the user still wishes to update the setting,
	// a new request has to be initiated.
	//
	// required: true
	ExpiresAt time.Time `json:"expires_at" faker:"time_type" db:"expires_at"`

	// IssuedAt is the time (UTC) when the request occurred.
	//
	// required: true
	IssuedAt time.Time `json:"issued_at" faker:"time_type" db:"issued_at"`

	// RequestURL is the initial URL that was requested from ORY Kratos. It can be used
	// to forward information contained in the URL's path or query for example.
	//
	// required: true
	RequestURL string `json:"request_url" db:"request_url"`

	// Active, if set, contains the registration method that is being used. It is initially
	// not set.
	Active sqlxx.NullString `json:"active,omitempty" db:"active_method"`

	// Messages contains a list of messages to be displayed in the Settings UI. Omitting these
	// messages makes it significantly harder for users to figure out what is going on.
	//
	// More documentation on messages can be found in the [User Interface Documentation](https://www.ory.sh/kratos/docs/concepts/ui-user-interface/).
	Messages text.Messages `json:"messages" db:"messages" faker:"-"`

	// Methods contains context for all enabled registration methods. If a registration request has been
	// processed, but for example the password is incorrect, this will contain error messages.
	//
	// required: true
	Methods map[string]*RequestMethod `json:"methods" faker:"settings_request_methods" db:"-"`

	// MethodsRaw is a helper struct field for gobuffalo.pop.
	MethodsRaw RequestMethodsRaw `json:"-" faker:"-" has_many:"selfservice_settings_request_methods" fk_id:"selfservice_settings_request_id"`

	// Identity contains all of the identity's data in raw form.
	//
	// required: true
	Identity *identity.Identity `json:"identity" faker:"identity" db:"-" belongs_to:"identities" fk_id:"IdentityID"`

	// Success, if true, indicates that the settings request has been updated successfully with the provided data.
	// Done will stay true when repeatedly checking. If set to true, done will revert back to false only
	// when a request with invalid (e.g. "please use a valid phone number") data was sent.
	//
	// required: true
	UpdateSuccessful bool `json:"update_successful" faker:"-" db:"update_successful"`

	// IdentityID is a helper struct field for gobuffalo.pop.
	IdentityID uuid.UUID `json:"-" faker:"-" db:"identity_id"`
	// CreatedAt is a helper struct field for gobuffalo.pop.
	CreatedAt time.Time `json:"-" faker:"-" db:"created_at"`
	// UpdatedAt is a helper struct field for gobuffalo.pop.
	UpdatedAt time.Time `json:"-" faker:"-" db:"updated_at"`
}

func NewRequest(exp time.Duration, r *http.Request, s *session.Session) *Request {
	return &Request{
		ID:         x.NewUUID(),
		ExpiresAt:  time.Now().UTC().Add(exp),
		IssuedAt:   time.Now().UTC(),
		RequestURL: x.RequestURL(r).String(),
		IdentityID: s.Identity.ID,
		Identity:   s.Identity,
		Methods:    map[string]*RequestMethod{},
	}
}

func (r *Request) TableName() string {
	return "selfservice_settings_requests"
}

func (r *Request) GetID() uuid.UUID {
	return r.ID
}

func (r *Request) URL(settingsURL *url.URL) *url.URL {
	return urlx.CopyWithQuery(settingsURL, url.Values{"request": {r.ID.String()}})
}

func (r *Request) Valid(s *session.Session) error {
	if r.ExpiresAt.Before(time.Now().UTC()) {
		return errors.WithStack(ErrRequestExpired.
			WithReasonf("The settings request expired %.2f minutes ago, please try again.",
				-time.Since(r.ExpiresAt).Minutes()))
	}
	if r.IdentityID != s.Identity.ID {
		return errors.WithStack(herodot.ErrBadRequest.WithReasonf(
			"You must restart the flow because the resumable session was initiated by another person."))
	}
	return nil
}

func (r *Request) BeforeSave(_ *pop.Connection) error {
	r.MethodsRaw = make([]RequestMethod, 0, len(r.Methods))
	for _, m := range r.Methods {
		r.MethodsRaw = append(r.MethodsRaw, *m)
	}
	r.Methods = nil
	return nil
}

func (r *Request) AfterSave(c *pop.Connection) error {
	return r.AfterFind(c)
}

func (r *Request) AfterFind(_ *pop.Connection) error {
	r.Methods = make(RequestMethods)
	for key := range r.MethodsRaw {
		m := r.MethodsRaw[key] // required for pointer dereference
		r.Methods[m.Method] = &m
	}
	r.MethodsRaw = nil
	return nil
}
