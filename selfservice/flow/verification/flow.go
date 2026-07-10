// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package verification

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"time"

	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"

	hydraclientgo "github.com/ory/hydra-client-go/v2"

	"github.com/ory/kratos/x/redir"

	"github.com/ory/pop/v6"

	"github.com/gofrs/uuid"
	"github.com/pkg/errors"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/session"
	"github.com/ory/kratos/ui/container"
	"github.com/ory/kratos/x"
	"github.com/ory/x/clock"
	"github.com/ory/x/sqlxx"
	"github.com/ory/x/urlx"
)

// A Verification Flow
//
// Used to verify an out-of-band communication
// channel such as an email address or a phone number.
//
// For more information head over to: https://www.ory.com/docs/kratos/self-service/flows/verify-email-account-activation
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

	// InternalContext stores internal-only data for this flow.
	InternalContext sqlxx.JSONRawMessage `db:"internal_context" json:"-" faker:"-"`

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

	// OAuth2LoginChallenge holds the login challenge originally set during the registration flow.
	OAuth2LoginChallenge sqlxx.NullString `json:"-" db:"oauth2_login_challenge"`
	OAuth2LoginChallengeParams

	// HydraLoginRequest is the OAuth2 login request behind OAuth2LoginChallenge. It is
	// hydrated on demand when a message is sent for this flow so that courier templates
	// can brand messages per OAuth2 client. It is neither persisted nor exposed via the API.
	HydraLoginRequest *hydraclientgo.OAuth2LoginRequest `json:"-" faker:"-" db:"-"`

	// CSRFToken contains the anti-csrf token associated with this request.
	CSRFToken string `json:"-" db:"csrf_token"`

	// CreatedAt is a helper struct field for gobuffalo.pop.
	CreatedAt time.Time `json:"-" faker:"-" db:"created_at"`
	// UpdatedAt is a helper struct field for gobuffalo.pop.
	UpdatedAt time.Time `json:"-" faker:"-" db:"updated_at"`
	NID       uuid.UUID `json:"-"  faker:"-" db:"nid"`

	// TransientPayload is used to pass data from the verification flow to hooks and email templates
	//
	// required: false
	TransientPayload json.RawMessage `json:"transient_payload,omitempty" faker:"-" db:"-"`
}

type OAuth2LoginChallengeParams struct {
	// SessionID holds the session id if set from a registraton hook.
	SessionID uuid.NullUUID `json:"-" faker:"-" db:"session_id"`

	// IdentityID holds the identity id if set from a registraton hook.
	IdentityID uuid.NullUUID `json:"-" faker:"-" db:"identity_id"`

	// AMR contains a list of authentication methods that were used to verify the
	// session if set from a registration hook.
	AMR session.AuthenticationMethods `db:"authentication_methods" json:"-"`
}

var _ flow.Flow = (*Flow)(nil)

// flowDependencies are the dependencies NewFlow needs to construct a
// verification flow: the configuration (for return-to validation) and the
// clock. The lifespan and CSRF token are passed explicitly because callers
// vary them (regenerated CSRF tokens, conditional tokens by flow type).
type flowDependencies interface {
	config.Provider
	clock.Provider
}

func NewFlow(reg flowDependencies, exp time.Duration, csrf string, r *http.Request, strategies Strategies, ft flow.Type) (*Flow, error) {
	conf := reg.Config()
	now := reg.Clock().Now().UTC()
	id := x.NewUUID()

	// Pre-validate the return to URL which is contained in the HTTP request.
	requestURL := x.RequestURL(r).String()
	_, err := redir.SecureRedirectTo(r,
		conf.SelfServiceBrowserDefaultReturnTo(r.Context()),
		redir.SecureRedirectUseSourceURL(requestURL),
		redir.SecureRedirectAllowURLs(conf.SelfServiceBrowserAllowedReturnToDomains(r.Context())),
		redir.SecureRedirectAllowSelfServiceURLs(conf.SelfPublicURL(r.Context())),
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
		State:     flow.StateChooseMethod,
		Type:      ft,
	}
	if err := f.setCourierBaseURL(x.BaseURLStringFromContext(r.Context())); err != nil {
		return nil, err
	}

	for _, strategy := range strategies {
		if ps, isPrimary := strategy.(PrimaryStrategy); isPrimary {
			f.Active = sqlxx.NullString(ps.NodeGroup())
		}
		if err := strategy.PopulateVerificationMethod(r, f); err != nil {
			return nil, err
		}
	}

	return f, nil
}

func FromOldFlow(reg flowDependencies, exp time.Duration, csrf string, r *http.Request, strategies Strategies, of *Flow) (*Flow, error) {
	f := of.Type
	// Using the same flow in the recovery/verification context can lead to using API flow in a verification/recovery email
	if of.Type == flow.TypeAPI {
		f = flow.TypeBrowser
	}
	nf, err := NewFlow(reg, exp, csrf, r, strategies, f)
	if err != nil {
		return nil, err
	}

	nf.RequestURL = of.RequestURL
	return nf, nil
}

