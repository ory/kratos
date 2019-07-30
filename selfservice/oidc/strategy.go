package oidc

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/coreos/go-oidc"
	"github.com/google/uuid"
	"github.com/julienschmidt/httprouter"
	"github.com/pkg/errors"

	"github.com/ory/x/jsonx"

	"github.com/ory/gojsonschema"
	"github.com/ory/herodot"
	"github.com/ory/x/stringsx"
	"github.com/ory/x/urlx"

	"github.com/ory/hive/driver/configuration"
	"github.com/ory/hive/errorx"
	"github.com/ory/hive/identity"
	"github.com/ory/hive/schema"
	"github.com/ory/hive/selfservice"
	"github.com/ory/hive/session"
	"github.com/ory/hive/x"
)

const (
	BasePath     = "/methods/oidc"
	AuthPath     = BasePath + "/auth/:provider/:request"
	CallbackPath = BasePath + "/callback/:provider"
)

var _ selfservice.Strategy = new(Strategy)

type dependencies interface {
	errorx.ManagementProvider
	x.CookieProvider
	identity.ValidationProvider
	identity.PoolProvider
	x.LoggingProvider
	session.ManagementProvider

	selfservice.RegistrationExecutionProvider
	selfservice.LoginExecutionProvider
	selfservice.LoginRequestManagementProvider
	selfservice.RegistrationRequestManagementProvider

	selfservice.PostRegistrationHookProvider
	selfservice.StrategyHandlerProvider
	selfservice.PostLoginHookProvider
	selfservice.RequestErrorHandlerProvider
}

// Strategy implements selfservice.LoginStrategy, selfservice.RegistrationStrategy. It supports both login
// and registration via OpenID Providers.
type Strategy struct {
	c configuration.Provider
	d dependencies

	validator *schema.Validator
}

func (s *Strategy) SetRoutes(r *x.RouterPublic) {
	if _, _, ok := r.Lookup("GET", CallbackPath); !ok {
		r.GET(CallbackPath, s.handleCallback)
	}

	if _, _, ok := r.Lookup("GET", AuthPath); !ok {
		r.GET(AuthPath, s.handleAuth)
	}
}

func NewStrategy(
	d dependencies,
	c configuration.Provider,
) *Strategy {
	return &Strategy{
		c:         c,
		d:         d,
		validator: schema.NewValidator(),
	}
}

func (s *Strategy) ID() identity.CredentialsType {
	return CredentialsType
}

func (s *Strategy) handleAuth(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var (
		pid = ps.ByName("provider")
		rid = ps.ByName("request")
	)

	provider, err := s.provider(pid)
	if err != nil {
		s.handleError(w, r, rid, err)
		return
	}

	config, err := provider.OAuth2(r.Context())
	if err != nil {
		s.handleError(w, r, rid, err)
		return
	}

	if _, err := s.validateRequest(r.Context(), rid); err != nil {
		s.handleError(w, r, rid, err)
		return
	}

	state := uuid.New().String()
	if err := x.SessionPersistValues(w, r, s.d.CookieManager(), sessionName, map[string]interface{}{
		sessionKeyState:  state,
		sessionRequestID: rid,
	}); err != nil {
		s.handleError(w, r, rid, err)
		return
	}

	http.Redirect(w, r, config.AuthCodeURL(state), http.StatusFound)
}

func (s *Strategy) validateRequest(ctx context.Context, rid string) (request, error) {
	if rid == "" {
		return nil, errors.WithStack(herodot.ErrBadRequest.WithReason("The session cookie contains invalid values and the request could not be executed. Please try again."))
	}

	if ar, err := s.d.RegistrationRequestManager().GetRegistrationRequest(ctx, rid); err == nil {
		if err := ar.Valid(); err != nil {
			return nil, err
		}
		return ar, nil
	}

	ar, err := s.d.LoginRequestManager().GetLoginRequest(ctx, rid)
	if err != nil {
		return nil, err
	}

	if err := ar.Valid(); err != nil {
		return nil, err
	}

	return ar, nil
}

