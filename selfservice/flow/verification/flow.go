package verification

import (
	"context"
	"net/http"
	"net/url"
	"time"

	"github.com/ory/kratos/corp"

	"github.com/gobuffalo/pop/v5"
	"github.com/gofrs/uuid"
	"github.com/pkg/errors"

	"github.com/ory/x/sqlxx"
	"github.com/ory/x/urlx"

	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/selfservice/form"
	"github.com/ory/kratos/text"
	"github.com/ory/kratos/x"
)

// A Verification Flow
//
// Used to verify an out-of-band communication
// channel such as an email address or a phone number.
//
// For more information head over to: https://www.ory.sh/docs/kratos/selfservice/flows/verify-email-account-activation
//
// swagger:model verificationFlow
type Flow struct {
	// ID represents the request's unique ID. When performing the verification flow, this
	// represents the id in the verify ui's query parameter: http://<selfservice.flows.verification.ui_url>?request=<id>
	//
	// type: string
	// format: uuid
	ID uuid.UUID `json:"id" db:"id" faker:"-"`

	// Type represents the flow's type which can be either "api" or "browser", depending on the flow interaction.
	Type flow.Type `json:"type" db:"type" faker:"flow_type"`

	// ExpiresAt is the time (UTC) when the request expires. If the user still wishes to verify the address,
	// a new request has to be initiated.
	ExpiresAt time.Time `json:"expires_at" faker:"time_type" db:"expires_at"`

	// IssuedAt is the time (UTC) when the request occurred.
	IssuedAt time.Time `json:"issued_at" faker:"time_type" db:"issued_at"`

	// RequestURL is the initial URL that was requested from ORY Kratos. It can be used
	// to forward information contained in the URL's path or query for example.
	RequestURL string `json:"request_url" db:"request_url"`

	// Active, if set, contains the registration method that is being used. It is initially
	// not set.
	Active sqlxx.NullString `json:"active,omitempty" faker:"-" db:"active_method"`

	// Messages contains a list of messages to be displayed in the Verification UI. Omitting these
	// messages makes it significantly harder for users to figure out what is going on.
	//
	// More documentation on messages can be found in the [User Interface Documentation](https://www.ory.sh/kratos/docs/concepts/ui-user-interface/).
	Messages text.Messages `json:"messages" faker:"-" db:"messages"`

	// Methods contains context for all account verification methods. If a registration request has been
	// processed, but for example the password is incorrect, this will contain error messages.
	//
	// required: true
	Methods map[string]*FlowMethod `json:"methods" faker:"verification_flow_methods" db:"-"`

	// MethodsRaw is a helper struct field for gobuffalo.pop.
	MethodsRaw []FlowMethod `json:"-" faker:"-" has_many:"selfservice_verification_flow_methods" fk_id:"selfservice_verification_flow_id"`

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
}

func (f Flow) TableName(ctx context.Context) string {
	return corp.ContextualizeTableName(ctx, "selfservice_verification_flows")
}

func NewFlow(exp time.Duration, csrf string, r *http.Request, strategies Strategies, ft flow.Type) (*Flow, error) {
	now := time.Now().UTC()
	f := &Flow{
		ID:         x.NewUUID(),
		ExpiresAt:  now.Add(exp),
		IssuedAt:   now,
		RequestURL: x.RequestURL(r).String(),
		Methods:    map[string]*FlowMethod{},
		CSRFToken:  csrf,
		State:      StateChooseMethod,
		Type:       ft,
	}

	for _, strategy := range strategies {
		if err := strategy.PopulateVerificationMethod(r, f); err != nil {
			return nil, err
		}
	}

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

func (f *Flow) MethodToForm(id string) (form.Form, error) {
	method, ok := f.Methods[id]
	if !ok {
		return nil, errors.WithStack(x.PseudoPanic.WithReasonf("Expected method %s to exist.", id))
	}

	config, ok := method.Config.FlowMethodConfigurator.(form.Form)
	if !ok {
		return nil, errors.WithStack(x.PseudoPanic.WithReasonf(
			"Expected method config %s to be of type *form.HTMLForm but got: %T", id,
			method.Config.FlowMethodConfigurator))
	}

	return config, nil
}

func (f *Flow) BeforeSave(_ *pop.Connection) error {
	f.MethodsRaw = make([]FlowMethod, 0, len(f.Methods))
	for _, m := range f.Methods {
		f.MethodsRaw = append(f.MethodsRaw, *m)
	}
	f.Methods = nil
	return nil
}

func (f *Flow) AfterSave(c *pop.Connection) error {
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
