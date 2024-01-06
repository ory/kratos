// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package oidc

import (
	"bytes"
	"context"
	"crypto/sha512"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/ory/x/urlx"

	"go.opentelemetry.io/otel/attribute"
	"golang.org/x/oauth2"

	"github.com/ory/kratos/cipher"
	"github.com/ory/kratos/selfservice/flowhelpers"
	"github.com/ory/kratos/selfservice/sessiontokenexchange"
	"github.com/ory/x/jsonnetsecure"
	"github.com/ory/x/otelx"

	"github.com/ory/kratos/text"

	"github.com/ory/kratos/ui/container"
	"github.com/ory/x/decoderx"
	"github.com/ory/x/stringsx"

	"github.com/ory/kratos/ui/node"

	"github.com/gofrs/uuid"
	"github.com/julienschmidt/httprouter"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"

	"github.com/ory/x/jsonx"

	"github.com/ory/herodot"
	"github.com/ory/kratos/continuity"
	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/schema"
	"github.com/ory/kratos/selfservice/errorx"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/selfservice/flow/login"
	"github.com/ory/kratos/selfservice/flow/registration"
	"github.com/ory/kratos/selfservice/flow/settings"

	"github.com/ory/kratos/selfservice/strategy"
	"github.com/ory/kratos/session"
	"github.com/ory/kratos/x"
)

const (
	RouteBase = "/self-service/methods/oidc"

	RouteAuth                 = RouteBase + "/auth/:flow"
	RouteCallback             = RouteBase + "/callback/:provider"
	RouteOrganizationCallback = RouteBase + "/organization/:organization/callback/:provider"
)

var _ identity.ActiveCredentialsCounter = new(Strategy)

type Dependencies interface {
	errorx.ManagementProvider

	config.Provider

	x.LoggingProvider
	x.CookieProvider
	x.CSRFProvider
	x.CSRFTokenGeneratorProvider
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
		IsForced() bool
	})
	return ok && f.IsForced()
}

// Strategy implements selfservice.LoginStrategy, selfservice.RegistrationStrategy and selfservice.SettingsStrategy.
// It supports login, registration and settings via OpenID Providers.
type Strategy struct {
	d         Dependencies
	validator *schema.Validator
	dec       *decoderx.HTTP
}

type AuthCodeContainer struct {
	FlowID           string          `json:"flow_id"`
	State            string          `json:"state"`
	Traits           json.RawMessage `json:"traits"`
	TransientPayload json.RawMessage `json:"transient_payload"`
}

type State struct {
	FlowID string
	Data   []byte
}

func (s *State) String() string {
	return base64.RawURLEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", s.FlowID, s.Data)))
}

func generateState(flowID string) *State {
	return &State{
		FlowID: flowID,
		Data:   x.NewUUID().Bytes(),
	}
}

func (s *State) setCode(code string) {
	s.Data = sha512.New().Sum([]byte(code))
}

func (s *State) codeMatches(code string) bool {
	return bytes.Equal(s.Data, sha512.New().Sum([]byte(code)))
}

func parseState(s string) (*State, error) {
	raw, err := base64.RawURLEncoding.DecodeString(s)
	if err != nil {
		return nil, err
	}
	if id, data, ok := bytes.Cut(raw, []byte(":")); !ok {
		return nil, errors.New("state has invalid format")
	} else {
		return &State{FlowID: string(id), Data: data}, nil
	}
}

func (s *Strategy) CountActiveFirstFactorCredentials(cc map[identity.CredentialsType]identity.Credentials) (count int, err error) {
	for _, c := range cc {
		if c.Type == s.ID() && gjson.ValidBytes(c.Config) {
			var conf identity.CredentialsOIDC
			if err = json.Unmarshal(c.Config, &conf); err != nil {
				return 0, errors.WithStack(err)
			}

			for _, ider := range c.Identifiers {
				parts := strings.Split(ider, ":")
				if len(parts) != 2 {
					continue
				}

				for _, prov := range conf.Providers {
					if parts[0] == prov.Provider && parts[1] == prov.Subject && len(prov.Subject) > 1 && len(prov.Provider) > 1 {
						count++
					}
				}
			}
		}
	}
	return
}

func (s *Strategy) CountActiveMultiFactorCredentials(cc map[identity.CredentialsType]identity.Credentials) (count int, err error) {
	return 0, nil
}

