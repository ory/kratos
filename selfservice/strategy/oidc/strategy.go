package oidc

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/ory/kratos/text"

	"github.com/ory/kratos/ui/container"
	"github.com/ory/x/decoderx"

	"github.com/ory/kratos/ui/node"

	"github.com/gofrs/uuid"
	"github.com/julienschmidt/httprouter"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"

	"github.com/ory/x/jsonx"

	"github.com/ory/x/fetcher"

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
	dec       *decoderx.HTTP
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

func (s *Strategy) validateFlow(ctx context.Context, r *http.Request, rid uuid.UUID) (flow.Flow, error) {
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

func (s *Strategy) validateCallback(w http.ResponseWriter, r *http.Request) (flow.Flow, *authCodeContainer, error) {
	var (
		code  = r.URL.Query().Get("code")
		state = r.URL.Query().Get("state")
	)

	if state == "" {
		return nil, nil, errors.WithStack(herodot.ErrBadRequest.WithReasonf(`Unable to complete OpenID Connect flow because the OpenID Provider did not return the state query parameter.`))
	}

	var cntnr authCodeContainer
	if _, err := s.d.ContinuityManager().Continue(r.Context(), w, r, sessionName, continuity.WithPayload(&cntnr)); err != nil {
		return nil, nil, err
	}

	if state != cntnr.State {
		return nil, &cntnr, errors.WithStack(herodot.ErrBadRequest.WithReasonf(`Unable to complete OpenID Connect flow because the query state parameter does not match the state parameter from the session cookie.`))
	}

	req, err := s.validateFlow(r.Context(), r, x.ParseUUID(cntnr.FlowID))
	if err != nil {
		return nil, &cntnr, err
	}

	if r.URL.Query().Get("error") != "" {
		return req, &cntnr, errors.WithStack(herodot.ErrBadRequest.WithReasonf(`Unable to complete OpenID Connect flow because the OpenID Provider returned error "%s": %s`, r.URL.Query().Get("error"), r.URL.Query().Get("error_description")))
	}

	if code == "" {
		return req, &cntnr, errors.WithStack(herodot.ErrBadRequest.WithReasonf(`Unable to complete OpenID Connect flow because the OpenID Provider did not return the code query parameter.`))
	}

	return req, &cntnr, nil
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

	req, cntnr, err := s.validateCallback(w, r)
	if err != nil {
		if req != nil {
			s.forwardError(w, r, req, s.handleError(w, r, req, pid, nil, err))
		} else {
			s.d.SelfServiceErrorManager().Forward(r.Context(), w, r, s.handleError(w, r, nil, pid, nil, err))
		}
		return
	}

	if s.alreadyAuthenticated(w, r, req) {
		return
	}

	provider, err := s.provider(r.Context(), r, pid)
	if err != nil {
		s.forwardError(w, r, req, s.handleError(w, r, req, pid, nil, err))
		return
	}

	conf, err := provider.OAuth2(context.Background())
	if err != nil {
		s.forwardError(w, r, req, s.handleError(w, r, req, pid, nil, err))
		return
	}

	token, err := conf.Exchange(r.Context(), code)
	if err != nil {
		s.forwardError(w, r, req, s.handleError(w, r, req, pid, nil, err))
		return
	}

	claims, err := provider.Claims(r.Context(), token)
	if err != nil {
		s.forwardError(w, r, req, s.handleError(w, r, req, pid, nil, err))
		return
	}

	switch a := req.(type) {
	case *login.Flow:
		if ff, err := s.processLogin(w, r, a, claims, provider, cntnr); err != nil {
			if ff != nil {
				s.forwardError(w, r, ff, err)
				return
			}
			s.forwardError(w, r, a, err)
		}
		return
	case *registration.Flow:
		if ff, err := s.processRegistration(w, r, a, claims, provider, cntnr); err != nil {
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
		if err := s.linkProvider(w, r, &settings.UpdateContext{Session: sess, Flow: a}, claims, provider); err != nil {
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

func uid(provider, subject string) string {
	return fmt.Sprintf("%s:%s", provider, subject)
}

func (s *Strategy) populateMethod(r *http.Request, c *container.Container, message func(provider string) *text.Message) error {
	conf, err := s.Config(r.Context())
	if err != nil {
		return err
	}

	// does not need sorting because there is only one field
	c.SetCSRF(s.d.GenerateCSRFToken(r))
	AddProviders(c, conf.Providers, message)

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

func (s *Strategy) handleError(w http.ResponseWriter, r *http.Request, f flow.Flow, provider string, traits []byte, err error) error {
	switch rf := f.(type) {
	case *login.Flow:
		return err
	case *registration.Flow:
		// Reset all nodes to not confuse users.
		// This is kinda hacky and will probably need to be updated at some point.
		var nodes node.Nodes
		for _, n := range rf.UI.Nodes {
			if n.Group != node.DefaultGroup && n.Group != node.OpenIDConnectGroup {
				continue
			}

			nodes = append(nodes, n)
		}

		rf.UI.Nodes = nodes
		if traits != nil {
			rf.UI.UpdateNodesFromJSON(traits, "traits", node.OpenIDConnectGroup)
		}
		return err
	case *settings.Flow:
		return err
	}

	return err
}

func (s *Strategy) NodeGroup() node.Group {
	return node.OpenIDConnectGroup
}
