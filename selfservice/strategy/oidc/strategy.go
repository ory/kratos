package oidc

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gofrs/uuid"
	"github.com/julienschmidt/httprouter"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"

	"github.com/ory/x/fetcher"

	"github.com/ory/x/jsonx"

	"github.com/ory/herodot"
	"github.com/ory/x/urlx"

	"github.com/ory/kratos/continuity"
	"github.com/ory/kratos/driver/configuration"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/schema"
	"github.com/ory/kratos/selfservice/errorx"
	"github.com/ory/kratos/selfservice/flow/login"
	"github.com/ory/kratos/selfservice/flow/registration"
	"github.com/ory/kratos/selfservice/flow/settings"
	"github.com/ory/kratos/selfservice/form"
	"github.com/ory/kratos/session"
	"github.com/ory/kratos/x"
)

const (
	BasePath = "/self-service/browser/flows/strategies/oidc"

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

	settings.ErrorHandlerProvider
	settings.RequestPersistenceProvider
	settings.HookExecutorProvider

	continuity.ManagementProvider

	identity.ActiveCredentialsCounterStrategyProvider
}

func isForced(req interface{}) bool {
	f, ok := req.(interface {
		IsForced() bool
	})
	return ok && f.IsForced()
}

// Strategy implements selfservice.LoginStrategy, selfservice.RegistrationStrategy. It supports both login
// and registration via OpenID Providers.
type Strategy struct {
	c         configuration.Provider
	d         dependencies
	f         *fetcher.Fetcher
	validator *schema.Validator
}

type authCodeContainer struct {
	RequestID string     `json:"request_id"`
	State     string     `json:"state"`
	Form      url.Values `json:"form"`
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

	var pid = r.Form.Get("provider") // this can come from both url query and post body
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

	req, err := s.validateRequest(r.Context(), r, rid)
	if err != nil {
		s.handleError(w, r, rid, pid, nil, err)
		return
	}

	if s.alreadyAuthenticated(w, r, req) {
		return
	}

	state := x.NewUUID().String()
	if err := s.d.ContinuityManager().Pause(r.Context(), w, r, sessionName,
		continuity.WithPayload(&authCodeContainer{
			State:     state,
			RequestID: rid.String(),
			Form:      r.PostForm,
		}),
		continuity.WithLifespan(time.Minute*30)); err != nil {
		s.handleError(w, r, rid, pid, nil, err)
		return
	}

	http.Redirect(w, r, config.AuthCodeURL(state, provider.AuthCodeURLOptions(req)...), http.StatusFound)
}

