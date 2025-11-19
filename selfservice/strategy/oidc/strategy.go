// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package oidc

import (
	"bytes"
	"cmp"
	"context"
	"encoding/json"
	"maps"
	"net/http"
	"net/url"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/ory/kratos/x/nosurfx"
	"github.com/ory/kratos/x/redir"

	"github.com/gofrs/uuid"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/oauth2"

	"github.com/ory/herodot"
	"github.com/ory/kratos/cipher"
	"github.com/ory/kratos/continuity"
	"github.com/ory/kratos/driver/config"
	oidcv1 "github.com/ory/kratos/gen/oidc/v1"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/schema"
	"github.com/ory/kratos/selfservice/errorx"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/selfservice/flow/login"
	"github.com/ory/kratos/selfservice/flow/registration"
	"github.com/ory/kratos/selfservice/flow/settings"
	"github.com/ory/kratos/selfservice/sessiontokenexchange"
	"github.com/ory/kratos/selfservice/strategy"
	"github.com/ory/kratos/session"
	"github.com/ory/kratos/text"
	"github.com/ory/kratos/ui/container"
	"github.com/ory/kratos/ui/node"
	"github.com/ory/kratos/x"
	"github.com/ory/x/decoderx"
	"github.com/ory/x/jsonnetsecure"
	"github.com/ory/x/otelx"
	"github.com/ory/x/otelx/semconv"
	"github.com/ory/x/sqlxx"
	"github.com/ory/x/stringsx"
	"github.com/ory/x/urlx"
)

const (
	RouteBase = "/self-service/methods/oidc"

	RouteAuth                 = RouteBase + "/auth/{flow}"
	RouteCallback             = RouteBase + "/callback/{provider}"
	RouteCallbackGeneric      = RouteBase + "/callback"
	RouteOrganizationCallback = RouteBase + "/organization/{organization}/callback/{provider}"
)

var _ identity.ActiveCredentialsCounter = new(Strategy)

type Dependencies interface {
	errorx.ManagementProvider

	config.Provider

	x.LoggingProvider
	x.CookieProvider
	nosurfx.CSRFProvider
	nosurfx.CSRFTokenGeneratorProvider
	x.WriterProvider
	x.HTTPClientProvider
	x.TracingProvider

	identity.ValidationProvider
	identity.PrivilegedPoolProvider
	identity.ActiveCredentialsCounterStrategyProvider
	identity.ManagementProvider

	session.ManagementProvider
	session.HandlerProvider
	sessiontokenexchange.PersistenceProvider

	login.HookExecutorProvider
	login.FlowPersistenceProvider
	login.HooksProvider
	login.StrategyProvider
	login.HandlerProvider
	login.ErrorHandlerProvider

	registration.HookExecutorProvider
	registration.FlowPersistenceProvider
	registration.HooksProvider
	registration.StrategyProvider
	registration.HandlerProvider
	registration.ErrorHandlerProvider

	settings.ErrorHandlerProvider
	settings.FlowPersistenceProvider
	settings.HookExecutorProvider

	continuity.ManagementProvider

	cipher.Provider

	jsonnetsecure.VMProvider
}

func isForced(req interface{}) bool {
	f, ok := req.(interface {
		IsRefresh() bool
	})
	return ok && f.IsRefresh()
}

// ConflictingIdentityVerdict encodes the decision on what to do on a oconflict
// between an existing and a new identity.
type ConflictingIdentityVerdict int

const (
	// ConflictingIdentityVerdictUnknown is the default value and should not be used.
	ConflictingIdentityVerdictUnknown ConflictingIdentityVerdict = iota

	// ConflictingIdentityVerdictReject rejects the new identity. The flow will
	// continue with an explicit account linking step, where the user will need to
	// confirm an existing credential on the identity.
	ConflictingIdentityVerdictReject

	// ConflictingIdentityVerdictMerge merges the new identity into the existing.
	ConflictingIdentityVerdictMerge
)

// Strategy implements selfservice.LoginStrategy, selfservice.RegistrationStrategy and selfservice.SettingsStrategy.
// It supports login, registration and settings via OpenID Providers.
type Strategy struct {
	d                           Dependencies
	validator                   *schema.Validator
	dec                         *decoderx.HTTP
	credType                    identity.CredentialsType
	handleUnknownProviderError  func(err error) error
	handleMethodNotAllowedError func(err error) error

	conflictingIdentityPolicy ConflictingIdentityPolicy
}
type ConflictingIdentityPolicy func(ctx context.Context, existingIdentity, newIdentity *identity.Identity, provider Provider, claims *Claims) ConflictingIdentityVerdict

