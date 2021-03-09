package oidc

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/ory/kratos/ui/container"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/ory/kratos/ui/node"

	"github.com/gofrs/uuid"
	"github.com/julienschmidt/httprouter"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"

	"github.com/ory/x/jsonx"

	"github.com/ory/x/fetcher"

	"github.com/ory/herodot"
	"github.com/ory/x/urlx"

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

	RouteAuth     = RouteBase + "/auth/:flow"
	RouteCallback = RouteBase + "/callback/:provider"
)

var _ identity.ActiveCredentialsCounter = new(Strategy)

type dependencies interface {
	errorx.ManagementProvider

	config.Provider

	x.LoggingProvider
	x.CookieProvider
	x.CSRFTokenGeneratorProvider
	x.WriterProvider

	identity.ValidationProvider
	identity.PrivilegedPoolProvider
	identity.ActiveCredentialsCounterStrategyProvider

	session.ManagementProvider
	session.HandlerProvider

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
	d         dependencies
	f         *fetcher.Fetcher
	validator *schema.Validator
}

func (s *Strategy) Login(w http.ResponseWriter, r *http.Request, f *login.Flow) (i *identity.Identity, err error) {
	panic("implement me")
}

type authCodeContainer struct {
	FlowID string     `json:"flow_id"`
	State  string     `json:"state"`
	Form   url.Values `json:"form"`
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
	wrappedHandleCallback := strategy.IsDisabled(s.d, s.ID().String(), s.handleCallback)
	if handle, _, _ := r.Lookup("GET", RouteCallback); handle == nil {
		r.GET(RouteCallback, wrappedHandleCallback)
	}

	wrappedHandleAuth := strategy.IsDisabled(s.d, s.ID().String(), s.handleAuth)
	if handle, _, _ := r.Lookup("POST", RouteAuth); handle == nil {
		r.POST(RouteAuth, wrappedHandleAuth)
	}

	if handle, _, _ := r.Lookup("GET", RouteAuth); handle == nil {
		r.GET(RouteAuth, wrappedHandleAuth)
	}
}

func NewStrategy(d dependencies) *Strategy {
	return &Strategy{
		d:         d,
		f:         fetcher.NewFetcher(),
		validator: schema.NewValidator(),
	}
}

func (s *Strategy) ID() identity.CredentialsType {
	return identity.CredentialsTypeOIDC
}

func (s *Strategy) handleAuth(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	rid := x.ParseUUID(ps.ByName("flow"))
	if err := r.ParseForm(); err != nil {
		s.handleError(w, r, rid, "", nil, errors.WithStack(herodot.ErrBadRequest.WithDebug(err.Error()).WithReasonf("Unable to parse HTTP form request: %s", err.Error())))
		return
	}

	var pid = r.Form.Get("provider") // this can come from both url query and post body
	if pid == "" {
		s.handleError(w, r, rid, pid, nil, errors.WithStack(herodot.ErrBadRequest.WithReasonf(`The HTTP request did not contain the required "provider" form field`)))
		return
	}

	provider, err := s.provider(r.Context(), r, pid)
	if err != nil {
		s.handleError(w, r, rid, pid, nil, err)
		return
	}

	config, err := provider.OAuth2(r.Context())
	if err != nil {
		s.handleError(w, r, rid, pid, nil, err)
		return
	}

	req, err := s.validateFlow(r.Context(), r, rid)
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
			State:  state,
			FlowID: rid.String(),
			Form:   r.PostForm,
		}),
		continuity.WithLifespan(time.Minute*30)); err != nil {
		s.handleError(w, r, rid, pid, nil, err)
		return
	}

	http.Redirect(w, r, config.AuthCodeURL(state, provider.AuthCodeURLOptions(req)...), http.StatusFound)
}

