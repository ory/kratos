package registration

import (
	"net/http"
	"net/url"
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

// swagger:model registrationFlow
type Flow struct {
	// ID represents the flow's unique ID. When performing the registration flow, this
	// represents the id in the registration ui's query parameter: http://<selfservice.flows.registration.ui_url>/?flow=<id>
	//
	// required: true
	ID uuid.UUID `json:"id" faker:"-" db:"id"`

	// Type represents the flow's type which can be either "api" or "browser", depending on the flow interaction.
	Type flow.Type `json:"type" db:"type" faker:"flow_type"`

	// ExpiresAt is the time (UTC) when the flow expires. If the user still wishes to log in,
	// a new flow has to be initiated.
	//
	// required: true
	ExpiresAt time.Time `json:"expires_at" faker:"time_type" db:"expires_at"`

	// IssuedAt is the time (UTC) when the flow occurred.
	//
	// required: true
	IssuedAt time.Time `json:"issued_at" faker:"time_type" db:"issued_at"`

	// RequestURL is the initial URL that was requested from ORY Kratos. It can be used
	// to forward information contained in the URL's path or query for example.
	//
	// required: true
	RequestURL string `json:"request_url" faker:"url" db:"request_url"`

	// Active, if set, contains the registration method that is being used. It is initially
	// not set.
	Active identity.CredentialsType `json:"active,omitempty" faker:"identity_credentials_type" db:"active_method"`

	// Messages contains a list of messages to be displayed in the Registration UI. Omitting these
	// messages makes it significantly harder for users to figure out what is going on.
	//
	// More documentation on messages can be found in the [User Interface Documentation](https://www.ory.sh/kratos/docs/concepts/ui-user-interface/).
	Messages text.Messages `json:"messages" db:"messages" faker:"-"`

	// Methods contains context for all enabled registration methods. If a registration flow has been
	// processed, but for example the password is incorrect, this will contain error messages.
	//
	// required: true
	Methods map[identity.CredentialsType]*FlowMethod `json:"methods" faker:"registration_flow_methods" db:"-"`

	// MethodsRaw is a helper struct field for gobuffalo.pop.
	MethodsRaw FlowMethodsRaw `json:"-" faker:"-" has_many:"selfservice_registration_flow_methods" fk_id:"selfservice_registration_flow_id"`

	// CreatedAt is a helper struct field for gobuffalo.pop.
	CreatedAt time.Time `json:"-" faker:"-" db:"created_at"`

	// UpdatedAt is a helper struct field for gobuffalo.pop.
	UpdatedAt time.Time `json:"-" faker:"-" db:"updated_at"`

	// CSRFToken contains the anti-csrf token associated with this flow. Only set for browser flows.
	CSRFToken string `json:"-" db:"csrf_token"`
}

func NewFlow(exp time.Duration, csrf string, r *http.Request, ft flow.Type) *Flow {
	source := urlx.Copy(r.URL)
	source.Host = r.Host

	source.Scheme = "http"
	if r.TLS != nil {
		source.Scheme = "https"
	}

	return &Flow{
		ID:         x.NewUUID(),
		ExpiresAt:  time.Now().UTC().Add(exp),
		IssuedAt:   time.Now().UTC(),
		RequestURL: source.String(),
		Methods:    map[identity.CredentialsType]*FlowMethod{},
		CSRFToken:  csrf,
		Type:       ft,
	}
}

func (r *Flow) BeforeSave(_ *pop.Connection) error {
	r.MethodsRaw = make([]FlowMethod, 0, len(r.Methods))
	for _, m := range r.Methods {
		r.MethodsRaw = append(r.MethodsRaw, *m)
	}
	r.Methods = nil
	return nil
}

func (r *Flow) AfterCreate(c *pop.Connection) error {
	return r.AfterFind(c)
}

func (r *Flow) AfterUpdate(c *pop.Connection) error {
	return r.AfterFind(c)
}

func (r *Flow) AfterFind(_ *pop.Connection) error {
	r.Methods = make(FlowMethods)
	for key := range r.MethodsRaw {
		m := r.MethodsRaw[key] // required for pointer dereference
		r.Methods[m.Method] = &m
	}
	r.MethodsRaw = nil
	return nil
}

func (r Flow) TableName() string {
	// This must be stay a value receiver, using a pointer receiver will cause issues with pop.
	return "selfservice_registration_flows"
}

func (r *Flow) GetID() uuid.UUID {
	return r.ID
}

func (r *Flow) Valid() error {
	if r.ExpiresAt.Before(time.Now()) {
		return errors.WithStack(NewFlowExpiredError(time.Since(r.ExpiresAt)))
	}
	return nil
}

func (f *Flow) AppendTo(src *url.URL) *url.URL {
	return urlx.CopyWithQuery(src, url.Values{"flow": {f.ID.String()}})
}