type AuthCodeContainer struct {
	FlowID           string              `json:"flow_id"`
	State            string              `json:"state"`
	IdentitySchema   flow.IdentitySchema `json:"identity_schema_id,omitempty"`
	Traits           json.RawMessage     `json:"traits"`
	TransientPayload json.RawMessage     `json:"transient_payload"`
}

func (s *Strategy) CountActiveFirstFactorCredentials(ctx context.Context, cc map[identity.CredentialsType]identity.Credentials) (count int, err error) {
	return CountActiveFirstFactorCredentials(ctx, s.ID(), cc, false)
}

func CountActiveFirstFactorCredentials(_ context.Context, id identity.CredentialsType, cc map[identity.CredentialsType]identity.Credentials, withOrgs bool) (count int, err error) {
	for _, c := range cc {
		if c.Type == id && gjson.ValidBytes(c.Config) {
			var conf identity.CredentialsOIDC
			if err = json.Unmarshal(c.Config, &conf); err != nil {
				return 0, errors.WithStack(err)
			}

			for _, identifier := range c.Identifiers {
				provider, sub, ok := strings.Cut(identifier, ":")
				if !ok {
					continue
				}

				for _, prov := range conf.Providers {
					if withOrgs && len(prov.Organization) == 0 {
						continue
					} else if !withOrgs && len(prov.Organization) > 0 {
						continue
					}

					if provider == prov.Provider && sub == prov.Subject && prov.Subject != "" && prov.Provider != "" {
						count++
					}
				}
			}
		}
	}
	return
}

func (s *Strategy) CountActiveMultiFactorCredentials(_ context.Context, _ map[identity.CredentialsType]identity.Credentials) (count int, err error) {
	return 0, nil
}

func (s *Strategy) setRoutes(r *x.RouterPublic) {
	wrappedHandleCallback := strategy.IsDisabled(s.d, s.ID().String(), s.HandleCallback)
	if !r.HasRoute("GET", RouteCallback) {
		r.GET(RouteCallback, wrappedHandleCallback)
	}
	if !r.HasRoute("GET", RouteCallbackGeneric) {
		r.GET(RouteCallbackGeneric, wrappedHandleCallback)
	}

	// Apple can use the POST request method when calling the callback
	if !r.HasRoute("POST", RouteCallback) {
		// Apple is the only (known) provider that sometimes does a form POST to the callback URL.
		// This is a workaround to handle this case.
		// But since the URL contains the `id` of the provider, we just allow all OIDC provider callbacks to bypass CSRF.
		// This is fine, because all other providers seem to use GET, which is CSRF safe.
		s.d.CSRFHandler().IgnoreGlob(RouteBase + "/callback/*")

		// When handler is called using POST method, the cookies are not attached to the request
		// by the browser. So here we just redirect the request to the same location rewriting the
		// form fields to query params. This second GET request should have the cookies attached.
		r.POST(RouteCallback, s.redirectToGET)
	}
}

// Redirect POST request to GET rewriting form fields to query params.
func (s *Strategy) redirectToGET(w http.ResponseWriter, r *http.Request) {
	publicUrl := s.d.Config().SelfPublicURL(r.Context())
	dest := *r.URL
	dest.Host = publicUrl.Host
	dest.Scheme = publicUrl.Scheme
	if err := r.ParseForm(); err == nil {
		q := dest.Query()
		for key, values := range r.Form {
			for _, value := range values {
				q.Set(key, value)
			}
		}
		dest.RawQuery = q.Encode()
	}
	dest.Path = filepath.Join(publicUrl.Path, dest.Path)

	http.Redirect(w, r, dest.String(), http.StatusFound)
}

type NewStrategyOpt func(s *Strategy)

// ForCredentialType overrides the credentials type for this strategy.
func ForCredentialType(ct identity.CredentialsType) NewStrategyOpt {
	return func(s *Strategy) { s.credType = ct }
}

// WithUnknownProviderHandler overrides the error returned when the provider
// cannot be found.
func WithUnknownProviderHandler(handler func(error) error) NewStrategyOpt {
	return func(s *Strategy) { s.handleUnknownProviderError = handler }
}

// WithHandleMethodNotAllowedError overrides the error returned when method is
// not allowed.
func WithHandleMethodNotAllowedError(handler func(error) error) NewStrategyOpt {
	return func(s *Strategy) { s.handleMethodNotAllowedError = handler }
}

// WithOnConflictingIdentity sets a policy handler for deciding what to do when a
// new identity conflicts with an existing one during login.
func WithOnConflictingIdentity(handler ConflictingIdentityPolicy) NewStrategyOpt {
	return func(s *Strategy) { s.conflictingIdentityPolicy = handler }
}

