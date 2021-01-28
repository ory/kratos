package login

import (
	"context"
	"net/http"
	"net/url"
	"time"

	"github.com/ory/kratos/corp"

	"github.com/gobuffalo/pop/v5"
	"github.com/gofrs/uuid"
	"github.com/pkg/errors"

	"github.com/ory/x/urlx"

	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/text"
	"github.com/ory/kratos/x"
)

// Login Flow
//
// This object represents a login flow. A login flow is initiated at the "Initiate Login API / Browser Flow"
// endpoint by a client.
//
// Once a login flow is completed successfully, a session cookie or session token will be issued.
//
// swagger:model loginFlow
type Flow struct {
	// ID represents the flow's unique ID. When performing the login flow, this
	// represents the id in the login UI's query parameter: http://<selfservice.flows.login.ui_url>/?flow=<flow_id>
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

	// IssuedAt is the time (UTC) when the flow started.
	//
	// required: true
	IssuedAt time.Time `json:"issued_at" faker:"time_type" db:"issued_at"`

	// RequestURL is the initial URL that was requested from ORY Kratos. It can be used
	// to forward information contained in the URL's path or query for example.
	//
	// required: true
	RequestURL string `json:"request_url" db:"request_url"`

	// The active login method
	//
	// If set contains the login method used. If the flow is new, it is unset.
	Active identity.CredentialsType `json:"active,omitempty" db:"active_method"`

	// Messages contains a list of messages to be displayed in the Login UI. Omitting these
	// messages makes it significantly harder for users to figure out what is going on.
	//
	// More documentation on messages can be found in the [User Interface Documentation](https://www.ory.sh/kratos/docs/concepts/ui-user-interface/).
	Messages text.Messages `json:"messages" db:"messages" faker:"-"`

	// List of login methods
	//
	// This is the list of available login methods with their required form fields, such as `identifier` and `password`
	// for the password login method. This will also contain error messages such as "password can not be empty".
	//
	// required: true
	Methods map[identity.CredentialsType]*FlowMethod `json:"methods" faker:"login_flow_methods" db:"-"`

	// MethodsRaw is a helper struct field for gobuffalo.pop.
	MethodsRaw []FlowMethod `json:"-" faker:"-" has_many:"selfservice_login_flow_methods" fk_id:"selfservice_login_flow_id"`

	// CreatedAt is a helper struct field for gobuffalo.pop.
	CreatedAt time.Time `json:"-" db:"created_at"`

	// UpdatedAt is a helper struct field for gobuffalo.pop.
	UpdatedAt time.Time `json:"-" db:"updated_at"`

	// CSRFToken contains the anti-csrf token associated with this flow. Only set for browser flows.
	CSRFToken string `json:"-" db:"csrf_token"`

	// Forced stores whether this login flow should enforce re-authentication.
	Forced bool `json:"forced" db:"forced"`
}

func NewFlow(exp time.Duration, csrf string, r *http.Request, flowType flow.Type) *Flow {
	now := time.Now().UTC()
	return &Flow{
		ID:         x.NewUUID(),
		ExpiresAt:  now.Add(exp),
		IssuedAt:   now,
		RequestURL: x.RequestURL(r).String(),
		Methods:    map[identity.CredentialsType]*FlowMethod{},
		CSRFToken:  csrf,
		Type:       flowType,
		Forced:     r.URL.Query().Get("refresh") == "true",
	}
}

func (f *Flow) BeforeSave(_ *pop.Connection) error {
	f.MethodsRaw = make([]FlowMethod, 0, len(f.Methods))
	for _, m := range f.Methods {
		f.MethodsRaw = append(f.MethodsRaw, *m)
	}
	f.Methods = nil
	return nil
}

func (f *Flow) AfterCreate(c *pop.Connection) error {
	return f.AfterFind(c)
}

func (f *Flow) AfterUpdate(c *pop.Connection) error {
	return f.AfterFind(c)
}

func (f *Flow) AfterFind(_ *pop.Connection) error {
	f.Methods = make(FlowMethods)
	for key := range f.MethodsRaw {
		m := f.MethodsRaw[key] // required for pointer dereference
		f.Methods[m.Method] = &m
	}
	f.MethodsRaw = nil
	return nil
}

func (f Flow) TableName(ctx context.Context) string {
	return corp.ContextualizeTableName(ctx, "selfservice_login_flows")
}

func (f *Flow) Valid() error {
	if f.ExpiresAt.Before(time.Now()) {
		return errors.WithStack(NewFlowExpiredError(f.ExpiresAt))
	}
	return nil
}

func (f *Flow) GetID() uuid.UUID {
	return f.ID
}

func (f *Flow) IsForced() bool {
	return f.Forced
}

func (f *Flow) AppendTo(src *url.URL) *url.URL {
	return urlx.CopyWithQuery(src, url.Values{"flow": {f.ID.String()}})
}
