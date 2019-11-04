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

	"github.com/ory/hive/selfservice"
	"github.com/ory/hive/selfservice/hooks"
	"github.com/ory/hive/selfservice/oidc"

	"github.com/ory/herodot"

	"github.com/ory/hive/driver/configuration"
	"github.com/ory/hive/errorx"
	"github.com/ory/hive/identity"
	password2 "github.com/ory/hive/selfservice/password"
	"github.com/ory/hive/session"
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

	selfserviceRegistrationExecutor *selfservice.RegistrationExecutor
	selfserviceLoginExecutor        *selfservice.LoginExecutor
	selfserviceStrategyHandler      *selfservice.StrategyHandler
	selfserviceStrategies           []selfservice.Strategy
	seflserviceRequestErrorHandler  *selfservice.ErrorHandler

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

func (m *RegistryAbstract) SelfServiceStrategies() []selfservice.Strategy {
	if m.selfserviceStrategies == nil {
		m.selfserviceStrategies = []selfservice.Strategy{
			password2.NewStrategy(m.r, m.c),
			oidc.NewStrategy(m.r, m.c),
		}
	}
	return m.selfserviceStrategies
}

type postHooks []interface {
	selfservice.HookLoginPostExecutor
	selfservice.HookRegistrationPostExecutor
}

func (m *RegistryAbstract) hooksPost(credentialsType identity.CredentialsType, configs []configuration.SelfServiceHook) postHooks {
	var i postHooks

	for _, hook := range configs {
		switch hook.Run {
		case hooks.KeySessionIssuer:
			i = append(
				i,
				hooks.NewSessionIssuer(m.r),
			)
		case hooks.KeyRedirector:
			var rc struct {
				R string `json:"default_redirect_url"`
				A bool   `json:"allow_user_defined_redirect"`
			}

			if err := json.NewDecoder(bytes.NewBuffer(hook.Config)).Decode(&rc); err != nil {
				m.l.WithError(err).
					WithField("type", credentialsType).
					WithField("hook", hook.Run).
					WithField("config", fmt.Sprintf("%s", hook.Config)).
					Errorf("The after hook is misconfigured.")
				continue
			}

			rcr, err := url.ParseRequestURI(rc.R)
			if err != nil {
				m.l.WithError(err).
					WithField("type", credentialsType).
					WithField("hook", hook.Run).
					WithField("config", fmt.Sprintf("%s", hook.Config)).
					Errorf("The after hook is misconfigured.")
				continue
			}

			i = append(
				i,
				hooks.NewRedirector(
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
				WithField("hook", hook.Run).
				Errorf("A unknown post login hook was requested and can therefore not be used.")
		}
	}

	return i
}

func (m *RegistryAbstract) PostRegistrationHooks(credentialsType identity.CredentialsType) []selfservice.HookRegistrationPostExecutor {
	a := m.hooksPost(credentialsType, m.c.SelfServiceRegistrationAfterHooks(string(credentialsType)))
	b := make([]selfservice.HookRegistrationPostExecutor, len(a))
	for k, v := range a {
		b[k] = v
	}
	return b
}

func (m *RegistryAbstract) PostLoginHooks(credentialsType identity.CredentialsType) []selfservice.HookLoginPostExecutor {
	a := m.hooksPost(credentialsType, m.c.SelfServiceLoginAfterHooks(string(credentialsType)))
	b := make([]selfservice.HookLoginPostExecutor, len(a))
	for k, v := range a {
		b[k] = v
	}
	return b
}

func (m *RegistryAbstract) SelfServiceRequestErrorHandler() *selfservice.ErrorHandler {
	if m.seflserviceRequestErrorHandler == nil {
		m.seflserviceRequestErrorHandler = selfservice.NewErrorHandler(m.r, m.c)
	}
	return m.seflserviceRequestErrorHandler
}

func (m *RegistryAbstract) AuthHookRegistrationPreExecutors() []selfservice.HookRegistrationPreExecutor {
	return []selfservice.HookRegistrationPreExecutor{}
}

func (m *RegistryAbstract) AuthHookLoginPreExecutors() []selfservice.HookLoginPreExecutor {
	return []selfservice.HookLoginPreExecutor{}
}

func (m *RegistryAbstract) IdentityValidator() *identity.Validator {
	if m.identityValidator == nil {
		m.identityValidator = identity.NewValidator(m.c)
	}
	return m.identityValidator
}

func (m *RegistryAbstract) RegistrationExecutor() *selfservice.RegistrationExecutor {
	if m.selfserviceRegistrationExecutor == nil {
		m.selfserviceRegistrationExecutor = selfservice.NewRegistrationExecutor(m.r, m.c)
	}
	return m.selfserviceRegistrationExecutor
}

func (m *RegistryAbstract) LoginExecutor() *selfservice.LoginExecutor {
	if m.selfserviceLoginExecutor == nil {
		m.selfserviceLoginExecutor = selfservice.NewLoginExecutor(m.r, m.c)
	}
	return m.selfserviceLoginExecutor
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

func (m *RegistryAbstract) ErrorHandler() *errorx.Handler {
	if m.errorHandler == nil {
		m.errorHandler = errorx.NewHandler(m.r)
	}
	return m.errorHandler
}

func (m *RegistryAbstract) StrategyHandler() *selfservice.StrategyHandler {
	if m.selfserviceStrategyHandler == nil {
		m.selfserviceStrategyHandler = selfservice.NewStrategyHandler(m.r, m.c)
	}

	return m.selfserviceStrategyHandler
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