// SetOnConflictingIdentity sets a policy handler for deciding what to do when a
// new identity conflicts with an existing one during login. This should only be
// called in tests.
func (s *Strategy) SetOnConflictingIdentity(t testing.TB, handler ConflictingIdentityPolicy) {
	if t == nil {
		panic("this should only be called in tests")
	}
	s.conflictingIdentityPolicy = handler
}

func NewStrategy(d Dependencies, opts ...NewStrategyOpt) *Strategy {
	s := &Strategy{
		d:                           d,
		validator:                   schema.NewValidator(),
		credType:                    identity.CredentialsTypeOIDC,
		handleUnknownProviderError:  func(err error) error { return err },
		handleMethodNotAllowedError: func(err error) error { return err },
	}

	for _, opt := range opts {
		opt(s)
	}

	return s
}

func (s *Strategy) ID() identity.CredentialsType {
	return s.credType
}

func (s *Strategy) validateFlow(ctx context.Context, r *http.Request, rid uuid.UUID) (flow.Flow, error) {
	if rid.IsNil() {
		return nil, errors.WithStack(herodot.ErrBadRequest.WithReason("The session cookie contains invalid values and the flow could not be executed. Please try again."))
	}

	if ar, err := s.d.RegistrationFlowPersister().GetRegistrationFlow(ctx, rid); err == nil {
		if err := ar.Valid(); err != nil {
			return ar, err
		}
		return ar, nil
	}

	if ar, err := s.d.LoginFlowPersister().GetLoginFlow(ctx, rid); err == nil {
		if err := ar.Valid(); err != nil {
			return ar, err
		}
		return ar, nil
	}

	ar, err := s.d.SettingsFlowPersister().GetSettingsFlow(ctx, rid)
	if err == nil {
		sess, err := s.d.SessionManager().FetchFromRequest(ctx, r)
		if err != nil {
			return ar, err
		}

		if err := ar.Valid(sess); err != nil {
			return ar, err
		}
		return ar, nil
	}

	return ar, err // this must return the error
}

func (s *Strategy) ValidateCallback(w http.ResponseWriter, r *http.Request) (flow.Flow, *oidcv1.State, *AuthCodeContainer, error) {
	var (
		codeParam  = stringsx.Coalesce(r.URL.Query().Get("code"), r.URL.Query().Get("authCode"))
		stateParam = r.URL.Query().Get("state")
		errorParam = r.URL.Query().Get("error")
	)

	if stateParam == "" {
		return nil, nil, nil, errors.WithStack(herodot.ErrBadRequest.WithReasonf(`Unable to complete OpenID Connect flow because the OpenID Provider did not return the state query parameter.`))
	}
	state, err := DecryptState(r.Context(), s.d.Cipher(r.Context()), stateParam)
	if err != nil {
		return nil, nil, nil, errors.WithStack(herodot.ErrBadRequest.WithReasonf(`Unable to complete OpenID Connect flow because the state parameter is invalid.`))
	}

	if providerFromURL := r.PathValue("provider"); providerFromURL != "" {
		// We're serving an OIDC callback URL with provider in the URL.
		if state.ProviderId == "" {
			// provider in URL, but not in state: compatiblity mode, remove this fallback later
			state.ProviderId = providerFromURL
		} else if state.ProviderId != providerFromURL {
			// provider in state, but URL with different provider -> something's fishy
			return nil, nil, nil, errors.WithStack(herodot.ErrBadRequest.WithReasonf(`Unable to complete OpenID Connect flow: provider mismatch between internal state and URL.`))
		}
	}
	if state.ProviderId == "" {
		// weird: provider neither in the state nor in the URL
		return nil, nil, nil, errors.WithStack(herodot.ErrBadRequest.WithReasonf(`Unable to complete OpenID Connect flow: provider could not be retrieved from state nor URL.`))
	}

	f, err := s.validateFlow(r.Context(), r, uuid.FromBytesOrNil(state.FlowId))
	if err != nil {
		return nil, state, nil, err
	}

	tokenCode, hasSessionTokenCode, err := s.d.SessionTokenExchangePersister().CodeForFlow(r.Context(), f.GetID())
	if err != nil {
		return nil, state, nil, err
	}

	cntnr := AuthCodeContainer{}
	if f.GetType() == flow.TypeBrowser || !hasSessionTokenCode {
		if _, err := s.d.ContinuityManager().Continue(r.Context(), w, r, sessionName,
			continuity.WithPayload(&cntnr),
			continuity.WithExpireInsteadOfDelete(time.Minute),
		); err != nil {
			return nil, state, nil, err
		}
		if stateParam != cntnr.State {
			return nil, state, &cntnr, errors.WithStack(herodot.ErrBadRequest.WithReasonf(`Unable to complete OpenID Connect flow because the query state parameter does not match the state parameter from the session cookie.`))
		}
	} else {
		// We need to validate the tokenCode here
		if !codeMatches(state, tokenCode.InitCode) {
			return nil, state, &cntnr, errors.WithStack(herodot.ErrBadRequest.WithReasonf(`Unable to complete OpenID Connect flow because the query state parameter does not match the state parameter from the code.`))
		}
		cntnr.State = stateParam
		cntnr.FlowID = uuid.FromBytesOrNil(state.FlowId).String()
	}

	if errorParam != "" {
		return f, state, &cntnr, errors.WithStack(herodot.ErrBadRequest.WithReasonf(`Unable to complete OpenID Connect flow because the OpenID Provider returned error "%s": %s`, r.URL.Query().Get("error"), r.URL.Query().Get("error_description")))
	}

	if codeParam == "" {
		return f, state, &cntnr, errors.WithStack(herodot.ErrBadRequest.WithReasonf(`Unable to complete OpenID Connect flow because the OpenID Provider did not return the code query parameter.`))
	}

	return f, state, &cntnr, nil
}

