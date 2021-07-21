package verification

import (
	"context"
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

// A Verification Flow
//
// Used to verify an out-of-band communication
// channel such as an email address or a phone number.
//
// For more information head over to: https://www.ory.sh/docs/kratos/selfservice/flows/verify-email-account-activation
//
// swagger:model selfServiceVerificationFlow
type Flow struct {
	// ID represents the request's unique ID. When performing the verification flow, this
	// represents the id in the verify ui's query parameter: http://<selfservice.flows.verification.ui_url>?request=<id>
	//
	// type: string
	// format: uuid
	// required: true
	ID uuid.UUID `json:"id" db:"id" faker:"-"`

	// Type represents the flow's type which can be either "api" or "browser", depending on the flow interaction.
	// required: true
	Type flow.Type `json:"type" db:"type" faker:"flow_type"`

	// ExpiresAt is the time (UTC) when the request expires. If the user still wishes to verify the address,
	// a new request has to be initiated.
	ExpiresAt time.Time `json:"expires_at" faker:"time_type" db:"expires_at"`

	// IssuedAt is the time (UTC) when the request occurred.
	IssuedAt time.Time `json:"issued_at" faker:"time_type" db:"issued_at"`

	// RequestURL is the initial URL that was requested from Ory Kratos. It can be used
	// to forward information contained in the URL's path or query for example.
	RequestURL string `json:"request_url" db:"request_url"`

	// Active, if set, contains the registration method that is being used. It is initially
	// not set.
	Active sqlxx.NullString `json:"active,omitempty" faker:"-" db:"active_method"`

	// UI contains data which must be shown in the user interface.
	//
	// required: true
	UI *container.Container `json:"ui" db:"ui"`

	// State represents the state of this request:
	//
	// - choose_method: ask the user to choose a method (e.g. verify your email)
	// - sent_email: the email has been sent to the user
	// - passed_challenge: the request was successful and the verification challenge was passed.
	//
	// required: true
	State State `json:"state" faker:"-" db:"state"`

	// CSRFToken contains the anti-csrf token associated with this request.
	CSRFToken string `json:"-" db:"csrf_token"`

	// CreatedAt is a helper struct field for gobuffalo.pop.
	CreatedAt time.Time `json:"-" faker:"-" db:"created_at"`
	// UpdatedAt is a helper struct field for gobuffalo.pop.
	UpdatedAt time.Time `json:"-" faker:"-" db:"updated_at"`
	NID       uuid.UUID `json:"-"  faker:"-" db:"nid"`
}

func (f *Flow) GetType() flow.Type {
	return f.Type
}

func (f *Flow) GetRequestURL() string {
	return f.RequestURL
}

func (f Flow) TableName(ctx context.Context) string {
	return corp.ContextualizeTableName(ctx, "selfservice_verification_flows")
}

func NewFlow(conf *config.Config, exp time.Duration, csrf string, r *http.Request, strategies Strategies, ft flow.Type) (*Flow, error) {
	now := time.Now().UTC()
	id := x.NewUUID()
	f := &Flow{
		ID:        id,
		ExpiresAt: now.Add(exp), IssuedAt: now,
		RequestURL: x.RequestURL(r).String(),
		UI: &container.Container{
			Method: "POST",
			Action: flow.AppendFlowTo(urlx.AppendPaths(conf.SelfPublicURL(r), RouteSubmitFlow), id).String(),
		},
		CSRFToken: csrf,
		State:     StateChooseMethod,
		Type:      ft,
	}

	for _, strategy := range strategies {
		if err := strategy.PopulateVerificationMethod(r, f); err != nil {
			return nil, err
		}
	}

	return f, nil
}

func NewPostHookFlow(conf *config.Config, exp time.Duration, csrf string, r *http.Request, strategies Strategies, original flow.Flow) (*Flow, error) {
	f, err := NewFlow(conf, exp, csrf, r, strategies, original.GetType())
	if err != nil {
		return nil, err
	}
	requestURL, err := url.ParseRequestURI(original.GetRequestURL())
	if err != nil {
		requestURL = new(url.URL)
	}
	query := requestURL.Query()
	query.Set("return_to", query.Get("after_verification_return_to"))
	query.Del("after_verification_return_to")
	requestURL.RawQuery = query.Encode()
	f.RequestURL = requestURL.String()
	return f, nil
}

func (f *Flow) Valid() error {
	if f.ExpiresAt.Before(time.Now()) {
		return errors.WithStack(NewFlowExpiredError(f.ExpiresAt))
	}
	return nil
}

func (f *Flow) AppendTo(src *url.URL) *url.URL {
	return urlx.CopyWithQuery(src, url.Values{"flow": {f.ID.String()}})
}

func (f Flow) GetID() uuid.UUID {
	return f.ID
}

func (f Flow) GetNID() uuid.UUID {
	return f.NID
}
