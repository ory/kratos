package driver

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/gorilla/sessions"
	"github.com/justinas/nosurf"
	"github.com/sirupsen/logrus"

	"github.com/ory/x/healthx"

	"github.com/ory/x/tracing"

	"github.com/ory/x/logrusx"

	"github.com/ory/kratos/selfservice/flow/login"
	"github.com/ory/kratos/selfservice/flow/logout"
	"github.com/ory/kratos/selfservice/flow/profile"
	"github.com/ory/kratos/selfservice/flow/registration"
	"github.com/ory/kratos/selfservice/hook"

	"github.com/ory/kratos/selfservice/strategy/oidc"

	"github.com/ory/herodot"

	"github.com/ory/kratos/driver/configuration"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/selfservice/errorx"
	password2 "github.com/ory/kratos/selfservice/strategy/password"
	"github.com/ory/kratos/session"
)

type RegistryAbstract struct {
	l              logrus.FieldLogger
	c              configuration.Provider
	nosurf         *nosurf.CSRFHandler
	trc            *tracing.Tracer
	writer         herodot.Writer
	healthxHandler *healthx.Handler

	identityHandler   *identity.Handler
	identityValidator *identity.Validator
	sessionHandler    *session.Handler
	errorHandler      *errorx.Handler
	passwordHasher    password2.Hasher
	passwordValidator password2.Validator
	sessionsStore     sessions.Store

	selfserviceRegistrationExecutor            *registration.HookExecutor
	selfserviceRegistrationHandler             *registration.Handler
	seflserviceRegistrationErrorHandler        *registration.ErrorHandler
	selfserviceRegistrationRequestErrorHandler *registration.ErrorHandler

	selfserviceLoginExecutor            *login.HookExecutor
	selfserviceLoginHandler             *login.Handler
	selfserviceLoginRequestErrorHandler *login.ErrorHandler

	selfserviceProfileManagementHandler          *profile.Handler
	selfserviceProfileRequestRequestErrorHandler *profile.ErrorHandler

	selfserviceLogoutHandler *logout.Handler

	selfserviceStrategies []selfServiceStrategy

	buildVersion string
	buildHash    string
	buildDate    string

	r Registry
}

func (m *RegistryAbstract) WithBuildInfo(version, hash, date string) Registry {
	m.buildVersion = version
	m.buildHash = hash
	m.buildDate = date
	return m.r
}

func (m *RegistryAbstract) BuildVersion() string {
	return m.buildVersion
}

func (m *RegistryAbstract) BuildDate() string {
	return m.buildDate
}

func (m *RegistryAbstract) BuildHash() string {
	return m.buildHash
}

func (m *RegistryAbstract) with(r Registry) *RegistryAbstract {
	m.r = r
	return m
}

func (m *RegistryAbstract) WithLogger(l logrus.FieldLogger) Registry {
	m.l = l
	return m.r
}

func (m *RegistryAbstract) ProfileManagementHandler() *profile.Handler {
	if m.selfserviceProfileManagementHandler == nil {
		m.selfserviceProfileManagementHandler = profile.NewHandler(m.r, m.c)
	}
	return m.selfserviceProfileManagementHandler
}

func (m *RegistryAbstract) ProfileRequestRequestErrorHandler() *profile.ErrorHandler {
	if m.selfserviceProfileRequestRequestErrorHandler == nil {
		m.selfserviceProfileRequestRequestErrorHandler = profile.NewErrorHandler(m.r, m.c)
	}
	return m.selfserviceProfileRequestRequestErrorHandler
}

func (m *RegistryAbstract) LogoutHandler() *logout.Handler {
	if m.selfserviceLogoutHandler == nil {
		m.selfserviceLogoutHandler = logout.NewHandler(m.r, m.c)
	}
	return m.selfserviceLogoutHandler
}

func (m *RegistryAbstract) HealthHandler() *healthx.Handler {
	if m.healthxHandler == nil {
		m.healthxHandler = healthx.NewHandler(m.Writer(), m.BuildVersion(), healthx.ReadyCheckers{
			"database": m.r.Ping,
		})
	}

	return m.healthxHandler
}

func (m *RegistryAbstract) WithCSRFHandler(c *nosurf.CSRFHandler) {
	m.nosurf = c
}

func (m *RegistryAbstract) CSRFHandler() *nosurf.CSRFHandler {
	if m.nosurf == nil {
		panic("csrf handler is not set")
	}
	return m.nosurf
}

func (m *RegistryAbstract) selfServiceStrategies() []selfServiceStrategy {
	if m.selfserviceStrategies == nil {
		m.selfserviceStrategies = []selfServiceStrategy{
			password2.NewStrategy(m.r, m.c),
			oidc.NewStrategy(m.r, m.c),
		}
	}

	return m.selfserviceStrategies
}