func registrationOrLoginFlowID(flow any) (uuid.UUID, bool) {
	switch f := flow.(type) {
	case *registration.Flow:
		return f.ID, true
	case *login.Flow:
		return f.ID, true
	default:
		return uuid.Nil, false
	}
}

func (s *Strategy) alreadyAuthenticated(ctx context.Context, w http.ResponseWriter, r *http.Request, f interface{}) (bool, error) {
	if sess, _ := s.d.SessionManager().FetchFromRequest(ctx, r); sess != nil {
		if _, ok := f.(*settings.Flow); ok {
			// ignore this if it's a settings flow
		} else if !isForced(f) {
			returnTo := s.d.Config().SelfServiceBrowserDefaultReturnTo(ctx)
			if redirecter, ok := f.(flow.FlowWithRedirect); ok {
				r, err := redir.SecureRedirectTo(r, returnTo, redirecter.SecureRedirectToOpts(ctx, s.d)...)
				if err == nil {
					returnTo = r
				}
			}
			if flowID, ok := registrationOrLoginFlowID(f); ok {
				if codes, hasCode, _ := s.d.SessionTokenExchangePersister().CodeForFlow(ctx, flowID); hasCode {
					if err := s.d.SessionTokenExchangePersister().UpdateSessionOnExchanger(ctx, flowID, sess.ID); err != nil {
						return false, err
					}
					q := returnTo.Query()
					q.Set("code", codes.ReturnToCode)
					returnTo.RawQuery = q.Encode()
				}
			}
			http.Redirect(w, r, returnTo.String(), http.StatusSeeOther)
			return true, nil
		}
	}

	return false, nil
}