func (s *Strategy) validateFlow(ctx context.Context, r *http.Request, rid uuid.UUID) (ider, error) {
	if x.IsZeroUUID(rid) {
		return nil, errors.WithStack(herodot.ErrBadRequest.WithReason("The session cookie contains invalid values and the flow could not be executed. Please try again."))
	}

	if ar, err := s.d.RegistrationFlowPersister().GetRegistrationFlow(ctx, rid); err == nil {
		if ar.Type != flow.TypeBrowser {
			return ar, ErrAPIFlowNotSupported
		}

		if err := ar.Valid(); err != nil {
			return ar, err
		}
		return ar, nil
	}

	if ar, err := s.d.LoginFlowPersister().GetLoginFlow(ctx, rid); err == nil {
		if ar.Type != flow.TypeBrowser {
			return ar, ErrAPIFlowNotSupported
		}

		if err := ar.Valid(); err != nil {
			return ar, err
		}
		return ar, nil
	}

	ar, err := s.d.SettingsFlowPersister().GetSettingsFlow(ctx, rid)
	if err == nil {
		if ar.Type != flow.TypeBrowser {
			return ar, ErrAPIFlowNotSupported
		}

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

func (s *Strategy) validateCallback(w http.ResponseWriter, r *http.Request) (ider, *authCodeContainer, error) {
	var (
		code  = r.URL.Query().Get("code")
		state = r.URL.Query().Get("state")
	)

	if state == "" {
		return nil, nil, errors.WithStack(herodot.ErrBadRequest.WithReasonf(`Unable to complete OpenID Connect flow because the OpenID Provider did not return the state query parameter.`))
	}

	var container authCodeContainer
	if _, err := s.d.ContinuityManager().Continue(r.Context(), w, r, sessionName, continuity.WithPayload(&container)); err != nil {
		return nil, nil, err
	}

	if state != container.State {
		return nil, &container, errors.WithStack(herodot.ErrBadRequest.WithReasonf(`Unable to complete OpenID Connect flow because the query state parameter does not match the state parameter from the session cookie.`))
	}

	req, err := s.validateFlow(r.Context(), r, x.ParseUUID(container.FlowID))
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
		if _, ok := req.(*settings.Flow); ok {
			// ignore this if it's a settings flow
		} else if !isForced(req) {
			http.Redirect(w, r, s.d.Config(r.Context()).SelfServiceBrowserDefaultReturnTo().String(), http.StatusFound)
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

	provider, err := s.provider(r.Context(), r, pid)
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
	case *login.Flow:
		s.processLogin(w, r, a, claims, provider, container)
		return
	case *registration.Flow:
		s.processRegistration(w, r, a, claims, provider, container)
		return
	case *settings.Flow:
		sess, err := s.d.SessionManager().FetchFromRequest(r.Context(), r)
		if err != nil {
			s.handleError(w, r, req.GetID(), pid, nil, err)
			return
		}
		s.linkProvider(w, r, &settings.UpdateContext{Session: sess, Flow: a}, claims, provider)
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

func (s *Strategy) authURL(ctx context.Context, r *http.Request, flowID uuid.UUID) string {
	return urlx.AppendPaths(
		urlx.Copy(s.d.Config(ctx).SelfPublicURL(r)),
		strings.Replace(
			RouteAuth, ":flow", flowID.String(), 1,
		),
	).String()
}

func (s *Strategy) populateMethod(r *http.Request, c *container.Container) error {
	conf, err := s.Config(r.Context())
	if err != nil {
		return err
	}

	// does not need sorting because there is only one field
	c.SetCSRF(s.d.GenerateCSRFToken(r))
	AddProviders(c,conf.Providers)

	return nil
}

func (s *Strategy) Config(ctx context.Context) (*ConfigurationCollection, error) {
	var c ConfigurationCollection

	conf := s.d.Config(ctx).SelfServiceStrategy(string(s.ID())).Config
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
	} else if provider, err := c.Provider(id, s.d.Config(ctx).SelfPublicURL(r)); err != nil {
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

	if lr, rerr := s.d.LoginFlowPersister().GetLoginFlow(r.Context(), rid); rerr == nil {
		s.d.LoginFlowErrorHandler().WriteFlowError(w, r, lr, s.NodeGroup(), err)
		return
	} else if sr, rerr := s.d.SettingsFlowPersister().GetSettingsFlow(r.Context(), rid); rerr == nil {
		sess, sessErr := s.d.SessionManager().FetchFromRequest(r.Context(), r)
		if sessErr != nil {
			s.d.SelfServiceErrorManager().Forward(r.Context(), w, r, sessErr)
			return
		}

		s.d.SettingsFlowErrorHandler().WriteFlowError(w, r, s.SettingsStrategyID(), sr, sess.Identity, err)
		return
	} else if rr, rerr := s.d.RegistrationFlowPersister().GetRegistrationFlow(r.Context(), rid); rerr == nil {
		if method, ok := rr.Methods[s.ID()]; ok {
			method.Config.UnsetNode("provider")
			method.Config.Reset()

			if traits != nil {
				method.Config.UpdateNodesFromJSON(traits, "traits", node.OpenIDConnectGroup)
			}

			if errSec := method.Config.ParseError(node.OpenIDConnectGroup, err); errSec != nil {
				s.d.RegistrationFlowErrorHandler().WriteFlowError(w, r, s.ID(), rr, errSec)
				return
			}
			method.Config.ResetMessages()

			method.Config.SetCSRF(s.d.GenerateCSRFToken(r))
			if errSec := method.Config.SortNodes(s.d.Config(r.Context()).DefaultIdentityTraitsSchemaURL().String()); errSec != nil {
				s.d.RegistrationFlowErrorHandler().WriteFlowError(w, r, s.ID(), rr, errors.Wrap(err, errSec.Error()))
				return
			}

			method.Config.UnsetNode("provider")
			method.Config.GetNodes().Upsert(
				// v0.5: form.Field{Name: "provider", Value: provider, Type: "submit"}
				node.NewInputField("provider", provider, node.OpenIDConnectGroup, node.InputAttributeTypeSubmit),
			)
			rr.Methods[s.ID()] = method
		}

		s.d.RegistrationFlowErrorHandler().WriteFlowError(w, r, s.ID(), rr, err)
		return
	}

	s.d.SelfServiceErrorManager().Forward(r.Context(), w, r, err)
}

func (s *Strategy) NodeGroup() node.Group {
	return node.OpenIDConnectGroup
}
