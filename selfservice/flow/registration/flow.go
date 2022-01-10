package registration

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"time"

	"github.com/tidwall/gjson"

	"github.com/ory/x/sqlxx"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/ui/container"

	"github.com/ory/kratos/corp"

	"github.com/gofrs/uuid"
	"github.com/pkg/errors"

	"github.com/ory/x/urlx"

	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/x"
)

// swagger:model selfServiceRegistrationFlow
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

	// InternalContext stores internal context used by internals - for example MFA keys.
	InternalContext sqlxx.JSONRawMessage `db:"internal_context" json:"-" faker:"-"`

	// RequestURL is the initial URL that was requested from Ory Kratos. It can be used
	// to forward information contained in the URL's path or query for example.
	//
	// required: true
	RequestURL string `json:"request_url" faker:"url" db:"request_url"`

	// ReturnTo contains the requested return_to URL.
	ReturnTo string `json:"return_to,omitempty" db:"-"`

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
	NID       uuid.UUID `json:"-"  faker:"-" db:"nid"`
}

func NewFlow(conf *config.Config, exp time.Duration, csrf string, r *http.Request, ft flow.Type) (*Flow, error) {
	now := time.Now().UTC()
	id := x.NewUUID()

	// Pre-validate the return to URL which is contained in the HTTP request.
	requestURL := x.RequestURL(r).String()
	_, err := x.SecureRedirectTo(r,
		conf.SelfServiceBrowserDefaultReturnTo(),
		x.SecureRedirectUseSourceURL(requestURL),
		x.SecureRedirectAllowURLs(conf.SelfServiceBrowserWhitelistedReturnToDomains()),
		x.SecureRedirectAllowSelfServiceURLs(conf.SelfPublicURL()),
	)
	if err != nil {
		return nil, err
	}

	return &Flow{
		ID:         id,
		ExpiresAt:  now.Add(exp),
		IssuedAt:   now,
		RequestURL: requestURL,
		UI: &container.Container{
			Method: "POST",
			Action: flow.AppendFlowTo(urlx.AppendPaths(conf.SelfPublicURL(), RouteSubmitFlow), id).String(),
		},
		CSRFToken:       csrf,
		Type:            ft,
		InternalContext: []byte("{}"),
	}, nil
}

func (f Flow) TableName(ctx context.Context) string {
	return corp.ContextualizeTableName(ctx, "selfservice_registration_flows")
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
	if u, err := url.Parse(f.RequestURL); err == nil {
		f.ReturnTo = u.Query().Get("return_to")
	}
	return json.Marshal(local(f))
}