func (s *Strategy) validateRequest(ctx context.Context, r *http.Request, rid uuid.UUID) (request, error) {
	if x.IsZeroUUID(rid) {
		return nil, errors.WithStack(herodot.ErrBadRequest.WithReason("The session cookie contains invalid values and the request could not be executed. Please try again."))
	}

	if ar, err := s.d.RegistrationRequestPersister().GetRegistrationRequest(ctx, rid); err == nil {
		if err := ar.Valid(); err != nil {
			return ar, err
		}
		return ar, nil
	}

	if ar, err := s.d.LoginRequestPersister().GetLoginRequest(ctx, rid); err == nil {
		if err := ar.Valid(); err != nil {
			return ar, err
		}
		return ar, nil
	}

	ar, err := s.d.SettingsRequestPersister().GetSettingsRequest(ctx, rid)
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

func (s *Strategy) validateCallback(w http.ResponseWriter, r *http.Request) (request, *authCodeContainer, error) {
	var (
		code  = r.URL.Query().Get("code")
		state = r.URL.Query().Get("state")
	)

	if state == "" {
		return nil, nil, errors.WithStack(herodot.ErrBadRequest.WithReasonf(`Unable to complete OpenID Connect flow because the OpenID Provider did not return the state query parameter.`))
	}

	var container authCodeContainer
	if _, err := s.d.ContinuityManager().Continue(context.Background(), w, r, sessionName, continuity.WithPayload(&container)); err != nil {
		return nil, nil, err
	}

	if state != container.State {
		return nil, &container, errors.WithStack(herodot.ErrBadRequest.WithReasonf(`Unable to complete OpenID Connect flow because the query state parameter does not match the state parameter from the session cookie.`))
	}

	req, err := s.validateRequest(r.Context(), r, x.ParseUUID(container.RequestID))
	if err != nil {
		return nil, &container, err
	}

	if r.URL.Query().Get("error") != "" {
		return req, &container, errors.WithStack(herodot.ErrBadRequest.WithReasonf(`Unable to complete OpenID Connect flow because the OpenID Provider returned error "%s": %s`, r.URL.Query().Get("error"), r.URL.Query().Get("error_description")))
	}

	if code == "" {
		return req, &container, errors.WithStack(herodot.ErrBadRequest.WithReasonf(`Unable to complete OpenID Connect flow because the OpenID Provider did not return the code query parameter.`))
	}

	return req, &container, nil
}

func (s *Strategy) alreadyAuthenticated(w http.ResponseWriter, r *http.Request, req interface{}) bool {
	// we assume an error means the user has no session
	if _, err := s.d.SessionManager().FetchFromRequest(r.Context(), r); err == nil {
		if _, ok := req.(*settings.Request); ok {
			// ignore this if it's a settings request
		} else if !isForced(req) {
			http.Redirect(w, r, s.c.DefaultReturnToURL().String(), http.StatusFound)
			return true
		}
	}

	return false
}

func (s *Strategy) handleCallback(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var (
		code = r.URL.Query().Get("code")
		pid  = ps.ByName("provider")
	)

	req, container, err := s.validateCallback(w, r)
	if err != nil {
		if req != nil {
			s.handleError(w, r, req.GetID(), pid, nil, err)
		} else {
			s.handleError(w, r, x.EmptyUUID, pid, nil, err)
		}
		return
	}

	if s.alreadyAuthenticated(w, r, req) {
		return
	}

	provider, err := s.provider(pid)
	if err != nil {
		s.handleError(w, r, req.GetID(), pid, nil, err)
		return
	}

	config, err := provider.OAuth2(context.Background())
	if err != nil {
		s.handleError(w, r, req.GetID(), pid, nil, err)
		return
	}

	token, err := config.Exchange(r.Context(), code)
	if err != nil {
		s.handleError(w, r, req.GetID(), pid, nil, err)
		return
	}

	claims, err := provider.Claims(r.Context(), token)
	if err != nil {
		s.handleError(w, r, req.GetID(), pid, nil, err)
		return
	}

	switch a := req.(type) {
	case *login.Request:
		s.processLogin(w, r, a, claims, provider, container)
		return
	case *registration.Request:
		s.processRegistration(w, r, a, claims, provider, container)
		return
	case *settings.Request:
		sess, err := s.d.SessionManager().FetchFromRequest(r.Context(), r)
		if err != nil {
			s.handleError(w, r, req.GetID(), pid, nil, err)
			return
		}
		s.linkProvider(w, r, &settings.UpdateContext{Session: sess, Request: a}, claims, provider)
		return
	default:
		s.handleError(w, r, req.GetID(), pid, nil, errors.WithStack(x.PseudoPanic.
			WithDetailf("cause", "Unexpected type in OpenID Connect flow: %T", a)))
		return
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
			bytes.NewBuffer(s.c.SelfServiceStrategy(string(s.ID())).Config),
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
		s.d.LoginRequestErrorHandler().HandleLoginError(w, r, s.ID(), lr, err)
		return
	} else if sr, rerr := s.d.SettingsRequestPersister().GetSettingsRequest(r.Context(), rid); rerr == nil {
		s.d.SettingsRequestErrorHandler().HandleSettingsError(w, r, sr, err, s.SettingsStrategyID())
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
				s.d.RegistrationRequestErrorHandler().HandleRegistrationError(w, r, s.ID(), rr, errors.Wrap(err, errSec.Error()))
				return
			}
			method.Config.ResetErrors()

			method.Config.SetCSRF(s.d.GenerateCSRFToken(r))
			if errSec := method.Config.SortFields(s.c.DefaultIdentityTraitsSchemaURL().String(), "traits"); errSec != nil {
				s.d.RegistrationRequestErrorHandler().HandleRegistrationError(w, r, s.ID(), rr, errors.Wrap(err, errSec.Error()))
				return
			}

			method.Config.UnsetField("provider")
			method.Config.SetField(form.Field{Name: "provider", Value: provider, Type: "submit"})
			rr.Methods[s.ID()] = method
		}

		s.d.RegistrationRequestErrorHandler().HandleRegistrationError(w, r, s.ID(), rr, err)
		return
	}

	s.d.SelfServiceErrorManager().Forward(r.Context(), w, r, err)
}
