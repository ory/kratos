// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package settings

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"time"

	"github.com/gobuffalo/pop/v6"

	"github.com/ory/kratos/text"

	"github.com/tidwall/gjson"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/ui/container"
	"github.com/ory/x/urlx"

	"github.com/gofrs/uuid"
	"github.com/pkg/errors"

	"github.com/ory/x/sqlxx"

	"github.com/ory/herodot"

	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/session"
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
	//
	// required: true
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

	// RequestURL is the initial URL that was requested from Ory Kratos. It can be used
	// to forward information contained in the URL's path or query for example.
	//
	// required: true
	RequestURL string `json:"request_url" db:"request_url"`

	// ReturnTo contains the requested return_to URL.
	ReturnTo string `json:"return_to,omitempty" db:"-"`

	// Active, if set, contains the registration method that is being used. It is initially
	// not set.
	Active sqlxx.NullString `json:"active,omitempty" db:"active_method"`

	// UI contains data which must be shown in the user interface.
	//
	// required: true
	UI *container.Container `json:"ui" db:"ui"`

	// Identity contains the identity's data in raw form.
	//
	// If `state` is `success` this will be the updated identity!
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

	// InternalContext stores internal context used by internals - for example MFA keys.
	InternalContext sqlxx.JSONRawMessage `db:"internal_context" json:"-" faker:"-"`

	// IdentityID is a helper struct field for gobuffalo.pop.
	IdentityID uuid.UUID `json:"-" faker:"-" db:"identity_id"`
	// CreatedAt is a helper struct field for gobuffalo.pop.
	CreatedAt time.Time `json:"-" faker:"-" db:"created_at"`
	// UpdatedAt is a helper struct field for gobuffalo.pop.
	UpdatedAt time.Time `json:"-" faker:"-" db:"updated_at"`
	NID       uuid.UUID `json:"-" faker:"-" db:"nid"`

	// Contains a list of actions, that could follow this flow
	//
	// It can, for example, contain a reference to the verification flow, created as part of the user's
	// registration.
	//
	// required: false
	ContinueWithItems []flow.ContinueWith `json:"continue_with,omitempty" db:"-" faker:"-" `
}

var _ flow.Flow = new(Flow)

func MustNewFlow(conf *config.Config, exp time.Duration, r *http.Request, i *identity.Identity, ft flow.Type) *Flow {
	f, err := NewFlow(conf, exp, r, i, ft)
	if err != nil {
		panic(err)
	}
	return f
}

func NewFlow(conf *config.Config, exp time.Duration, r *http.Request, i *identity.Identity, ft flow.Type) (*Flow, error) {
	now := time.Now().UTC()
	id := x.NewUUID()

	// Pre-validate the return to URL which is contained in the HTTP request.
	requestURL := x.RequestURL(r).String()
	_, err := x.SecureRedirectTo(r,
		conf.SelfServiceBrowserDefaultReturnTo(r.Context()),
		x.SecureRedirectUseSourceURL(requestURL),
		x.SecureRedirectAllowURLs(conf.SelfServiceBrowserAllowedReturnToDomains(r.Context())),
		x.SecureRedirectAllowSelfServiceURLs(conf.SelfPublicURL(r.Context())),
	)
	if err != nil {
		return nil, err
	}

	return &Flow{
		ID:         id,
		ExpiresAt:  now.Add(exp),
		IssuedAt:   now,
		RequestURL: requestURL,
		IdentityID: i.ID,
		Identity:   i,
		Type:       ft,
		State:      flow.StateShowForm,
		UI: &container.Container{
			Method: "POST",
			Action: flow.AppendFlowTo(urlx.AppendPaths(conf.SelfPublicURL(r.Context()), RouteSubmitFlow), id).String(),
		},
		InternalContext: []byte("{}"),
	}, nil
}

func (f *Flow) GetType() flow.Type {
	return f.Type
}

func (f *Flow) GetRequestURL() string {
	return f.RequestURL
}

func (f Flow) TableName(ctx context.Context) string {
	return "selfservice_settings_flows"
}

func (f Flow) GetID() uuid.UUID {
	return f.ID
}

func (f Flow) GetNID() uuid.UUID {
	return f.NID
}

func (f *Flow) AppendTo(src *url.URL) *url.URL {
	return flow.AppendFlowTo(src, f.ID)
}

func (f *Flow) Valid(s *session.Session) error {
	if f.ExpiresAt.Before(time.Now().UTC()) {
		return errors.WithStack(flow.NewFlowExpiredError(f.ExpiresAt))
	}

	if f.IdentityID != s.Identity.ID {
		return errors.WithStack(herodot.ErrBadRequest.WithID(text.ErrIDInitiatedBySomeoneElse).WithReasonf(
			"You must restart the flow because the resumable session was initiated by another person."))
	}

	return nil
}

func (f *Flow) EnsureInternalContext() {
	if !gjson.ParseBytes(f.InternalContext).IsObject() {
		f.InternalContext = []byte("{}")
	}
}

func (f Flow) MarshalJSON() ([]byte, error) {
	type local Flow
	f.SetReturnTo()
	return json.Marshal(local(f))
}

func (f *Flow) SetReturnTo() {
	// Return to is already set, do not overwrite it.
	if len(f.ReturnTo) > 0 {
		return
	}
	if u, err := url.Parse(f.RequestURL); err == nil {
		f.ReturnTo = u.Query().Get("return_to")
	}
}

func (f *Flow) AfterFind(*pop.Connection) error {
	f.SetReturnTo()
	return nil
}

func (f *Flow) AfterSave(*pop.Connection) error {
	f.SetReturnTo()
	return nil
}

func (f *Flow) GetUI() *container.Container {
	return f.UI
}

func (f *Flow) AddContinueWith(c flow.ContinueWith) {
	f.ContinueWithItems = append(f.ContinueWithItems, c)
}

func (f *Flow) ContinueWith() []flow.ContinueWith {
	return f.ContinueWithItems
}

func (f *Flow) GetState() State {
	return f.State
}

func (f *Flow) GetFlowName() flow.FlowName {
	return flow.SettingsFlow
}

func (f *Flow) SetState(state State) {
	f.State = state
}