func (m *RegistryAbstract) RegistrationStrategies() registration.Strategies {
	strategies := make([]registration.Strategy, len(m.selfServiceStrategies()))
	for i := range strategies {
		strategies[i] = m.selfServiceStrategies()[i]
	}
	return strategies
}

func (m *RegistryAbstract) LoginStrategies() login.Strategies {
	strategies := make([]login.Strategy, len(m.selfServiceStrategies()))
	for i := range strategies {
		strategies[i] = m.selfServiceStrategies()[i]
	}
	return strategies
}

func (m *RegistryAbstract) hooksPost(credentialsType identity.CredentialsType, configs []configuration.SelfServiceHook) postHooks {
	var i postHooks

	for _, h := range configs {
		switch h.Run {
		case hook.KeySessionIssuer:
			i = append(
				i,
				hook.NewSessionIssuer(m.r),
			)
		case hook.KeyRedirector:
			var rc struct {
				R string `json:"default_redirect_url"`
				A bool   `json:"allow_user_defined_redirect"`
			}

			if err := json.NewDecoder(bytes.NewBuffer(h.Config)).Decode(&rc); err != nil {
				m.l.WithError(err).
					WithField("type", credentialsType).
					WithField("hook", h.Run).
					WithField("config", fmt.Sprintf("%s", h.Config)).
					Errorf("The after hook is misconfigured.")
				continue
			}

			rcr, err := url.ParseRequestURI(rc.R)
			if err != nil {
				m.l.WithError(err).
					WithField("type", credentialsType).
					WithField("hook", h.Run).
					WithField("config", fmt.Sprintf("%s", h.Config)).
					Errorf("The after hook is misconfigured.")
				continue
			}

			i = append(
				i,
				hook.NewRedirector(
					func() *url.URL {
						return rcr
					},
					m.c.WhitelistedReturnToDomains,
					func() bool {
						return rc.A
					},
				),
			)
		default:
			m.l.
				WithField("type", credentialsType).
				WithField("hook", h.Run).
				Errorf("A unknown post login hook was requested and can therefore not be used.")
		}
	}

	return i
}

func (m *RegistryAbstract) IdentityValidator() *identity.Validator {
	if m.identityValidator == nil {
		m.identityValidator = identity.NewValidator(m.c)
	}
	return m.identityValidator
}

func (m *RegistryAbstract) WithConfig(c configuration.Provider) Registry {
	m.c = c
	return m.r
}

func (m *RegistryAbstract) Writer() herodot.Writer {
	if m.writer == nil {
		h := herodot.NewJSONWriter(m.Logger())
		m.writer = h
	}
	return m.writer
}

func (m *RegistryAbstract) Logger() logrus.FieldLogger {
	if m.l == nil {
		m.l = logrusx.New()
	}
	return m.l
}

func (m *RegistryAbstract) IdentityHandler() *identity.Handler {
	if m.identityHandler == nil {
		m.identityHandler = identity.NewHandler(m.c, m.r)
	}
	return m.identityHandler
}

func (m *RegistryAbstract) SessionHandler() *session.Handler {
	if m.sessionHandler == nil {
		m.sessionHandler = session.NewHandler(m.r, m.Writer())
	}
	return m.sessionHandler
}

func (m *RegistryAbstract) PasswordHasher() password2.Hasher {
	if m.passwordHasher == nil {
		m.passwordHasher = password2.NewHasherArgon2(m.c)
	}
	return m.passwordHasher
}

func (m *RegistryAbstract) PasswordValidator() password2.Validator {
	if m.passwordValidator == nil {
		m.passwordValidator = password2.NewDefaultPasswordValidatorStrategy()
	}
	return m.passwordValidator
}

func (m *RegistryAbstract) SelfServiceErrorHandler() *errorx.Handler {
	if m.errorHandler == nil {
		m.errorHandler = errorx.NewHandler(m.r)
	}
	return m.errorHandler
}

func (m *RegistryAbstract) CookieManager() sessions.Store {
	if m.sessionsStore == nil {
		cs := sessions.NewCookieStore(m.c.SessionSecrets()...)
		cs.Options.Secure = m.c.SelfPublicURL().Scheme == "https"
		cs.Options.HttpOnly = true
		m.sessionsStore = cs
	}
	return m.sessionsStore
}

func (m *RegistryAbstract) Tracer() *tracing.Tracer {
	if m.trc == nil {
		m.trc = &tracing.Tracer{
			ServiceName:  m.c.TracingServiceName(),
			JaegerConfig: m.c.TracingJaegerConfig(),
			Provider:     m.c.TracingProvider(),
			Logger:       m.Logger(),
		}

		if err := m.trc.Setup(); err != nil {
			m.Logger().WithError(err).Fatalf("Unable to initialize Tracer.")
		}
	}

	return m.trc
}
