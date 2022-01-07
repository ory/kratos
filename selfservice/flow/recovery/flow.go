package recovery

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"time"

	"github.com/gofrs/uuid"
	"github.com/pkg/errors"

	"github.com/ory/kratos/corp"
	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/ui/container"
	"github.com/ory/kratos/x"
	"github.com/ory/x/sqlxx"
	"github.com/ory/x/urlx"
)

// A Recovery Flow
//
// This request is used when an identity wants to recover their account.
//
// We recommend reading the [Account Recovery Documentation](../self-service/flows/password-reset-account-recovery)
//
// swagger:model selfServiceRecoveryFlow
type Flow struct {
	// ID represents the request's unique ID. When performing the recovery flow, this
	// represents the id in the recovery ui's query parameter: http://<selfservice.flows.recovery.ui_url>?request=<id>
	//
	// required: true
	// type: string
	// format: uuid
	ID uuid.UUID `json:"id" db:"id" faker:"-"`

	// Type represents the flow's type which can be either "api" or "browser", depending on the flow interaction.
	Type flow.Type `json:"type" db:"type" faker:"flow_type"`

	// ExpiresAt is the time (UTC) when the request expires. If the user still wishes to update the setting,
	// a new request has to be initiated.
	//
	// required: true
	ExpiresAt time.Time `json:"expires_at" faker:"time_type" db:"expires_at"`

	// IssuedAt is the time (UTC) when the request occurred.
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
	Active sqlxx.NullString `json:"active,omitempty" faker:"-" db:"active_method"`

	// UI contains data which must be shown in the user interface.
	//
	// required: true
	UI *container.Container `json:"ui" db:"ui"`

	// State represents the state of this request:
	//
	// - choose_method: ask the user to choose a method (e.g. recover account via email)
	// - sent_email: the email has been sent to the user
	// - passed_challenge: the request was successful and the recovery challenge was passed.
	//
	// required: true
	State State `json:"state" faker:"-" db:"state"`

	// CSRFToken contains the anti-csrf token associated with this request.
	CSRFToken string `json:"-" db:"csrf_token"`

	// CreatedAt is a helper struct field for gobuffalo.pop.
	CreatedAt time.Time `json:"-" faker:"-" db:"created_at"`

	// UpdatedAt is a helper struct field for gobuffalo.pop.
	UpdatedAt time.Time `json:"-" faker:"-" db:"updated_at"`

	// RecoveredIdentityID is a helper struct field for gobuffalo.pop.
	RecoveredIdentityID uuid.NullUUID `json:"-" faker:"-" db:"recovered_identity_id"`
	NID                 uuid.UUID     `json:"-"  faker:"-" db:"nid"`
}

func NewFlow(conf *config.Config, exp time.Duration, csrf string, r *http.Request, strategies Strategies, ft flow.Type) (*Flow, error) {
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

	req := &Flow{
		ID:         id,
		ExpiresAt:  now.Add(exp),
		IssuedAt:   now,
		RequestURL: requestURL,
		UI: &container.Container{
			Method: "POST",
			Action: flow.AppendFlowTo(urlx.AppendPaths(conf.SelfPublicURL(), RouteSubmitFlow), id).String(),
		},
		State:     StateChooseMethod,
		CSRFToken: csrf,
		Type:      ft,
	}

	for _, strategy := range strategies {
		if err := strategy.PopulateRecoveryMethod(r, req); err != nil {
			return nil, err
		}
	}

	return req, nil
}

func FromOldFlow(conf *config.Config, exp time.Duration, csrf string, r *http.Request, strategies Strategies, of Flow) (*Flow, error) {
	nf, err := NewFlow(conf, exp, csrf, r, strategies, of.Type)
	if err != nil {
		return nil, err
	}

	nf.RequestURL = of.RequestURL
	return nf, nil
}

func (f *Flow) GetType() flow.Type {
	return f.Type
}

func (f *Flow) GetRequestURL() string {
	return f.RequestURL
}

func (f Flow) TableName(ctx context.Context) string {
	return corp.ContextualizeTableName(ctx, "selfservice_recovery_flows")
}

func (f Flow) GetID() uuid.UUID {
	return f.ID
}

func (f Flow) GetNID() uuid.UUID {
	return f.NID
}

func (f *Flow) Valid() error {
	if f.ExpiresAt.Before(time.Now().UTC()) {
		return errors.WithStack(flow.NewFlowExpiredError(f.ExpiresAt))
	}
	return nil
}

func (f *Flow) AppendTo(src *url.URL) *url.URL {
	return urlx.CopyWithQuery(src, url.Values{"flow": {f.ID.String()}})
}

func (f *Flow) SetCSRFToken(token string) {
	f.CSRFToken = token
	f.UI.SetCSRF(token)
}

func (f Flow) MarshalJSON() ([]byte, error) {
	type local Flow
	if u, err := url.Parse(f.RequestURL); err == nil {
		f.ReturnTo = u.Query().Get("return_to")
	}
	return json.Marshal(local(f))
}