func (s *Strategy) validateCallback(r *http.Request) (request, error) {
	var (
		code = r.URL.Query().Get("code")
	)
	if state := r.URL.Query().Get("state"); state == "" {
		return nil, errors.WithStack(herodot.ErrBadRequest.WithReasonf(`Unable to complete OpenID Connect flow because the OpenID Provider did not return the state query parameter.`))
	} else if state != x.SessionGetStringOr(r, s.d.CookieManager(), sessionName, sessionKeyState, "") {
		return nil, errors.WithStack(herodot.ErrBadRequest.WithReasonf(`Unable to complete OpenID Connect flow because the query state parameter does not match the state parameter from the session cookie.`))
	}

	ar, err := s.validateRequest(r.Context(), x.SessionGetStringOr(r, s.d.CookieManager(), sessionName, sessionRequestID, ""))
	if err != nil {
		return nil, err
	}

	if r.URL.Query().Get("error") != "" {
		return ar, errors.WithStack(herodot.ErrBadRequest.WithReasonf(`Unable to complete OpenID Connect flow because the OpenID Provider returned error "%s": %s`, r.URL.Query().Get("error"), r.URL.Query().Get("error_description")))
	}

	if code == "" {
		return ar, errors.WithStack(herodot.ErrBadRequest.WithReasonf(`Unable to complete OpenID Connect flow because the OpenID Provider did not return the code query parameter.`))
	}

	return ar, nil
}

func (s *Strategy) handleCallback(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	var (
		code = r.URL.Query().Get("code")
		pid  = ps.ByName("provider")
	)

	ar, err := s.validateCallback(r)
	if err != nil {
		if ar != nil {
			s.handleError(w, r, ar.GetID(), err)
		} else {
			s.handleError(w, r, "", err)
		}
		return
	}

	provider, err := s.provider(pid)
	if err != nil {
		s.handleError(w, r, ar.GetID(), err)
		return
	}

	config, err := provider.OAuth2(context.Background())
	if err != nil {
		s.handleError(w, r, ar.GetID(), err)
		return
	}

	token, err := config.Exchange(r.Context(), code)
	if err != nil {
		s.handleError(w, r, ar.GetID(), err)
		return
	}

	claims, err := provider.Claims(r.Context(), token)
	if err != nil {
		s.handleError(w, r, ar.GetID(), err)
		return
	}

	switch a := ar.(type) {
	case *selfservice.LoginRequest:
		err = s.processLogin(w, r, a, claims, provider)
	case *selfservice.RegistrationRequest:
		err = s.processRegistration(w, r, a, claims, provider)
	default:
		panic(fmt.Sprintf("unexpected type: %T", a))
	}

	if err != nil {
		s.handleError(w, r, ar.GetID(), err)
		return
	}

	return
}

func uid(provider, subject string) string {
	return fmt.Sprintf("%s:%s", provider, subject)
}

func (s *Strategy) authURL(request, provider string) string {
	return urlx.AppendPaths(
		urlx.Copy(s.c.SelfPublicURL()),
		strings.Replace(
			strings.Replace(
				AuthPath, ":request", request, 1,
			),
			":provider", provider, 1),
	).String()
}