func (s *Strategy) HandleCallback(w http.ResponseWriter, r *http.Request) {
	var (
		code = cmp.Or(r.URL.Query().Get("code"), r.URL.Query().Get("authCode"))
		err  error
	)

	ctx := r.Context()
	ctx, span := s.d.Tracer(ctx).Tracer().Start(ctx, "strategy.oidc.HandleCallback")
	defer otelx.End(span, &err)
	r = r.WithContext(ctx)

	req, state, cntnr, err := s.ValidateCallback(w, r)
	if err != nil {
		if req != nil {
			s.forwardError(ctx, w, r, req, s.HandleError(ctx, w, r, req, state.ProviderId, nil, err))
		} else {
			s.d.SelfServiceErrorManager().Forward(ctx, w, r, s.HandleError(ctx, w, r, nil, "", nil, err))
		}
		return
	}

	if authenticated, err := s.alreadyAuthenticated(ctx, w, r, req); err != nil {
		s.forwardError(ctx, w, r, req, s.HandleError(ctx, w, r, req, state.ProviderId, nil, err))
	} else if authenticated {
		return
	}

	provider, err := s.Provider(ctx, state.ProviderId)
	if err != nil {
		s.forwardError(ctx, w, r, req, s.HandleError(ctx, w, r, req, state.ProviderId, nil, err))
		return
	}

	var claims *Claims
	var et *identity.CredentialsOIDCEncryptedTokens
	switch p := provider.(type) {
	case OAuth2Provider:
		token, err := s.exchangeCode(ctx, p, code, PKCEVerifier(state))
		if err != nil {
			s.forwardError(ctx, w, r, req, s.HandleError(ctx, w, r, req, state.ProviderId, nil, err))
			return
		}

		et, err = s.encryptOAuth2Tokens(ctx, token)
		if err != nil {
			s.forwardError(ctx, w, r, req, s.HandleError(ctx, w, r, req, state.ProviderId, nil, err))
			return
		}

		claims, err = p.Claims(ctx, token, r.URL.Query())
		if err != nil {
			s.forwardError(ctx, w, r, req, s.HandleError(ctx, w, r, req, state.ProviderId, nil, err))
			return
		}
	case OAuth1Provider:
		token, err := p.ExchangeToken(ctx, r)
		if err != nil {
			s.forwardError(ctx, w, r, req, s.HandleError(ctx, w, r, req, state.ProviderId, nil, err))
			return
		}

		claims, err = p.Claims(ctx, token)
		if err != nil {
			s.forwardError(ctx, w, r, req, s.HandleError(ctx, w, r, req, state.ProviderId, nil, err))
			return
		}
	}

	if err = claims.Validate(); err != nil {
		s.forwardError(ctx, w, r, req, s.HandleError(ctx, w, r, req, state.ProviderId, nil, err))
		return
	}

	span.SetAttributes(attribute.StringSlice("claims", slices.Collect(maps.Keys(claims.RawClaims))))

	switch a := req.(type) {
	case *login.Flow:
		a.Active = s.ID()
		a.TransientPayload = cntnr.TransientPayload
		a.IdentitySchema = cntnr.IdentitySchema
		if ff, err := s.ProcessLogin(ctx, w, r, a, et, claims, provider, cntnr); err != nil {
			if errors.Is(err, flow.ErrCompletedByStrategy) {
				return
			}
			if ff != nil {
				s.forwardError(ctx, w, r, ff, err)
				return
			}
			s.forwardError(ctx, w, r, a, err)
		}
		return
	case *registration.Flow:
		a.Active = s.ID()
		a.TransientPayload = cntnr.TransientPayload
		a.IdentitySchema = cntnr.IdentitySchema
		if ff, err := s.processRegistration(ctx, w, r, a, et, claims, provider, cntnr); errors.Is(err, flow.ErrCompletedByStrategy) {
			return
		} else if err != nil {
			if ff != nil {
				s.forwardError(ctx, w, r, ff, err)
				return
			}
			s.forwardError(ctx, w, r, a, err)
		}
		return
	case *settings.Flow:
		a.Active = sqlxx.NullString(s.ID())
		a.TransientPayload = cntnr.TransientPayload
		sess, err := s.d.SessionManager().FetchFromRequest(ctx, r)
		if err != nil {
			s.forwardError(ctx, w, r, a, s.HandleError(ctx, w, r, a, state.ProviderId, nil, err))
			return
		}
		if err := s.linkProvider(ctx, w, r, &settings.UpdateContext{Session: sess, Flow: a}, et, claims, provider); err != nil {
			s.forwardError(ctx, w, r, a, s.HandleError(ctx, w, r, a, state.ProviderId, nil, err))
			return
		}
		return
	default:
		s.forwardError(ctx, w, r, req, s.HandleError(ctx, w, r, req, state.ProviderId, nil, errors.WithStack(x.PseudoPanic.
			WithDetailf("cause", "Unexpected type in OpenID Connect flow: %T", a))))
		return
	}
}

func (s *Strategy) exchangeCode(ctx context.Context, provider OAuth2Provider, code string, opts []oauth2.AuthCodeOption) (token *oauth2.Token, err error) {
	ctx, span := s.d.Tracer(ctx).Tracer().Start(ctx, "strategy.oidc.exchangeCode", trace.WithAttributes(
		attribute.String("provider_id", provider.Config().ID),
		attribute.String("provider_label", provider.Config().Label)))
	defer otelx.End(span, &err)

	te, ok := provider.(OAuth2TokenExchanger)
	if !ok {
		te, err = provider.OAuth2(ctx)
		if err != nil {
			return nil, err
		}
	}

	client := s.d.HTTPClient(ctx)
	ctx = context.WithValue(ctx, oauth2.HTTPClient, client.HTTPClient)
	token, err = te.Exchange(ctx, code, opts...)
	return token, err
}

func (s *Strategy) populateMethod(r *http.Request, f flow.Flow, message func(provider string, providerId string) *text.Message) error {
	conf, err := s.Config(r.Context())
	if err != nil {
		return err
	}

	f.GetUI().SetCSRF(s.d.GenerateCSRFToken(r))
	AddProviders(f.GetUI(), conf.Providers, message, s.ID())

	return nil
}

func (s *Strategy) Config(ctx context.Context) (*ConfigurationCollection, error) {
	var c ConfigurationCollection

	conf := s.d.Config().SelfServiceStrategy(ctx, string(s.ID())).Config
	if err := json.
		NewDecoder(bytes.NewBuffer(conf)).
		Decode(&c); err != nil {
		s.d.Logger().WithError(err).WithField("config", conf)
		return nil, errors.WithStack(herodot.ErrMisconfiguration.WithReasonf("Unable to decode OpenID Connect Provider configuration: %s", err))
	}

	return &c, nil
}