func (s *Strategy) setRoutes(r *x.RouterPublic) {
	wrappedHandleCallback := strategy.IsDisabled(s.d, s.ID().String(), s.HandleCallback)
	if handle, _, _ := r.Lookup("GET", RouteCallback); handle == nil {
		r.GET(RouteCallback, wrappedHandleCallback)
	}

	// Apple can use the POST request method when calling the callback
	if handle, _, _ := r.Lookup("POST", RouteCallback); handle == nil {
		// Hardcoded path to Apple provider, I don't have a better way of doing it right now.
		// Also this ignore disables CSRF checks for both GET and POST requests. Unfortunately
		// CSRF handler does not allow to define a rule based on the request method, at least not yet.
		s.d.CSRFHandler().IgnorePath(RouteBase + "/callback/apple")

		// When handler is called using POST method, the cookies are not attached to the request
		// by the browser. So here we just redirect the request to the same location rewriting the
		// form fields to query params. This second GET request should have the cookies attached.
		r.POST(RouteCallback, s.redirectToGET)
	}
}

// Redirect POST request to GET rewriting form fields to query params.
func (s *Strategy) redirectToGET(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
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

func NewStrategy(d any) *Strategy {
	return &Strategy{
		d:         d.(Dependencies),
		validator: schema.NewValidator(),
	}
}

func (s *Strategy) ID() identity.CredentialsType {
	return identity.CredentialsTypeOIDC
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

func (s *Strategy) ValidateCallback(w http.ResponseWriter, r *http.Request) (flow.Flow, *AuthCodeContainer, error) {
	var (
		codeParam  = stringsx.Coalesce(r.URL.Query().Get("code"), r.URL.Query().Get("authCode"))
		stateParam = r.URL.Query().Get("state")
		errorParam = r.URL.Query().Get("error")
	)

	if stateParam == "" {
		return nil, nil, errors.WithStack(herodot.ErrBadRequest.WithReasonf(`Unable to complete OpenID Connect flow because the OpenID Provider did not return the state query parameter.`))
	}
	state, err := parseState(stateParam)
	if err != nil {
		return nil, nil, errors.WithStack(herodot.ErrBadRequest.WithReasonf(`Unable to complete OpenID Connect flow because the state parameter was invalid.`))
	}

	f, err := s.validateFlow(r.Context(), r, x.ParseUUID(state.FlowID))
	if err != nil {
		return nil, nil, err
	}

	tokenCode, hasSessionTokenCode, err := s.d.SessionTokenExchangePersister().CodeForFlow(r.Context(), f.GetID())
	if err != nil {
		return nil, nil, err
	}

	cntnr := AuthCodeContainer{}
	if f.GetType() == flow.TypeBrowser || !hasSessionTokenCode {
		if _, err := s.d.ContinuityManager().Continue(r.Context(), w, r, sessionName, continuity.WithPayload(&cntnr)); err != nil {
			return nil, nil, err
		}
		if stateParam != cntnr.State {
			return nil, &cntnr, errors.WithStack(herodot.ErrBadRequest.WithReasonf(`Unable to complete OpenID Connect flow because the query state parameter does not match the state parameter from the session cookie.`))
		}
	} else {
		// We need to validate the tokenCode here
		if !state.codeMatches(tokenCode.InitCode) {
			return nil, &cntnr, errors.WithStack(herodot.ErrBadRequest.WithReasonf(`Unable to complete OpenID Connect flow because the query state parameter does not match the state parameter from the code.`))
		}
		cntnr.State = stateParam
		cntnr.FlowID = state.FlowID
	}

	if errorParam != "" {
		return f, &cntnr, errors.WithStack(herodot.ErrBadRequest.WithReasonf(`Unable to complete OpenID Connect flow because the OpenID Provider returned error "%s": %s`, r.URL.Query().Get("error"), r.URL.Query().Get("error_description")))
	}
	if codeParam == "" {
		return f, &cntnr, errors.WithStack(herodot.ErrBadRequest.WithReasonf(`Unable to complete OpenID Connect flow because the OpenID Provider did not return the code query parameter.`))
	}

	return f, &cntnr, nil
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

func (s *Strategy) alreadyAuthenticated(w http.ResponseWriter, r *http.Request, f interface{}) (bool, error) {
	ctx := r.Context()

	if sess, _ := s.d.SessionManager().FetchFromRequest(ctx, r); sess != nil {
		if _, ok := f.(*settings.Flow); ok {
			// ignore this if it's a settings flow
		} else if !isForced(f) {
			if flowID, ok := registrationOrLoginFlowID(f); ok {
				if _, hasCode, _ := s.d.SessionTokenExchangePersister().CodeForFlow(ctx, flowID); hasCode {
					err := s.d.SessionTokenExchangePersister().UpdateSessionOnExchanger(ctx, flowID, sess.ID)
					if err != nil {
						return false, err
					}
				}
			}
			returnTo := s.d.Config().SelfServiceBrowserDefaultReturnTo(ctx)
			if redirecter, ok := f.(flow.FlowWithRedirect); ok {
				r, err := x.SecureRedirectTo(r, returnTo, redirecter.SecureRedirectToOpts(ctx, s.d)...)
				if err != nil {
					returnTo = r
				}
			}
			http.Redirect(w, r, returnTo.String(), http.StatusSeeOther)
			return true, nil
		}
	}

	return false, nil
}

func (s *Strategy) HandleCallback(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var (
		code = stringsx.Coalesce(r.URL.Query().Get("code"), r.URL.Query().Get("authCode"))
		pid  = ps.ByName("provider")
	)

	ctx := r.Context()
	ctx = context.WithValue(ctx, httprouter.ParamsKey, ps)
	r = r.WithContext(ctx)

	req, cntnr, err := s.ValidateCallback(w, r)
	if err != nil {
		if req != nil {
			s.forwardError(w, r, req, s.handleError(w, r, req, pid, nil, err))
		} else {
			s.d.SelfServiceErrorManager().Forward(r.Context(), w, r, s.handleError(w, r, nil, pid, nil, err))
		}
		return
	}

	if authenticated, err := s.alreadyAuthenticated(w, r, req); err != nil {
		s.forwardError(w, r, req, s.handleError(w, r, req, pid, nil, err))
	} else if authenticated {
		return
	}

	provider, err := s.provider(r.Context(), r, pid)
	if err != nil {
		s.forwardError(w, r, req, s.handleError(w, r, req, pid, nil, err))
		return
	}

	token, err := s.ExchangeCode(r.Context(), provider, code)
	if err != nil {
		s.forwardError(w, r, req, s.handleError(w, r, req, pid, nil, err))
		return
	}

	claims, err := provider.Claims(r.Context(), token, r.URL.Query())
	if err != nil {
		s.forwardError(w, r, req, s.handleError(w, r, req, pid, nil, err))
		return
	}

	if err := claims.Validate(); err != nil {
		s.forwardError(w, r, req, s.handleError(w, r, req, pid, nil, err))
		return
	}

	switch a := req.(type) {
	case *login.Flow:
		if ff, err := s.processLogin(w, r, a, token, claims, provider, cntnr); err != nil {
			if ff != nil {
				s.forwardError(w, r, ff, err)
				return
			}
			s.forwardError(w, r, a, err)
		}
		return
	case *registration.Flow:
		a.TransientPayload = cntnr.TransientPayload
		if ff, err := s.processRegistration(w, r, a, token, claims, provider, cntnr, ""); err != nil {
			if ff != nil {
				s.forwardError(w, r, ff, err)
				return
			}
			s.forwardError(w, r, a, err)
		}
		return
	case *settings.Flow:
		sess, err := s.d.SessionManager().FetchFromRequest(r.Context(), r)
		if err != nil {
			s.forwardError(w, r, a, s.handleError(w, r, a, pid, nil, err))
			return
		}
		if err := s.linkProvider(w, r, &settings.UpdateContext{Session: sess, Flow: a}, token, claims, provider); err != nil {
			s.forwardError(w, r, a, s.handleError(w, r, a, pid, nil, err))
			return
		}
		return
	default:
		s.forwardError(w, r, req, s.handleError(w, r, req, pid, nil, errors.WithStack(x.PseudoPanic.
			WithDetailf("cause", "Unexpected type in OpenID Connect flow: %T", a))))
		return
	}
}

func (s *Strategy) ExchangeCode(ctx context.Context, provider Provider, code string) (token *oauth2.Token, err error) {
	ctx, span := s.d.Tracer(ctx).Tracer().Start(ctx, "strategy.oidc.ExchangeCode")
	defer otelx.End(span, &err)
	span.SetAttributes(attribute.String("provider_id", provider.Config().ID))
	span.SetAttributes(attribute.String("provider_label", provider.Config().Label))

	te, ok := provider.(TokenExchanger)
	if !ok {
		te, err = provider.OAuth2(ctx)
		if err != nil {
			return nil, err
		}
	}

	client := s.d.HTTPClient(ctx)
	ctx = context.WithValue(ctx, oauth2.HTTPClient, client.HTTPClient)
	token, err = te.Exchange(ctx, code)
	return token, err
}

func (s *Strategy) populateMethod(r *http.Request, f flow.Flow, message func(provider string) *text.Message) error {
	conf, err := s.Config(r.Context())
	if err != nil {
		return err
	}

	providers := conf.Providers

	if lf, ok := f.(*login.Flow); ok && lf.IsForced() {
		if _, id, c := flowhelpers.GuessForcedLoginIdentifier(r, s.d, lf, s.ID()); id != nil {
			if c == nil {
				// no OIDC credentials, don't add any providers
				providers = nil
			} else {
				var credentials identity.CredentialsOIDC
				if err := json.Unmarshal(c.Config, &credentials); err != nil {
					// failed to read OIDC credentials, don't add any providers
					providers = nil
				} else {
					// add only providers that can actually be used to log in as this identity
					providers = make([]Configuration, 0, len(conf.Providers))
					for i := range conf.Providers {
						for j := range credentials.Providers {
							if conf.Providers[i].ID == credentials.Providers[j].Provider {
								providers = append(providers, conf.Providers[i])
								break
							}
						}
					}
				}
			}
		}
	}

	// does not need sorting because there is only one field
	c := f.GetUI()
	c.SetCSRF(s.d.GenerateCSRFToken(r))
	AddProviders(c, providers, message)

	return nil
}

func (s *Strategy) Config(ctx context.Context) (*ConfigurationCollection, error) {
	var c ConfigurationCollection

	conf := s.d.Config().SelfServiceStrategy(ctx, string(s.ID())).Config
	if err := jsonx.
		NewStrictDecoder(bytes.NewBuffer(conf)).
		Decode(&c); err != nil {
		s.d.Logger().WithError(err).WithField("config", conf)
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("Unable to decode OpenID Connect Provider configuration: %s", err))
	}

	return &c, nil
}

func (s *Strategy) provider(ctx context.Context, r *http.Request, id string) (Provider, error) {
	if c, err := s.Config(ctx); err != nil {
		return nil, err
	} else if provider, err := c.Provider(id, s.d); err != nil {
		return nil, err
	} else {
		return provider, nil
	}
}

func (s *Strategy) forwardError(w http.ResponseWriter, r *http.Request, f flow.Flow, err error) {
	switch ff := f.(type) {
	case *login.Flow:
		s.d.LoginFlowErrorHandler().WriteFlowError(w, r, ff, s.NodeGroup(), err)
	case *registration.Flow:
		s.d.RegistrationFlowErrorHandler().WriteFlowError(w, r, ff, s.NodeGroup(), err)
	case *settings.Flow:
		var i *identity.Identity
		if sess, err := s.d.SessionManager().FetchFromRequest(r.Context(), r); err == nil {
			i = sess.Identity
		}
		s.d.SettingsFlowErrorHandler().WriteFlowError(w, r, s.NodeGroup(), ff, i, err)
	default:
		panic(errors.Errorf("unexpected type: %T", ff))
	}
}

func (s *Strategy) handleError(w http.ResponseWriter, r *http.Request, f flow.Flow, providerID string, traits []byte, err error) error {
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

			lf, err := s.registrationToLogin(w, r, rf, providerID)
			if err != nil {
				return err
			}
			// return a new login flow with the error message embedded in the login flow.
			redirectURL := lf.AppendTo(s.d.Config().SelfServiceFlowLoginUI(r.Context()))
			if dc, err := flow.DuplicateCredentials(lf); err == nil && dc != nil {
				redirectURL = urlx.CopyWithQuery(redirectURL, url.Values{"no_org_ui": {"true"}})

				for i, n := range lf.UI.Nodes {
					if n.Meta == nil || n.Meta.Label == nil {
						continue
					}
					switch n.Meta.Label.ID {
					case text.InfoSelfServiceLogin:
						lf.UI.Nodes[i].Meta.Label = text.NewInfoLoginAndLink()
					case text.InfoSelfServiceLoginWith:
						p := gjson.GetBytes(n.Meta.Label.Context, "provider").String()
						lf.UI.Nodes[i].Meta.Label = text.NewInfoLoginWithAndLink(p)
					}
				}

				newLoginURL := s.d.Config().SelfServiceFlowLoginUI(r.Context()).String()
				providerLabel := providerID
				provider, _ := s.provider(r.Context(), r, providerID)
				if provider != nil && provider.Config() != nil {
					providerLabel = provider.Config().Label
				}
				lf.UI.Messages.Add(text.NewInfoLoginLinkMessage(dc.DuplicateIdentifier, providerLabel, newLoginURL))

				err := s.d.LoginFlowPersister().UpdateLoginFlow(r.Context(), lf)
				if err != nil {
					return err
				}
			}
			x.AcceptToRedirectOrJSON(w, r, s.d.Writer(), lf, redirectURL.String())
			// ensure the function does not continue to execute
			return flow.ErrCompletedByStrategy
		}

		rf.UI.Nodes = node.Nodes{}

		// Adds the "Continue" button
		rf.UI.SetCSRF(s.d.GenerateCSRFToken(r))
		AddProvider(rf.UI, providerID, text.NewInfoRegistrationContinue())

		if traits != nil {
			ds, err := s.d.Config().DefaultIdentityTraitsSchemaURL(r.Context())
			if err != nil {
				return err
			}

			traitNodes, err := container.NodesFromJSONSchema(r.Context(), node.OpenIDConnectGroup, ds.String(), "", nil)
			if err != nil {
				return err
			}

			rf.UI.Nodes = append(rf.UI.Nodes, traitNodes...)
			rf.UI.UpdateNodeValuesFromJSON(traits, "traits", node.OpenIDConnectGroup)
		}

		return err
	case *settings.Flow:
		return err
	}

	return err
}

