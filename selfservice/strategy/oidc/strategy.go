package oidc

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gofrs/uuid"
	"github.com/julienschmidt/httprouter"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"

	"github.com/ory/x/fetcher"

	"github.com/ory/x/jsonx"

	"github.com/ory/herodot"
	"github.com/ory/x/urlx"

	"github.com/ory/kratos/selfservice/flow/login"
	"github.com/ory/kratos/selfservice/flow/registration"
	"github.com/ory/kratos/selfservice/form"

	"github.com/ory/kratos/driver/configuration"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/schema"
	"github.com/ory/kratos/selfservice/errorx"
	"github.com/ory/kratos/session"
	"github.com/ory/kratos/x"
)

const (
	BasePath = "/self-service/browser/flows/registration/strategies/oidc"

	AuthPath     = BasePath + "/auth/:request"
	CallbackPath = BasePath + "/callback/:provider"
)

var _ identity.ActiveCredentialsCounter = new(Strategy)

type dependencies interface {
	errorx.ManagementProvider

	x.LoggingProvider
	x.CookieProvider
	x.CSRFTokenGeneratorProvider

	identity.ValidationProvider
	identity.PrivilegedPoolProvider

	session.ManagementProvider
	session.HandlerProvider

	login.HookExecutorProvider
	login.RequestPersistenceProvider
	login.HooksProvider
	login.StrategyProvider
	login.HandlerProvider
	login.ErrorHandlerProvider

	registration.HookExecutorProvider
	registration.RequestPersistenceProvider
	registration.HooksProvider
	registration.StrategyProvider
	registration.HandlerProvider
	registration.ErrorHandlerProvider

	identity.ActiveCredentialsCounterStrategyProvider
}

// Strategy implements selfservice.LoginStrategy, selfservice.RegistrationStrategy. It supports both login
// and registration via OpenID Providers.
type Strategy struct {
	c         configuration.Provider
	d         dependencies
	f         *fetcher.Fetcher
	validator *schema.Validator
}