func (s *Strategy) Provider(ctx context.Context, id string) (Provider, error) {
	if c, err := s.Config(ctx); err != nil {
		return nil, err
	} else if provider, err := c.Provider(id, s.d); err != nil {
		return nil, s.handleUnknownProviderError(err)
	} else {
		return provider, nil
	}
}

func (s *Strategy) forwardError(ctx context.Context, w http.ResponseWriter, r *http.Request, f flow.Flow, err error) {
	switch ff := f.(type) {
	case *login.Flow:
		s.d.LoginFlowErrorHandler().WriteFlowError(w, r, ff, s.NodeGroup(), err)
	case *registration.Flow:
		s.d.RegistrationFlowErrorHandler().WriteFlowError(w, r, ff, s.NodeGroup(), err)
	case *settings.Flow:
		var i *identity.Identity
		var sess *session.Session
		if currentSession, err := s.d.SessionManager().FetchFromRequest(ctx, r); err == nil {
			i = currentSession.Identity
			sess = currentSession
		}
		s.d.SettingsFlowErrorHandler().WriteFlowError(ctx, w, r, s.NodeGroup(), ff, i, sess, err)
	default:
		panic(errors.Errorf("unexpected type: %T", ff))
	}
}

func (s *Strategy) HandleError(ctx context.Context, w http.ResponseWriter, r *http.Request, f flow.Flow, usedProviderID string, traits []byte, err error) error {
	switch rf := f.(type) {
	case *login.Flow:
		return err
	case *registration.Flow:
		// Reset all nodes to not confuse users.
		// This is kinda hacky and will probably need to be updated at some point.

		if dup := new(identity.ErrDuplicateCredentials); errors.As(err, &dup) {
			err = schema.NewDuplicateCredentialsError(dup)

			if validationErr := new(schema.ValidationError); errors.As(err, &validationErr) {
				for _, m := range validationErr.Messages {
					m := m
					rf.UI.Messages.Add(&m)
				}
			} else {
				rf.UI.Messages.Add(text.NewErrorValidationDuplicateCredentialsOnOIDCLink())
			}

			lf, err := s.registrationToLogin(ctx, w, r, rf)
			if err != nil {
				return err
			}
			// return a new login flow with the error message embedded in the login flow.
			var redirectURL *url.URL
			if lf.Type == flow.TypeAPI {
				returnTo := s.d.Config().SelfServiceBrowserDefaultReturnTo(ctx)
				if redirecter, ok := f.(flow.FlowWithRedirect); ok {
					secureReturnTo, err := redir.SecureRedirectTo(r, returnTo, redirecter.SecureRedirectToOpts(ctx, s.d)...)
					if err == nil {
						returnTo = secureReturnTo
					}
				}
				redirectURL = lf.AppendTo(returnTo)
			} else {
				redirectURL = lf.AppendTo(s.d.Config().SelfServiceFlowLoginUI(ctx))
			}
			if dc, err := flow.DuplicateCredentials(lf); err == nil && dc != nil {
				redirectURL = urlx.CopyWithQuery(redirectURL, url.Values{"no_org_ui": {"true"}})
				s.populateAccountLinkingUI(ctx, lf, usedProviderID, dc.DuplicateIdentifier, dup.AvailableCredentials(), dup.AvailableOIDCProviders())
				if err := s.d.LoginFlowPersister().UpdateLoginFlow(ctx, lf); err != nil {
					return err
				}
			}
			x.SendFlowErrorAsRedirectOrJSON(w, r, s.d.Writer(), lf, redirectURL.String())
			// ensure the function does not continue to execute
			return errors.WithStack(flow.ErrCompletedByStrategy)
		}

		rf.UI.Nodes = node.Nodes{}

		// Adds the "Continue" button
		rf.UI.SetCSRF(s.d.GenerateCSRFToken(r))
		AddProvider(rf.UI, usedProviderID, text.NewInfoRegistrationContinue(), s.ID())

		group := node.DefaultGroup
		if s.d.Config().SelfServiceLegacyOIDCRegistrationGroup(ctx) {
			group = node.OpenIDConnectGroup
			trace.SpanFromContext(r.Context()).AddEvent(semconv.NewDeprecatedFeatureUsedEvent(r.Context(), "legacy_oidc_registration_group"))
		}

		if traits != nil {
			ds, err := rf.IdentitySchema.URL(ctx, s.d.Config())
			if err != nil {
				return err
			}

			traitNodes, err := container.NodesFromJSONSchema(ctx, group, ds.String(), "", nil)
			if err != nil {
				return err
			}

			rf.UI.Nodes = append(rf.UI.Nodes, traitNodes...)
			rf.UI.UpdateNodeValuesFromJSON(traits, "traits", group)
		}

		return err
	case *settings.Flow:
		return err
	}

	return err
}