func NewPostHookFlow(reg flowDependencies, exp time.Duration, csrf string, r *http.Request, strategies Strategies, original flow.Flow) (*Flow, error) {
	f, err := NewFlow(reg, exp, csrf, r, strategies, original.GetType())
	if err != nil {
		return nil, err
	}
	f.TransientPayload = original.GetTransientPayload()
	requestURL, err := url.ParseRequestURI(original.GetRequestURL())
	if err != nil {
		requestURL = new(url.URL)
	}
	query := requestURL.Query()
	// we need to keep the return_to in-tact if the `after_verification_return_to` is empty
	// otherwise we take the `after_verification_return_to` query parameter over the current `return_to`
	if afterVerificationReturn := query.Get("after_verification_return_to"); afterVerificationReturn != "" {
		query.Set("return_to", afterVerificationReturn)
	}
	query.Del("after_verification_return_to")
	requestURL.RawQuery = query.Encode()
	f.RequestURL = requestURL.String()
	if t, ok := original.(flow.OAuth2ChallengeProvider); ok {
		f.OAuth2LoginChallenge = t.GetOAuth2LoginChallenge()
	}
	return f, nil
}

func (f *Flow) GetType() flow.Type                        { return f.Type }
func (f *Flow) GetRequestURL() string                     { return f.RequestURL }
func (Flow) TableName() string                            { return "selfservice_verification_flows" }
func (f Flow) GetID() uuid.UUID                           { return f.ID }
func (f *Flow) GetState() State                           { return f.State }
func (Flow) GetFlowName() flow.FlowName                   { return flow.VerificationFlow }
func (f *Flow) SetState(state State)                      { f.State = state }
func (f *Flow) GetTransientPayload() json.RawMessage      { return f.TransientPayload }
func (f *Flow) GetOAuth2LoginChallenge() sqlxx.NullString { return f.OAuth2LoginChallenge }
func (f *Flow) GetHydraLoginRequest() *hydraclientgo.OAuth2LoginRequest {
	return f.HydraLoginRequest
}
func (f *Flow) GetUI() *container.Container               { return f.UI }
func (f *Flow) GetInternalContext() sqlxx.JSONRawMessage  { return f.InternalContext }
func (f *Flow) SetInternalContext(c sqlxx.JSONRawMessage) { f.InternalContext = c }

// EnsureInternalContext initializes InternalContext to an empty JSON object
// if it is missing or not valid JSON. Mirrors the registration / settings /
// login flow implementations.
func (f *Flow) EnsureInternalContext() {
	if !gjson.ValidBytes(f.InternalContext) {
		f.InternalContext = []byte("{}")
	}
}

// GetCourierBaseURL returns the base URL captured at flow init from the
// request context, or the empty string when nothing was captured (the email
// senders then fall back to Config.SelfPublicURL).
func (f *Flow) GetCourierBaseURL() string {
	return gjson.GetBytes(f.InternalContext, flow.InternalContextKeyCourierBaseURL).String()
}

// setCourierBaseURL writes the captured base URL into InternalContext under
// the well-known key. Empty input is a no-op (preserves the fall-back
// path). Inputs longer than 8192 bytes are rejected — the same implicit
// ceiling the dedicated VARCHAR(8192) column used to enforce — so a
// pathological header cannot bloat the row.
func (f *Flow) setCourierBaseURL(s string) error {
	if s == "" {
		return nil
	}
	if len(s) > 8192 {
		return nil
	}
	f.EnsureInternalContext()
	out, err := sjson.SetBytes(f.InternalContext, flow.InternalContextKeyCourierBaseURL, s)
	if err != nil {
		return errors.WithStack(err)
	}
	f.InternalContext = out
	return nil
}

func (f *Flow) Valid(c clock.Clock) error {
	if f.ExpiresAt.Before(c.Now()) {
		return errors.WithStack(flow.NewFlowExpiredError(c, f.ExpiresAt))
	}
	return nil
}

func (f *Flow) AppendTo(src *url.URL) *url.URL {
	values := src.Query()
	values.Set("flow", f.ID.String())
	return urlx.CopyWithQuery(src, values)
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

	returnTo, err := redir.SecureRedirectTo(&verificationRequest, flowContinueURL,
		redir.SecureRedirectAllowSelfServiceURLs(config.SelfPublicURL(ctx)),
		redir.SecureRedirectAllowURLs(config.SelfServiceBrowserAllowedReturnToDomains(ctx)),
	)
	if err != nil {
		// an error occured return flow default, or global default return URL
		return flowContinueURL
	}
	return returnTo
}

func (f *Flow) ToLoggerField() map[string]any {
	if f == nil {
		return map[string]any{}
	}
	return map[string]any{
		"id":          f.ID.String(),
		"return_to":   f.ReturnTo,
		"request_url": f.RequestURL,
		"active":      f.Active,
		"Type":        f.Type,
		"nid":         f.NID,
		"state":       f.State,
	}
}