func (s *Strategy) processLogin(w http.ResponseWriter, r *http.Request, a *selfservice.LoginRequest, claims *Claims, provider Provider) error {
	i, c, err := s.d.IdentityPool().FindByCredentialsIdentifier(r.Context(), CredentialsType, uid(provider.Config().ID, claims.Subject))
	if err != nil {
		if errors.Cause(err).Error() == herodot.ErrNotFound.Error() {
			// If no account was found we're "manually" creating a new registration request and redirecting the browser
			// to that endpoint.

			// That will execute the "pre registration" hook which allows to e.g. disallow this request. The registration
			// ui however will NOT be shown, instead the user is directly redirected to the auth path. That should then
			// do a silent re-request. While this might be a bit excessive from a network perspective it should usually
			// happen without any downsides to user experience as the request has already been authorized and should
			// not need additional consent/login.

			// This is kinda hacky but the only way to ensure seamless login/registration flows when using OIDC.

			s.d.Logger().WithField("provider", provider.Config().ID).WithField("subject", claims.Subject).Debug("Received successful OpenID Connect callback but user is not registered. Re-initializing registration flow now.")
			return s.d.StrategyHandler().NewRegistrationRequest(w, r, func(aa *selfservice.RegistrationRequest) string {
				return s.authURL(aa.ID, provider.Config().ID)
			})
		}
		return err
	}

	var o CredentialsConfig
	if err := json.NewDecoder(bytes.NewBuffer(c.Options)).Decode(&o); err != nil {
		return errors.WithStack(herodot.ErrInternalServerError.WithReason("The password credentials could not be decoded properly").WithDebug(err.Error()))
	}

	if o.Subject != claims.Subject {
		return errors.WithStack(herodot.ErrInternalServerError.WithReason("The subjects do not match").WithDebugf("Expected credential subject to match subject from RequestID Token but values are not equal: %s != %s", o.Subject, claims.Subject))
	} else if o.Provider != provider.Config().ID {
		return errors.WithStack(herodot.ErrInternalServerError.WithReason("The providers do not match").WithDebugf("Expected credential provider to match provider from path but values are not equal: %s != %s", o.Subject, provider.Config().ID))
	}

	if err = s.d.LoginExecutor().PostLoginHook(w, r, s.d.PostLoginHooks(CredentialsType), a, i); err != nil {
		return err
	}

	return nil
}

func (s *Strategy) processRegistration(w http.ResponseWriter, r *http.Request, a *selfservice.RegistrationRequest, claims *Claims, provider Provider) error {
	if _, _, err := s.d.IdentityPool().FindByCredentialsIdentifier(r.Context(), CredentialsType, uid(provider.Config().ID, claims.Subject)); err == nil {
		// If the identity already exists, we should perform the login flow instead.

		// That will execute the "pre login" hook which allows to e.g. disallow this request. The login
		// ui however will NOT be shown, instead the user is directly redirected to the auth path. That should then
		// do a silent re-request. While this might be a bit excessive from a network perspective it should usually
		// happen without any downsides to user experience as the request has already been authorized and should
		// not need additional consent/login.

		// This is kinda hacky but the only way to ensure seamless login/registration flows when using OIDC.
		s.d.Logger().WithField("provider", provider.Config().ID).WithField("subject", claims.Subject).Debug("Received successful OpenID Connect callback but user is already registered. Re-initializing login flow now.")
		return s.d.StrategyHandler().NewLoginRequest(w, r, func(aa *selfservice.LoginRequest) string {
			return s.authURL(aa.ID, provider.Config().ID)
		})
	}

	i := identity.NewIdentity(s.c.DefaultIdentityTraitsSchemaURL().String())

	// Validate the claims first (which will also copy the values around based on the schema)
	if err := s.validator.Validate(
		stringsx.Coalesce(
			provider.Config().SchemaURL,
		),
		gojsonschema.NewGoLoader(claims),
		NewValidationExtension().WithIdentity(i),
	); err != nil {
		s.d.Logger().
			WithField("provider", provider.Config().ID).
			WithField("schema_url", provider.Config().SchemaURL).
			WithField("claims", fmt.Sprintf("%+v", claims)).
			Error("Unable to validate claims against provider schema. Your schema should work regardless of these values.")
		// Force a system error because this can not be resolved by the user.
		return errors.WithStack(herodot.ErrInternalServerError.WithTrace(err).WithReasonf("%s", err))
	}

	// Validate the identity itself
	if err := s.d.IdentityValidator().Validate(i); err != nil {
		return err
	}

	var b bytes.Buffer
	if err := json.NewEncoder(&b).Encode(&CredentialsConfig{
		Subject:  claims.Subject,
		Provider: provider.Config().ID,
	}); err != nil {
		return errors.WithStack(herodot.ErrInternalServerError.WithReasonf("Unable to encode password options to JSON: %s", err))
	}

	i.SetCredentials(s.ID(), identity.Credentials{
		ID:          s.ID(),
		Identifiers: []string{uid(provider.Config().ID, claims.Subject)},
		Options:     b.Bytes(),
	})

	if err := s.d.RegistrationExecutor().PostRegistrationHook(w, r, s.d.PostRegistrationHooks(CredentialsType), a, i); err != nil {
		return err
	}

	return nil
}