func (s *Strategy) NodeGroup() node.UiNodeGroup {
	return node.OpenIDConnectGroup
}

func (s *Strategy) CompletedAuthenticationMethod(ctx context.Context, _ session.AuthenticationMethods) session.AuthenticationMethod {
	return session.AuthenticationMethod{
		Method: s.ID(),
		AAL:    identity.AuthenticatorAssuranceLevel1,
	}
}

func (s *Strategy) processIDToken(w http.ResponseWriter, r *http.Request, provider Provider, idToken, idTokenNonce string) (*Claims, error) {
	verifier, ok := provider.(IDTokenVerifier)
	if !ok {
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("The provider %s does not support id_token verification", provider.Config().Provider))
	}
	claims, err := verifier.Verify(r.Context(), idToken)
	if err != nil {
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("Could not verify id_token").WithError(err.Error()))
	}

	if err := claims.Validate(); err != nil {
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("The id_token claims were invalid").WithError(err.Error()))
	}

	// First check if the JWT contains the nonce claim.
	if claims.Nonce == "" {
		// If it doesn't, check if the provider supports nonces.
		if nonceSkipper, ok := verifier.(NonceValidationSkipper); !ok || !nonceSkipper.CanSkipNonce(claims) {
			// If the provider supports nonces, abort the flow!
			return nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("No nonce was included in the id_token but is required by the provider"))
		}
		// If the provider does not support nonces, we don't do validation and return the claim.
		// This case only applies to Apple, as some of their devices do not support nonces.
		// https://developer.apple.com/documentation/sign_in_with_apple/sign_in_with_apple_rest_api/authenticating_users_with_sign_in_with_apple
	} else if idTokenNonce == "" {
		// A nonce was present in the JWT token, but no nonce was submitted in the flow
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("No nonce was provided but is required by the provider"))
	} else if idTokenNonce != claims.Nonce {
		// The nonce from the JWT token does not match the nonce from the flow.
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("The supplied nonce does not match the nonce from the id_token"))
	}
	// Nonce checking was successful

	return claims, nil
}

func (s *Strategy) linkCredentials(ctx context.Context, i *identity.Identity, idToken, accessToken, refreshToken, provider, subject, organization string) error {
	if err := s.d.PrivilegedIdentityPool().HydrateIdentityAssociations(ctx, i, identity.ExpandCredentials); err != nil {
		return err
	}
	var conf identity.CredentialsOIDC
	creds, err := i.ParseCredentials(s.ID(), &conf)
	if errors.Is(err, herodot.ErrNotFound) {
		var err error
		if creds, err = identity.NewCredentialsOIDC(idToken, accessToken, refreshToken, provider, subject, organization); err != nil {
			return err
		}
	} else if err != nil {
		return err
	} else {
		creds.Identifiers = append(creds.Identifiers, identity.OIDCUniqueID(provider, subject))
		conf.Providers = append(conf.Providers, identity.CredentialsOIDCProvider{
			Subject: subject, Provider: provider,
			InitialAccessToken:  accessToken,
			InitialRefreshToken: refreshToken,
			InitialIDToken:      idToken,
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
