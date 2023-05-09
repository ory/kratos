// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package verification

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"time"

	"github.com/gobuffalo/pop/v6"

	"github.com/gofrs/uuid"
	"github.com/pkg/errors"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/ui/container"
	"github.com/ory/kratos/x"
	"github.com/ory/x/sqlxx"
	"github.com/ory/x/urlx"
)

var _ flow.Flow = new(Flow)

// A Verification Flow
//
// Used to verify an out-of-band communication
// channel such as an email address or a phone number.
//
// For more information head over to: https://www.ory.sh/docs/kratos/self-service/flows/verify-email-account-activation
//
// swagger:model verificationFlow
type Flow struct {
	// ID represents the request's unique ID. When performing the verification flow, this
	// represents the id in the verify ui's query parameter: http://<selfservice.flows.verification.ui_url>?request=<id>
	//
	// type: string
	// format: uuid
	// required: true
	ID uuid.UUID `json:"id" db:"id" faker:"-"`

	// Type represents the flow's type which can be either "api" or "browser", depending on the flow interaction.
	//
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
	return "selfservice_verification_flows"
}

func NewFlow(conf *config.Config, exp time.Duration, csrf string, r *http.Request, strategy Strategy, ft flow.Type) (*Flow, error) {
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

	f := &Flow{
		ID:         id,
		ExpiresAt:  now.Add(exp),
		IssuedAt:   now,
		RequestURL: requestURL,
		UI: &container.Container{
			Method: "POST",
			Action: flow.AppendFlowTo(urlx.AppendPaths(conf.SelfPublicURL(r.Context()), RouteSubmitFlow), id).String(),
		},
		CSRFToken: csrf,
		State:     StateChooseMethod,
		Type:      ft,
	}

	if strategy != nil {
		f.Active = sqlxx.NullString(strategy.VerificationNodeGroup())
		if err := strategy.PopulateVerificationMethod(r, f); err != nil {
			return nil, err
		}
	}

	return f, nil
}

func FromOldFlow(conf *config.Config, exp time.Duration, csrf string, r *http.Request, strategy Strategy, of *Flow) (*Flow, error) {
	f := of.Type
	// Using the same flow in the recovery/verification context can lead to using API flow in a verification/recovery email
	if of.Type == flow.TypeAPI {
		f = flow.TypeBrowser
	}
	nf, err := NewFlow(conf, exp, csrf, r, strategy, f)
	if err != nil {
		return nil, err
	}

	nf.RequestURL = of.RequestURL
	return nf, nil
}

func NewPostHookFlow(conf *config.Config, exp time.Duration, csrf string, r *http.Request, strategy Strategy, original flow.Flow) (*Flow, error) {
	f, err := NewFlow(conf, exp, csrf, r, strategy, original.GetType())
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
		return errors.WithStack(flow.NewFlowExpiredError(f.ExpiresAt))
	}
	return nil
}

func (f *Flow) AppendTo(src *url.URL) *url.URL {
	values := src.Query()
	values.Set("flow", f.ID.String())
	return urlx.CopyWithQuery(src, values)
}

func (f Flow) GetID() uuid.UUID {
	return f.ID
}

func (f Flow) GetNID() uuid.UUID {
	return f.NID
}

func (f *Flow) SetCSRFToken(token string) {
	f.CSRFToken = token
	f.UI.SetCSRF(token)
}

func (f Flow) MarshalJSON() ([]byte, error) {
	type local Flow
	f.SetReturnTo()
	return json.Marshal(local(f))
}

func (f *Flow) SetReturnTo() {
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

// ContinueURL generates the URL to show on the continue screen after succesful verification
//
// It follows the following precedence:
//  1. If a `return_to` parameter has been passed to the flow's creation, is a valid URL and it's in the `selfservice.allowed_return_urls` that URL is returned
//  2. If `selfservice.flows.verification.after` is set, that URL is returned
//  3. As a fallback, the `selfservice.default_browser_return_url` URL is returned
func (f *Flow) ContinueURL(ctx context.Context, config *config.Config) *url.URL {
	flowContinueURL := config.SelfServiceFlowVerificationReturnTo(ctx, config.SelfServiceBrowserDefaultReturnTo(ctx))

	// Parse the flows request URL
	verificationRequestURL, err := urlx.Parse(f.GetRequestURL())
	if err != nil {
		// Return flow default, or global default return URL
		return flowContinueURL
	}

	verificationRequest := http.Request{URL: verificationRequestURL}

	returnTo, err := x.SecureRedirectTo(&verificationRequest, flowContinueURL,
		x.SecureRedirectAllowSelfServiceURLs(config.SelfPublicURL(ctx)),
		x.SecureRedirectAllowURLs(config.SelfServiceBrowserAllowedReturnToDomains(ctx)),
	)
	if err != nil {
		// an error occured return flow default, or global default return URL
		return flowContinueURL
	}
	return returnTo
}
