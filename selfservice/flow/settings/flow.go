package settings

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

	"github.com/ory/x/sqlxx"

	"github.com/ory/herodot"

	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/session"
	"github.com/ory/kratos/text"
	"github.com/ory/kratos/x"
)

// Flow represents a Settings Flow
//
// This flow is used when an identity wants to update settings
// (e.g. profile data, passwords, ...) in a selfservice manner.
//
// We recommend reading the [User Settings Documentation](../self-service/flows/user-settings)
//
// swagger:model settingsFlow
type Flow struct {
	// ID represents the flow's unique ID. When performing the settings flow, this
	// represents the id in the settings ui's query parameter: http://<selfservice.flows.settings.ui_url>?flow=<id>
	//
	// required: true
	// type: string
	// format: uuid
	ID uuid.UUID `json:"id" db:"id" faker:"-"`

	// Type represents the flow's type which can be either "api" or "browser", depending on the flow interaction.
	Type flow.Type `json:"type" db:"type" faker:"flow_type"`

	// ExpiresAt is the time (UTC) when the flow expires. If the user still wishes to update the setting,
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
	RequestURL string `json:"request_url" db:"request_url"`

	// Active, if set, contains the registration method that is being used. It is initially
	// not set.
	Active sqlxx.NullString `json:"active,omitempty" db:"active_method"`

	// Messages contains a list of messages to be displayed in the Settings UI. Omitting these
	// messages makes it significantly harder for users to figure out what is going on.
	//
	// More documentation on messages can be found in the [User Interface Documentation](https://www.ory.sh/kratos/docs/concepts/ui-user-interface/).
	Messages text.Messages `json:"messages" db:"messages" faker:"-"`

	// Methods contains context for all enabled registration methods. If a settings flow has been
	// processed, but for example the first name is empty, this will contain error messages.
	//
	// required: true
	Methods map[string]*FlowMethod `json:"methods" faker:"settings_flow_methods" db:"-"`

	// MethodsRaw is a helper struct field for gobuffalo.pop.
	MethodsRaw []FlowMethod `json:"-" faker:"-" has_many:"selfservice_settings_flow_methods" fk_id:"selfservice_settings_flow_id"`

	// Identity contains all of the identity's data in raw form.
	//
	// required: true
	Identity *identity.Identity `json:"identity" faker:"identity" db:"-" belongs_to:"identities" fk_id:"IdentityID"`

	// State represents the state of this flow. It knows two states:
	//
	// - show_form: No user data has been collected, or it is invalid, and thus the form should be shown.
	// - success: Indicates that the settings flow has been updated successfully with the provided data.
	//	   Done will stay true when repeatedly checking. If set to true, done will revert back to false only
	//	   when a flow with invalid (e.g. "please use a valid phone number") data was sent.
	//
	// required: true
	State State `json:"state" faker:"-" db:"state"`

	// IdentityID is a helper struct field for gobuffalo.pop.
	IdentityID uuid.UUID `json:"-" faker:"-" db:"identity_id"`
	// CreatedAt is a helper struct field for gobuffalo.pop.
	CreatedAt time.Time `json:"-" faker:"-" db:"created_at"`
	// UpdatedAt is a helper struct field for gobuffalo.pop.
	UpdatedAt time.Time `json:"-" faker:"-" db:"updated_at"`
}

// The Response for Settings Flows via API
//
// swagger:model settingsViaApiResponse
type APIFlowResponse struct {
	// The Flow
	//
	// required: true
	Flow *Flow `json:"flow"`

	// The Identity
	//
	// The updated identity
	//
	// required: true
	Identity *identity.Identity `json:"identity"`
}

func NewFlow(exp time.Duration, r *http.Request, i *identity.Identity, ft flow.Type) *Flow {
	now := time.Now().UTC()
	return &Flow{
		ID:         x.NewUUID(),
		ExpiresAt:  now.Add(exp),
		IssuedAt:   now,
		RequestURL: x.RequestURL(r).String(),
		IdentityID: i.ID,
		Identity:   i,
		Type:       ft,
		State:      StateShowForm,
		Methods:    map[string]*FlowMethod{},
	}
}

func (f *Flow) GetType() flow.Type {
	return f.Type
}

func (f *Flow) GetRequestURL() string {
	return f.RequestURL
}

func (r Flow) TableName(ctx context.Context) string {
	return corp.ContextualizeTableName(ctx, "selfservice_settings_flows")
}

func (r *Flow) GetID() uuid.UUID {
	return r.ID
}

func (r *Flow) AppendTo(settingsURL *url.URL) *url.URL {
	return urlx.CopyWithQuery(settingsURL, url.Values{"flow": {r.ID.String()}})
}

func (r *Flow) Valid(s *session.Session) error {
	if r.ExpiresAt.Before(time.Now().UTC()) {
		return errors.WithStack(NewFlowExpiredError(r.ExpiresAt))
	}

	if r.IdentityID != s.Identity.ID {
		return errors.WithStack(herodot.ErrBadRequest.WithReasonf(
			"You must restart the flow because the resumable session was initiated by another person."))
	}

	return nil
}

func (r *Flow) BeforeSave(_ *pop.Connection) error {
	r.MethodsRaw = make([]FlowMethod, 0, len(r.Methods))
	for _, m := range r.Methods {
		r.MethodsRaw = append(r.MethodsRaw, *m)
	}
	r.Methods = nil
	return nil
}

func (r *Flow) AfterSave(c *pop.Connection) error {
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