func (s *Strategy) CountActiveCredentials(cc map[identity.CredentialsType]identity.Credentials) (count int, err error) {
	for _, c := range cc {
		if c.Type == s.ID() && gjson.ValidBytes(c.Config) {
			var conf CredentialsConfig
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

func (s *Strategy) setRoutes(r *x.RouterPublic) {
	if handle, _, _ := r.Lookup("GET", CallbackPath); handle == nil {
		r.GET(CallbackPath, s.handleCallback)
	}

	if handle, _, _ := r.Lookup("POST", AuthPath); handle == nil {
		r.POST(AuthPath, s.handleAuth)
	}

	if handle, _, _ := r.Lookup("GET", AuthPath); handle == nil {
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
		f:         fetcher.NewFetcher(),
		validator: schema.NewValidator(),
	}
}

func (s *Strategy) ID() identity.CredentialsType {
	return identity.CredentialsTypeOIDC
}

func (s *Strategy) handleAuth(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	rid := x.ParseUUID(ps.ByName("request"))

	if err := r.ParseForm(); err != nil {
		s.handleError(w, r, rid, "", nil, errors.WithStack(herodot.ErrBadRequest.WithDebug(err.Error()).WithReasonf("Unable to parse HTTP form request: %s", err.Error())))
		return
	}

	var (
		pid = r.Form.Get("provider") // this can come from both url query and post body
	)

	if pid == "" {
		s.handleError(w, r, rid, pid, nil, errors.WithStack(herodot.ErrBadRequest.WithReasonf(`The HTTP request did not contain the required "provider" form field`)))
		return
	}

	provider, err := s.provider(pid)
	if err != nil {
		s.handleError(w, r, rid, pid, nil, err)
		return
	}

	config, err := provider.OAuth2(r.Context())
	if err != nil {
		s.handleError(w, r, rid, pid, nil, err)
		return
	}

	ar, err := s.validateRequest(r.Context(), rid)
	if err != nil {
		s.handleError(w, r, rid, pid, nil, err)
		return
	}

	// we assume an error means the user has no session
	if _, err := s.d.SessionManager().FetchFromRequest(r.Context(), r); err == nil {
		if !ar.IsForced() {
			http.Redirect(w, r, s.c.DefaultReturnToURL().String(), http.StatusFound)
			return
		}
	}

	state := x.NewUUID().String()
	// Any data that is posted to this endpoint will be used to fill out missing data from the oidc provider.
	if err := x.SessionPersistValues(w, r, s.d.CookieManager(), sessionName, map[string]interface{}{
		sessionKeyState:  state,
		sessionRequestID: rid.String(),
		sessionFormState: r.PostForm.Encode(),
	}); err != nil {
		s.handleError(w, r, rid, pid, nil, err)
		return
	}

	http.Redirect(w, r, config.AuthCodeURL(state, provider.AuthCodeURLOptions(ar)...), http.StatusFound)
}

func (s *Strategy) validateRequest(ctx context.Context, rid uuid.UUID) (request, error) {
	if x.IsZeroUUID(rid) {
		return nil, errors.WithStack(herodot.ErrBadRequest.WithReason("The session cookie contains invalid values and the request could not be executed. Please try again."))
	}

	if ar, err := s.d.RegistrationRequestPersister().GetRegistrationRequest(ctx, rid); err == nil {
		if err := ar.Valid(); err != nil {
			return ar, err
		}
		return ar, nil
	}

	ar, err := s.d.LoginRequestPersister().GetLoginRequest(ctx, rid)
	if err != nil {
		return nil, err
	}

	if err := ar.Valid(); err != nil {
		return ar, err
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

	ar, err := s.validateRequest(r.Context(), x.ParseUUID(x.SessionGetStringOr(r, s.d.CookieManager(), sessionName, sessionRequestID, "")))
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
			s.handleError(w, r, ar.GetID(), pid, nil, err)
		} else {
			s.handleError(w, r, x.EmptyUUID, pid, nil, err)
		}
		return
	}

	// we assume an error means the user has no session
	if _, err := s.d.SessionManager().FetchFromRequest(r.Context(), r); err == nil {
		if !ar.IsForced() {
			http.Redirect(w, r, s.c.DefaultReturnToURL().String(), http.StatusFound)
			return
		}
	}

	provider, err := s.provider(pid)
	if err != nil {
		s.handleError(w, r, ar.GetID(), pid, nil, err)
		return
	}

	config, err := provider.OAuth2(context.Background())
	if err != nil {
		s.handleError(w, r, ar.GetID(), pid, nil, err)
		return
	}

	token, err := config.Exchange(r.Context(), code)
	if err != nil {
		s.handleError(w, r, ar.GetID(), pid, nil, err)
		return
	}

	claims, err := provider.Claims(r.Context(), token)
	if err != nil {
		s.handleError(w, r, ar.GetID(), pid, nil, err)
		return
	}

	switch a := ar.(type) {
	case *login.Request:
		s.processLogin(w, r, a, claims, provider)
		return
	case *registration.Request:
		s.processRegistration(w, r, a, claims, provider)
		return
	default:
		panic(fmt.Sprintf("unexpected type: %T", a))
	}
}

func uid(provider, subject string) string {
	return fmt.Sprintf("%s:%s", provider, subject)
}

func (s *Strategy) authURL(request uuid.UUID) string {
	return urlx.AppendPaths(
		urlx.Copy(s.c.SelfPublicURL()),
		strings.Replace(
			AuthPath, ":request", request.String(), 1,
		),
	).String()
}

func (s *Strategy) populateMethod(r *http.Request, request uuid.UUID) (*RequestMethod, error) {
	conf, err := s.Config()
	if err != nil {
		return nil, err
	}

	f := form.NewHTMLForm(s.authURL(request))
	f.SetCSRF(s.d.GenerateCSRFToken(r))
	// does not need sorting because there is only one field

	return NewRequestMethodConfig(f).AddProviders(conf.Providers), nil
}

func (s *Strategy) Config() (*ConfigurationCollection, error) {
	var c ConfigurationCollection

	if err := jsonx.
		NewStrictDecoder(
			bytes.NewBuffer(s.c.SelfServiceStrategy(string(identity.CredentialsTypeOIDC)).Config),
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

func (s *Strategy) handleError(w http.ResponseWriter, r *http.Request, rid uuid.UUID, provider string, traits []byte, err error) {
	if x.IsZeroUUID(rid) {
		s.d.SelfServiceErrorManager().Forward(r.Context(), w, r, err)
		return
	}

	if lr, rerr := s.d.LoginRequestPersister().GetLoginRequest(r.Context(), rid); rerr == nil {
		s.d.LoginRequestErrorHandler().HandleLoginError(w, r, identity.CredentialsTypeOIDC, lr, err)
		return
	} else if rr, rerr := s.d.RegistrationRequestPersister().GetRegistrationRequest(r.Context(), rid); rerr == nil {
		if method, ok := rr.Methods[s.ID()]; ok {
			method.Config.UnsetField("provider")
			method.Config.Reset()

			if traits != nil {
				for _, field := range form.NewHTMLFormFromJSON("", traits, "traits").Fields {
					method.Config.SetField(field)
				}
			}

			if errSec := method.Config.ParseError(err); errSec != nil {
				s.d.RegistrationRequestErrorHandler().HandleRegistrationError(w, r, identity.CredentialsTypeOIDC, rr, errors.Wrap(err, errSec.Error()))
				return
			}
			method.Config.ResetErrors()

			method.Config.SetCSRF(s.d.GenerateCSRFToken(r))
			if errSec := method.Config.SortFields(s.c.DefaultIdentityTraitsSchemaURL().String(), "traits"); errSec != nil {
				s.d.RegistrationRequestErrorHandler().HandleRegistrationError(w, r, identity.CredentialsTypeOIDC, rr, errors.Wrap(err, errSec.Error()))
				return
			}

			method.Config.SetField(form.Field{Name: "provider", Value: provider, Type: "submit"})
			rr.Methods[s.ID()] = method
		}

		s.d.RegistrationRequestErrorHandler().HandleRegistrationError(w, r, identity.CredentialsTypeOIDC, rr, err)
		return
	}

	s.d.SelfServiceErrorManager().Forward(r.Context(), w, r, err)
}