func (s *Strategy) populateAccountLinkingUI(ctx context.Context, lf *login.Flow, usedProviderID string, duplicateIdentifier string, availableCredentials []string, availableProviders []string) {
	newLoginURL := s.d.Config().SelfServiceFlowLoginUI(ctx).String()
	usedProviderLabel := usedProviderID
	provider, _ := s.Provider(ctx, usedProviderID)
	if provider != nil && provider.Config() != nil {
		usedProviderLabel = provider.Config().Label
		if usedProviderLabel == "" {
			usedProviderLabel = provider.Config().Provider
		}
	}
	loginHintsEnabled := s.d.Config().SelfServiceFlowRegistrationLoginHints(ctx)
	nodes := []*node.Node{}
	for _, n := range lf.UI.Nodes {
		// We don't want to touch nodes unecessary nodes
		if n.Meta == nil || n.Meta.Label == nil || n.Group == node.DefaultGroup {
			nodes = append(nodes, n)
			continue
		}

		// Skip the provider that was used to get here (in case they used an OIDC provider)
		pID := gjson.GetBytes(n.Meta.Label.Context, "provider_id").String()
		if n.Group == node.OpenIDConnectGroup {
			if pID == usedProviderID {
				continue
			}
			// Hide any provider that is not available for the user
			if loginHintsEnabled && !slices.Contains(availableProviders, pID) {
				continue
			}
		}

		// Replace some labels to make it easier for the user to understand what's going on.
		switch n.Meta.Label.ID {
		case text.InfoSelfServiceLogin:
			n.Meta.Label = text.NewInfoLoginAndLink()
		case text.InfoSelfServiceLoginWith:
			p := gjson.GetBytes(n.Meta.Label.Context, "provider").String()
			n.Meta.Label = text.NewInfoLoginWithAndLink(p)
		}

		// This can happen, if login hints are disabled. In that case, we need to make sure to show all credential options.
		// It could in theory also happen due to a mis-configuration, and in that case, we should make sure to not delete the entire flow.
		if !loginHintsEnabled {
			nodes = append(nodes, n)
		} else {
			// Hide nodes from credentials that are not relevant for the user
			for _, ct := range availableCredentials {
				if ct == string(n.Group) {
					nodes = append(nodes, n)
					break
				}
			}
		}
	}

	// Hide the "primary" identifier field present for Password, webauthn or passwordless, as we already know the identifier
	identifierNode := lf.UI.Nodes.Find("identifier")
	if identifierNode != nil {
		if attributes, ok := identifierNode.Attributes.(*node.InputAttributes); ok {
			attributes.Type = node.InputAttributeTypeHidden
			attributes.SetValue(duplicateIdentifier)
			identifierNode.Attributes = attributes
		}
	}

	lf.UI.Nodes = nodes
	lf.UI.Messages.Clear()
	lf.UI.Messages.Add(text.NewInfoLoginLinkMessage(duplicateIdentifier, usedProviderLabel, newLoginURL, availableCredentials, availableProviders))
}

func (s *Strategy) NodeGroup() node.UiNodeGroup {
	return node.OpenIDConnectGroup
}

func (s *Strategy) CompletedAuthenticationMethod(context.Context) session.AuthenticationMethod {
	return session.AuthenticationMethod{
		Method: s.ID(),
		AAL:    identity.AuthenticatorAssuranceLevel1,
	}
}

