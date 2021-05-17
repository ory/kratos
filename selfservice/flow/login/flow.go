package login

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

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
	ID  uuid.UUID `json:"id" faker:"-" db:"id" rw:"r"`
	NID uuid.UUID `json:"-"  faker:"-" db:"nid"`

	// Type represents the flow's type which can be either "api" or "browser", depending on the flow interaction.
	//
	// required: true
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

	// RequestURL is the initial URL that was requested from Ory Kratos. It can be used
	// to forward information contained in the URL's path or query for example.
	//
	// required: true
	RequestURL string `json:"request_url" db:"request_url"`

	// The active login method
	//
	// If set contains the login method used. If the flow is new, it is unset.
	Active identity.CredentialsType `json:"active,omitempty" db:"active_method"`

	// UI contains data which must be shown in the user interface.
	//
	// required: true
	UI *container.Container `json:"ui" db:"ui"`

	// CreatedAt is a helper struct field for gobuffalo.pop.
	CreatedAt time.Time `json:"created_at" db:"created_at"`

	// UpdatedAt is a helper struct field for gobuffalo.pop.
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`

	// CSRFToken contains the anti-csrf token associated with this flow. Only set for browser flows.
	CSRFToken string `json:"-" db:"csrf_token"`

	// Forced stores whether this login flow should enforce re-authentication.
	Forced bool `json:"forced" db:"forced"`
}

func NewFlow(conf *config.Config, exp time.Duration, csrf string, r *http.Request, flowType flow.Type) *Flow {
	now := time.Now().UTC()
	id := x.NewUUID()
	return &Flow{
		ID:        id,
		ExpiresAt: now.Add(exp),
		IssuedAt:  now,
		UI: &container.Container{
			Method: "POST",
			Action: flow.AppendFlowTo(urlx.AppendPaths(conf.SelfPublicURL(r), RouteSubmitFlow), id).String(),
		},
		RequestURL: x.RequestURL(r).String(),
		CSRFToken:  csrf,
		Type:       flowType,
		Forced:     r.URL.Query().Get("refresh") == "true",
	}
}

func (f *Flow) GetType() flow.Type {
	return f.Type
}

func (f *Flow) GetRequestURL() string {
	return f.RequestURL
}

func (f Flow) TableName(ctx context.Context) string {
	return corp.ContextualizeTableName(ctx, "selfservice_login_flows")
}

func (f Flow) WhereID(ctx context.Context, alias string) string {
	return fmt.Sprintf("%s.%s = ? AND %s.%s = ?", alias, "id", alias, "nid")
}

func (f *Flow) Valid() error {
	if f.ExpiresAt.Before(time.Now()) {
		return errors.WithStack(NewFlowExpiredError(f.ExpiresAt))
	}
	return nil
}

func (f Flow) GetID() uuid.UUID {
	return f.ID
}

func (f *Flow) IsForced() bool {
	return f.Forced
}

func (f *Flow) AppendTo(src *url.URL) *url.URL {
	return flow.AppendFlowTo(src, f.ID)
}

func (f Flow) GetNID() uuid.UUID {
	return f.NID
}
