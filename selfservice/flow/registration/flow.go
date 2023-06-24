// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package registration

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"time"

	"github.com/gobuffalo/pop/v6"

	"github.com/tidwall/gjson"

	"github.com/ory/x/sqlxx"

	hydraclientgo "github.com/ory/hydra-client-go/v2"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/hydra"
	"github.com/ory/kratos/ui/container"

	"github.com/gofrs/uuid"
	"github.com/pkg/errors"

	"github.com/ory/x/urlx"

	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/x"
)

// swagger:model registrationFlow
type Flow struct {
	// ID represents the flow's unique ID. When performing the registration flow, this
	// represents the id in the registration ui's query parameter: http://<selfservice.flows.registration.ui_url>/?flow=<id>
	//
	// required: true
	ID uuid.UUID `json:"id" faker:"-" db:"id"`

	// Ory OAuth 2.0 Login Challenge.
	//
	// This value is set using the `login_challenge` query parameter of the registration and login endpoints.
	// If set will cooperate with Ory OAuth2 and OpenID to act as an OAuth2 server / OpenID Provider.
	OAuth2LoginChallenge sqlxx.NullString `json:"oauth2_login_challenge,omitempty" faker:"-" db:"oauth2_login_challenge_data"`

	// HydraLoginRequest is an optional field whose presence indicates that Kratos
	// is being used as an identity provider in a Hydra OAuth2 flow. Kratos
	// populates this field by retrieving its value from Hydra and it is used by
	// the login and consent UIs.
	HydraLoginRequest *hydraclientgo.OAuth2LoginRequest `json:"oauth2_login_request,omitempty" faker:"-" db:"-"`

	// Type represents the flow's type which can be either "api" or "browser", depending on the flow interaction.
	//
	// required: true
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

	// InternalContext stores internal context used by internals - for example MFA keys.
	InternalContext sqlxx.JSONRawMessage `db:"internal_context" json:"-" faker:"-"`

	// RequestURL is the initial URL that was requested from Ory Kratos. It can be used
	// to forward information contained in the URL's path or query for example.
	//
	// required: true
	RequestURL string `json:"request_url" faker:"url" db:"request_url"`

	// ReturnTo contains the requested return_to URL.
	ReturnTo string `json:"return_to,omitempty" db:"-"`

	// ReturnToVerification contains the redirect URL for the verification flow.
	ReturnToVerification string `json:"-" db:"-"`

	// Active, if set, contains the registration method that is being used. It is initially
	// not set.
	Active identity.CredentialsType `json:"active,omitempty" faker:"identity_credentials_type" db:"active_method"`

	// UI contains data which must be shown in the user interface.
	//
	// required: true
	UI *container.Container `json:"ui" db:"ui"`

	// CreatedAt is a helper struct field for gobuffalo.pop.
	CreatedAt time.Time `json:"-" faker:"-" db:"created_at"`

	// UpdatedAt is a helper struct field for gobuffalo.pop.
	UpdatedAt time.Time `json:"-" faker:"-" db:"updated_at"`

	// CSRFToken contains the anti-csrf token associated with this flow. Only set for browser flows.
	CSRFToken string    `json:"-" db:"csrf_token"`
	NID       uuid.UUID `json:"-" faker:"-" db:"nid"`

	// TransientPayload is used to pass data from the registration to a webhook
	TransientPayload json.RawMessage `json:"transient_payload,omitempty" faker:"-" db:"-"`

	// Contains a list of actions, that could follow this flow
	//
	// It can, for example, contain a reference to the verification flow, created as part of the user's
	// registration.
	ContinueWithItems []flow.ContinueWith `json:"-" db:"-" faker:"-" `

	// SessionTokenExchangeCode holds the secret code that the client can use to retrieve a session token after the flow has been completed.
	// This is only set if the client has requested a session token exchange code, and if the flow is of type "api",
	// and only on creating the flow.
	SessionTokenExchangeCode string `json:"session_token_exchange_code,omitempty" faker:"-" db:"-"`
}

func NewFlow(conf *config.Config, exp time.Duration, csrf string, r *http.Request, ft flow.Type) (*Flow, error) {
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

	hlc, err := hydra.GetLoginChallengeID(conf, r)
	if err != nil {
		return nil, err
	}

	return &Flow{
		ID:                   id,
		OAuth2LoginChallenge: hlc,
		ExpiresAt:            now.Add(exp),
		IssuedAt:             now,
		RequestURL:           requestURL,
		UI: &container.Container{
			Method: "POST",
			Action: flow.AppendFlowTo(urlx.AppendPaths(conf.SelfPublicURL(r.Context()), RouteSubmitFlow), id).String(),
		},
		CSRFToken:       csrf,
		Type:            ft,
		InternalContext: []byte("{}"),
	}, nil
}

func (f Flow) TableName(ctx context.Context) string {
	return "selfservice_registration_flows"
}

func (f Flow) GetID() uuid.UUID {
	return f.ID
}

func (f Flow) GetNID() uuid.UUID {
	return f.NID
}

func (f *Flow) Valid() error {
	if f.ExpiresAt.Before(time.Now()) {
		return errors.WithStack(flow.NewFlowExpiredError(f.ExpiresAt))
	}
	return nil
}

func (f *Flow) AppendTo(src *url.URL) *url.URL {
	return flow.AppendFlowTo(src, f.ID)
}

func (f *Flow) GetType() flow.Type {
	return f.Type
}

func (f *Flow) GetRequestURL() string {
	return f.RequestURL
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

func (f *Flow) SecureRedirectToOpts(ctx context.Context, cfg config.Provider) (opts []x.SecureRedirectOption) {
	return []x.SecureRedirectOption{
		x.SecureRedirectReturnTo(f.ReturnTo),
		x.SecureRedirectUseSourceURL(f.RequestURL),
		x.SecureRedirectAllowURLs(cfg.Config().SelfServiceBrowserAllowedReturnToDomains(ctx)),
		x.SecureRedirectAllowSelfServiceURLs(cfg.Config().SelfPublicURL(ctx)),
		x.SecureRedirectOverrideDefaultReturnTo(cfg.Config().SelfServiceFlowRegistrationReturnTo(ctx, f.Active.String())),
	}
}
