package driver

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/gorilla/sessions"
	"github.com/justinas/nosurf"
	"github.com/sirupsen/logrus"

	"github.com/ory/x/logrusx"

	"github.com/ory/x/dbal"

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

var _ Registry = new(RegistryMemory)

func init() {
	dbal.RegisterDriver(new(RegistryMemory))
}

type RegistryMemory struct {
	l             logrus.FieldLogger
	ip            identity.Pool
	ih            *identity.Handler
	sm            session.Manager
	sh            *session.Handler
	em            errorx.Manager
	eh            *errorx.Handler
	c             configuration.Provider
	ph            password2.Hasher
	pv            password2.Validator
	cookieManager sessions.Store

	selfserviceRequestManager       selfservice.RequestManager
	identityValidator               *identity.Validator
	selfserviceRegistrationExecutor *selfservice.RegistrationExecutor
	selfserviceLoginExecutor        *selfservice.LoginExecutor
	selfserviceStrategyHandler      *selfservice.StrategyHandler
	selfserviceStrategies           []selfservice.Strategy
	seflserviceRequestErrorHandler  *selfservice.ErrorHandler

	nosurf *nosurf.CSRFHandler

	writer herodot.Writer
}

func (m *RegistryMemory) WithLogger(l logrus.FieldLogger) Registry {
	m.l = l
	return m
}

func (m *RegistryMemory) WithCSRFHandler(c *nosurf.CSRFHandler) {
	m.nosurf = c
}

func (m *RegistryMemory) CSRFHandler() *nosurf.CSRFHandler {
	if m.nosurf == nil {
		panic("csrf handler is not set")
	}
	return m.nosurf
}

func (m *RegistryMemory) SelfServiceStrategies() []selfservice.Strategy {
	if m.selfserviceStrategies == nil {
		m.selfserviceStrategies = []selfservice.Strategy{
			password2.NewStrategy(m, m.c),
			oidc.NewStrategy(m, m.c),
		}
	}
	return m.selfserviceStrategies
}

type postHooks []interface {
	selfservice.HookLoginPostExecutor
	selfservice.HookRegistrationPostExecutor
}

