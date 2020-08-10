package login

import (
	"net/http"
	"time"

	"github.com/gobuffalo/pop/v5"
	"github.com/gofrs/uuid"
	"github.com/pkg/errors"

	"github.com/ory/x/urlx"

	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/text"
	"github.com/ory/kratos/x"
)

// swagger:model loginRequest
type Request struct {
	// ID represents the request's unique ID. When performing the login flow, this
	// represents the id in the login ui's query parameter: http://<selfservice.flows.login.ui_url>/?request=<id>
	//
	// required: true
	ID uuid.UUID `json:"id" faker:"-" db:"id"`

	// ExpiresAt is the time (UTC) when the request expires. If the user still wishes to log in,
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

	// Active, if set, contains the login method that is being used. It is initially
	// not set.
	Active identity.CredentialsType `json:"active,omitempty" db:"active_method"`

	// Messages contains a list of messages to be displayed in the Login UI. Omitting these
	// messages makes it significantly harder for users to figure out what is going on.
	//
	// More documentation on messages can be found in the [User Interface Documentation](https://www.ory.sh/kratos/docs/concepts/ui-user-interface/).
	Messages text.Messages `json:"messages" db:"messages" faker:"-"`

	// Methods contains context for all enabled login methods. If a login request has been
	// processed, but for example the password is incorrect, this will contain error messages.
	//
	// required: true
	Methods map[identity.CredentialsType]*RequestMethod `json:"methods" faker:"login_request_methods" db:"-"`

	// MethodsRaw is a helper struct field for gobuffalo.pop.
	MethodsRaw RequestMethodsRaw `json:"-" faker:"-" has_many:"selfservice_login_request_methods" fk_id:"selfservice_login_request_id"`

	// CreatedAt is a helper struct field for gobuffalo.pop.
	CreatedAt time.Time `json:"-" db:"created_at"`

	// UpdatedAt is a helper struct field for gobuffalo.pop.
	UpdatedAt time.Time `json:"-" db:"updated_at"`

	// CSRFToken contains the anti-csrf token associated with this request.
	CSRFToken string `json:"-" db:"csrf_token"`

	// Forced stores whether this login request should enforce reauthentication.
	Forced bool `json:"forced" db:"forced"`

	// Type represents the flow's type which can be either "api" or "browser", depending on the flow interaction.
	Type flow.Type `json:"type" db:"type" faker:"flow_type"`
}

func NewRequest(exp time.Duration, csrf string, r *http.Request) *Request {
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
		CSRFToken:  csrf,
		Forced:     r.URL.Query().Get("refresh") == "true",
	}
}

func (r *Request) BeforeSave(_ *pop.Connection) error {
	r.MethodsRaw = make([]RequestMethod, 0, len(r.Methods))
	for _, m := range r.Methods {
		r.MethodsRaw = append(r.MethodsRaw, *m)
	}
	r.Methods = nil
	return nil
}

func (r *Request) AfterCreate(c *pop.Connection) error {
	return r.AfterFind(c)
}

func (r *Request) AfterUpdate(c *pop.Connection) error {
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

func (r Request) TableName() string {
	// This must be stay a value receiver, using a pointer receiver will cause issues with pop.
	return "selfservice_login_requests"
}

func (r *Request) Valid() error {
	if r.ExpiresAt.Before(time.Now()) {
		return errors.WithStack(newRequestExpiredError(time.Since(r.ExpiresAt)))
	}
	return nil
}

func (r *Request) GetID() uuid.UUID {
	return r.ID
}

func (r *Request) IsForced() bool {
	return r.Forced
}