func (s *Strategy) verifyIdentity(i *identity.Identity, c identity.Credentials, token oidc.IDToken, pid string) error {
	var o CredentialsConfig

	if err := json.NewDecoder(bytes.NewBuffer(c.Options)).Decode(&o); err != nil {
		return errors.WithStack(herodot.ErrInternalServerError.WithReason("The password credentials could not be decoded properly").WithDebug(err.Error()))
	}

	if o.Subject != token.Subject {
		return errors.WithStack(herodot.ErrInternalServerError.WithReason("The subjects do not match").WithDebugf("Expected credential subject to match subject from RequestID Token but values are not equal: %s != %s", o.Subject, token.Subject))
	} else if o.Provider != pid {
		return errors.WithStack(herodot.ErrInternalServerError.WithReason("The providers do not match").WithDebugf("Expected credential provider to match provider from path but values are not equal: %s != %s", o.Subject, pid))
	}

	return nil
}

func (s *Strategy) populateMethod(r *http.Request, request string) (*RequestMethodConfig, error) {
	conf, err := s.Config()
	if err != nil {
		return nil, err
	}

	sc := NewRequestMethodConfig()
	for _, provider := range conf.Providers {
		sc.Providers = append(sc.Providers, RequestMethodConfigProvider{
			ID:  provider.ID,
			URL: s.authURL(request, provider.ID),
		})
	}

	return sc, nil
}

func (s *Strategy) PopulateLoginMethod(r *http.Request, sr *selfservice.LoginRequest) error {
	config, err := s.populateMethod(r, sr.ID)
	if err != nil {
		return err
	}
	sr.Methods[CredentialsType] = &selfservice.LoginRequestMethod{
		Method: CredentialsType,
		Config: config,
	}
	return nil
}

func (s *Strategy) PopulateRegistrationMethod(r *http.Request, sr *selfservice.RegistrationRequest) error {
	config, err := s.populateMethod(r, sr.ID)
	if err != nil {
		return err
	}
	sr.Methods[CredentialsType] = &selfservice.RegistrationRequestMethod{
		Method: CredentialsType,
		Config: config,
	}
	return nil
}

func (s *Strategy) Config() (*ConfigurationCollection, error) {
	var c ConfigurationCollection

	if err := jsonx.
		NewStrictDecoder(
			bytes.NewBuffer(s.c.SelfServiceStrategy(string(CredentialsType)).Config),
		).
		Decode(&c); err != nil {
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("Unable to decode OpenID Connect Provider configuration: %s", err))
	}

	return &c, nil
}

func (s *Strategy) provider(id string) (Provider, error) {
	if c, err := s.Config(); err != nil {
		return nil, err
	} else if provider, err := c.Provider(id, s.c.SelfPublicURL()); err != nil {
		return nil, err
	} else {
		return provider, nil
	}
}

func (s *Strategy) handleError(w http.ResponseWriter, r *http.Request, rid string, err error) {
	if rid == "" {
		s.d.ErrorManager().ForwardError(w, r, err)
		return
	}

	if lr, rerr := s.d.LoginRequestManager().GetLoginRequest(r.Context(), rid); rerr == nil {
		s.d.SelfServiceRequestErrorHandler().HandleLoginError(w, r, CredentialsType, lr, err, nil)
		return
	} else if rr, rerr := s.d.RegistrationRequestManager().GetRegistrationRequest(r.Context(), rid); rerr == nil {
		s.d.SelfServiceRequestErrorHandler().HandleRegistrationError(w, r, CredentialsType, rr, err, nil)
		return
	}

	s.d.ErrorManager().ForwardError(w, r, err)
	return
}