func (m *RegistryMemory) hooksPost(credentialsType identity.CredentialsType, configs []configuration.SelfServiceHook) postHooks {
	var i postHooks

	for _, hook := range configs {
		switch hook.Run {
		case hooks.KeySessionIssuer:
			i = append(
				i,
				hooks.NewSessionIssuer(m),
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

func (m *RegistryMemory) PostRegistrationHooks(credentialsType identity.CredentialsType) []selfservice.HookRegistrationPostExecutor {
	a := m.hooksPost(credentialsType, m.c.SelfServiceRegistrationAfterHooks(string(credentialsType)))
	b := make([]selfservice.HookRegistrationPostExecutor, len(a))
	for k, v := range a {
		b[k] = v
	}
	return b
}

func (m *RegistryMemory) PostLoginHooks(credentialsType identity.CredentialsType) []selfservice.HookLoginPostExecutor {
	a := m.hooksPost(credentialsType, m.c.SelfServiceLoginAfterHooks(string(credentialsType)))
	b := make([]selfservice.HookLoginPostExecutor, len(a))
	for k, v := range a {
		b[k] = v
	}
	return b
}

func (m *RegistryMemory) SelfServiceRequestErrorHandler() *selfservice.ErrorHandler {
	if m.seflserviceRequestErrorHandler == nil {
		m.seflserviceRequestErrorHandler = selfservice.NewErrorHandler(m, m.c)
	}
	return m.seflserviceRequestErrorHandler
}

func (m *RegistryMemory) AuthHookRegistrationPreExecutors() []selfservice.HookRegistrationPreExecutor {
	return []selfservice.HookRegistrationPreExecutor{}
}

func (m *RegistryMemory) AuthHookLoginPreExecutors() []selfservice.HookLoginPreExecutor {
	return []selfservice.HookLoginPreExecutor{}
}

func (m *RegistryMemory) RegistrationRequestManager() selfservice.RegistrationRequestManager {
	if m.selfserviceRequestManager == nil {
		m.selfserviceRequestManager = selfservice.NewRequestManagerMemory()
	}
	return m.selfserviceRequestManager
}

func (m *RegistryMemory) IdentityValidator() *identity.Validator {
	if m.identityValidator == nil {
		m.identityValidator = identity.NewValidator(m.c)
	}
	return m.identityValidator
}

func (m *RegistryMemory) RegistrationExecutor() *selfservice.RegistrationExecutor {
	if m.selfserviceRegistrationExecutor == nil {
		m.selfserviceRegistrationExecutor = selfservice.NewRegistrationExecutor(m, m.c)
	}
	return m.selfserviceRegistrationExecutor
}

func (m *RegistryMemory) LoginRequestManager() selfservice.LoginRequestManager {
	if m.selfserviceRequestManager == nil {
		m.selfserviceRequestManager = selfservice.NewRequestManagerMemory()
	}
	return m.selfserviceRequestManager
}

func (m *RegistryMemory) LoginExecutor() *selfservice.LoginExecutor {
	if m.selfserviceLoginExecutor == nil {
		m.selfserviceLoginExecutor = selfservice.NewLoginExecutor(m, m.c)
	}
	return m.selfserviceLoginExecutor
}

func (m *RegistryMemory) WithConfig(c configuration.Provider) Registry {
	m.c = c
	return m
}

func (m *RegistryMemory) Writer() herodot.Writer {
	if m.writer == nil {
		h := herodot.NewJSONWriter(m.Logger())
		m.writer = h
	}
	return m.writer
}

func (m *RegistryMemory) Logger() logrus.FieldLogger {
	if m.l == nil {
		m.l = logrusx.New()
	}
	return m.l
}

func (m *RegistryMemory) IdentityPool() identity.Pool {
	if m.ip == nil {
		m.ip = identity.NewPoolMemory(m)
	}
	return m.ip
}

func (m *RegistryMemory) IdentityHandler() *identity.Handler {
	if m.ih == nil {
		m.ih = identity.NewHandler(m.c, m)
	}
	return m.ih
}

func (m *RegistryMemory) SessionManager() session.Manager {
	if m.sm == nil {
		m.sm = session.NewManagerMemory(m.c, m)
	}
	return m.sm
}

func (m *RegistryMemory) SessionHandler() *session.Handler {
	if m.sh == nil {
		m.sh = session.NewHandler(m, m.Writer())
	}
	return m.sh
}

func (m *RegistryMemory) CanHandle(dsn string) bool {
	return dsn == "memory"
}

func (m *RegistryMemory) Ping() error {
	return nil
}

func (m *RegistryMemory) ErrorManager() errorx.Manager {
	if m.em == nil {
		m.em = errorx.NewMemoryManager(m.Logger(), m.writer, m.c)
	}
	return m.em
}

func (m *RegistryMemory) PasswordHasher() password2.Hasher {
	if m.ph == nil {
		m.ph = password2.NewHasherArgon2(m.c)
	}
	return m.ph
}

func (m *RegistryMemory) PasswordValidator() password2.Validator {
	if m.pv == nil {
		m.pv = password2.NewDefaultPasswordValidatorStrategy()
	}
	return m.pv
}

func (m *RegistryMemory) ErrorHandler() *errorx.Handler {
	if m.eh == nil {
		m.eh = errorx.NewHandler(m)
	}
	return m.eh
}

func (m *RegistryMemory) StrategyHandler() *selfservice.StrategyHandler {
	if m.selfserviceStrategyHandler == nil {
		m.selfserviceStrategyHandler = selfservice.NewStrategyHandler(m, m.c)
	}

	return m.selfserviceStrategyHandler
}

func (m *RegistryMemory) CookieManager() sessions.Store {
	if m.cookieManager == nil {
		cs := sessions.NewCookieStore(m.c.SessionSecrets()...)
		cs.Options.Secure = m.c.SelfPublicURL().Scheme == "https"
		cs.Options.HttpOnly = true
		m.cookieManager = cs
	}
	return m.cookieManager
}