func (s *Strategy) ProcessIDToken(r *http.Request, provider Provider, idToken, idTokenNonce string) (*Claims, error) {
	verifier, ok := provider.(IDTokenVerifier)
	if !ok {
		return nil, errors.WithStack(herodot.ErrUpstreamError.WithReasonf("The provider %s does not support id_token verification", provider.Config().Provider))
	}
	claims, err := verifier.Verify(r.Context(), idToken)
	if err != nil {
		return nil, errors.WithStack(herodot.ErrForbidden.WithReasonf("Could not verify id_token").WithWrap(err).WithError(err.Error()))
	}

	if err := claims.Validate(); err != nil {
		return nil, errors.WithStack(herodot.ErrForbidden.WithReasonf("The id_token claims were invalid").WithWrap(err))
	}

	// First check if the JWT contains the nonce claim.
	if claims.Nonce == "" {
		// If it doesn't, check if the provider supports nonces.
		if nonceSkipper, ok := verifier.(NonceValidationSkipper); !ok || !nonceSkipper.CanSkipNonce(claims) {
			// If the provider supports nonces, abort the flow!
			return nil, errors.WithStack(herodot.ErrUpstreamError.WithReasonf("No nonce was included in the id_token but is required by the provider"))
		}
		// If the provider does not support nonces, we don't do validation and return the claim.
		// This case only applies to Apple, as some of their devices do not support nonces.
		// https://developer.apple.com/documentation/sign_in_with_apple/sign_in_with_apple_rest_api/authenticating_users_with_sign_in_with_apple
	} else if idTokenNonce == "" {
		// A nonce was present in the JWT token, but no nonce was submitted in the flow
		return nil, errors.WithStack(herodot.ErrUpstreamError.WithReasonf("No nonce was provided but is required by the provider"))
	} else if idTokenNonce != claims.Nonce {
		// The nonce from the JWT token does not match the nonce from the flow.
		return nil, errors.WithStack(herodot.ErrUpstreamError.WithReasonf("The supplied nonce does not match the nonce from the id_token"))
	}
	// Nonce checking was successful

	return claims, nil
}

func (s *Strategy) linkCredentials(ctx context.Context, i *identity.Identity, tokens *identity.CredentialsOIDCEncryptedTokens, provider, subject, organization string) (err error) {
	ctx, span := s.d.Tracer(ctx).Tracer().Start(ctx, "strategy.oidc.linkCredentials", trace.WithAttributes(
		attribute.String("provider", provider),
		// attribute.String("subject", subject), // PII
		attribute.String("organization", organization)))
	defer otelx.End(span, &err)

	if len(i.Credentials) == 0 {
		if err := s.d.PrivilegedIdentityPool().HydrateIdentityAssociations(ctx, i, identity.ExpandCredentials); err != nil {
			return err
		}
	}

	var conf identity.CredentialsOIDC
	creds, err := i.ParseCredentials(s.ID(), &conf)
	if errors.Is(err, herodot.ErrNotFound) {
		var err error
		if creds, err = identity.NewOIDCLikeCredentials(tokens, s.ID(), provider, subject, organization); err != nil {
			return err
		}
	} else if err != nil {
		return err
	} else {
		creds.Identifiers = append(creds.Identifiers, identity.OIDCUniqueID(provider, subject))
		conf.Providers = append(conf.Providers, identity.CredentialsOIDCProvider{
			Subject:             subject,
			Provider:            provider,
			InitialAccessToken:  tokens.GetAccessToken(),
			InitialRefreshToken: tokens.GetRefreshToken(),
			InitialIDToken:      tokens.GetIDToken(),
			Organization:        organization,
		})

		creds.Config, err = json.Marshal(conf)
		if err != nil {
			return err
		}
	}

	i.Credentials[s.ID()] = *creds
	if orgID, err := uuid.FromString(organization); err == nil {
		i.OrganizationID = uuid.NullUUID{UUID: orgID, Valid: true}
	}

	return nil
}

func getAuthRedirectURL(ctx context.Context, provider Provider, req ider, state string, upstreamParameters map[string]string, opts []oauth2.AuthCodeOption) (codeURL string, err error) {
	switch p := provider.(type) {
	case OAuth2Provider:
		c, err := p.OAuth2(ctx)
		if err != nil {
			return "", err
		}
		opts = append(opts, UpstreamParameters(upstreamParameters)...)
		opts = append(opts, p.AuthCodeURLOptions(req)...)

		return c.AuthCodeURL(state, opts...), nil
	case OAuth1Provider:
		return p.AuthURL(ctx, state)
	default:
		return "", errors.WithStack(herodot.ErrInternalServerError.WithReasonf("The provider %s does not support the OAuth 2.0 or OAuth 1.0 protocol", provider.Config().Provider))
	}
}

func (s *Strategy) encryptOAuth2Tokens(ctx context.Context, token *oauth2.Token) (et *identity.CredentialsOIDCEncryptedTokens, err error) {
	et = new(identity.CredentialsOIDCEncryptedTokens)
	if token == nil {
		return et, nil
	}

	if idToken, ok := token.Extra("id_token").(string); ok {
		et.IDToken, err = s.d.Cipher(ctx).Encrypt(ctx, []byte(idToken))
		if err != nil {
			return nil, err
		}
	}

	et.AccessToken, err = s.d.Cipher(ctx).Encrypt(ctx, []byte(token.AccessToken))
	if err != nil {
		return nil, err
	}

	et.RefreshToken, err = s.d.Cipher(ctx).Encrypt(ctx, []byte(token.RefreshToken))
	if err != nil {
		return nil, err
	}

	return et, nil
}
