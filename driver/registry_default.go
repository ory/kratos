package driver

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/ory/kratos/metrics/prometheus"

	"github.com/gobuffalo/pop/v5"

	"github.com/ory/kratos/continuity"
	"github.com/ory/kratos/hash"
	"github.com/ory/kratos/schema"
	"github.com/ory/kratos/selfservice/flow/recovery"
	"github.com/ory/kratos/selfservice/flow/settings"
	"github.com/ory/kratos/selfservice/flow/verification"
	"github.com/ory/kratos/selfservice/hook"
	"github.com/ory/kratos/selfservice/strategy/link"
	"github.com/ory/kratos/selfservice/strategy/profile"
	"github.com/ory/kratos/x"

	"github.com/cenkalti/backoff"
	"github.com/gorilla/sessions"
	"github.com/pkg/errors"

	"github.com/ory/x/dbal"
	"github.com/ory/x/healthx"
	"github.com/ory/x/sqlcon"

	"github.com/ory/x/tracing"

	"github.com/ory/x/logrusx"

	"github.com/ory/kratos/courier"
	"github.com/ory/kratos/persistence"
	"github.com/ory/kratos/persistence/sql"
	"github.com/ory/kratos/selfservice/flow/login"
	"github.com/ory/kratos/selfservice/flow/logout"
	"github.com/ory/kratos/selfservice/flow/registration"
	"github.com/ory/kratos/selfservice/strategy/oidc"

	"github.com/ory/herodot"

	"github.com/ory/kratos/driver/configuration"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/selfservice/errorx"
	password2 "github.com/ory/kratos/selfservice/strategy/password"
	"github.com/ory/kratos/session"
)

var _ Registry = new(RegistryDefault)

func init() {
	dbal.RegisterDriver(func() dbal.Driver {
		return NewRegistryDefault()
	})
}

type RegistryDefault struct {
	l *logrusx.Logger
	a *logrusx.Logger
	c configuration.Provider

	injectedSelfserviceHooks map[string]func(configuration.SelfServiceHook) interface{}

	nosurf         x.CSRFHandler
	trc            *tracing.Tracer
	pmm            *prometheus.MetricsManager
	writer         herodot.Writer
	healthxHandler *healthx.Handler
	metricsHandler *prometheus.Handler

	courier   *courier.Courier
	persister persistence.Persister

	hookVerifier         *hook.Verifier
	hookSessionIssuer    *hook.SessionIssuer
	hookSessionDestroyer *hook.SessionDestroyer

	identityHandler   *identity.Handler
	identityValidator *identity.Validator
	identityManager   *identity.Manager

	continuityManager      continuity.Manager
	continuitySessionStore *sessions.CookieStore

	schemaHandler *schema.Handler

	sessionHandler *session.Handler
	sessionsStore  *sessions.CookieStore
	sessionManager session.Manager

	passwordHasher    hash.Hasher
	passwordValidator password2.Validator

	errorHandler *errorx.Handler
	errorManager *errorx.Manager

	selfserviceRegistrationExecutor            *registration.HookExecutor
	selfserviceRegistrationHandler             *registration.Handler
	seflserviceRegistrationErrorHandler        *registration.ErrorHandler
	selfserviceRegistrationRequestErrorHandler *registration.ErrorHandler

	selfserviceLoginExecutor            *login.HookExecutor
	selfserviceLoginHandler             *login.Handler
	selfserviceLoginRequestErrorHandler *login.ErrorHandler

	selfserviceSettingsHandler      *settings.Handler
	selfserviceSettingsErrorHandler *settings.ErrorHandler
	selfserviceSettingsExecutor     *settings.HookExecutor

	selfserviceVerifyErrorHandler *verification.ErrorHandler
	selfserviceVerifyManager      *identity.Manager
	selfserviceVerifyHandler      *verification.Handler

	selfserviceLinkSender *link.Sender

	selfserviceRecoveryErrorHandler *recovery.ErrorHandler
	selfserviceRecoveryHandler      *recovery.Handler

	selfserviceLogoutHandler *logout.Handler

	selfserviceStrategies              []interface{}
	loginStrategies                    []login.Strategy
	activeCredentialsCounterStrategies []identity.ActiveCredentialsCounter
	registrationStrategies             []registration.Strategy
	profileStrategies                  []settings.Strategy
	recoveryStrategies                 []recovery.Strategy
	verificationStrategies             []verification.Strategy

	buildVersion string
	buildHash    string
	buildDate    string

	csrfTokenGenerator x.CSRFToken
}

func (m *RegistryDefault) Audit() *logrusx.Logger {
	if m.a == nil {
		m.a = logrusx.NewAudit("ORY Kratos", m.BuildVersion())
	}
	return m.a
}

func (m *RegistryDefault) RegisterPublicRoutes(router *x.RouterPublic) {
	m.LoginHandler().RegisterPublicRoutes(router)
	m.RegistrationHandler().RegisterPublicRoutes(router)
	m.LogoutHandler().RegisterPublicRoutes(router)
	m.SettingsHandler().RegisterPublicRoutes(router)
	m.LoginStrategies().RegisterPublicRoutes(router)
	m.SettingsStrategies().RegisterPublicRoutes(router)
	m.RegistrationStrategies().RegisterPublicRoutes(router)
	m.SessionHandler().RegisterPublicRoutes(router)
	m.SelfServiceErrorHandler().RegisterPublicRoutes(router)
	m.SchemaHandler().RegisterPublicRoutes(router)

	if m.c.SelfServiceFlowRecoveryEnabled() {
		m.RecoveryStrategies().RegisterPublicRoutes(router)
		m.RecoveryHandler().RegisterPublicRoutes(router)
	}

	if m.c.SelfServiceFlowVerificationEnabled() {
		m.VerificationHandler().RegisterPublicRoutes(router)
		m.VerificationStrategies().RegisterPublicRoutes(router)
	}

	m.HealthHandler().SetRoutes(router.Router, false)
}

func (m *RegistryDefault) RegisterAdminRoutes(router *x.RouterAdmin) {
	m.RegistrationHandler().RegisterAdminRoutes(router)
	m.LoginHandler().RegisterAdminRoutes(router)
	m.SchemaHandler().RegisterAdminRoutes(router)
	m.SettingsHandler().RegisterAdminRoutes(router)
	m.IdentityHandler().RegisterAdminRoutes(router)
	m.SessionHandler().RegisterAdminRoutes(router)
	m.SelfServiceErrorHandler().RegisterAdminRoutes(router)

	if m.c.SelfServiceFlowRecoveryEnabled() {
		m.RecoveryHandler().RegisterAdminRoutes(router)
		m.RecoveryStrategies().RegisterAdminRoutes(router)
	}

	if m.c.SelfServiceFlowVerificationEnabled() {
		m.VerificationHandler().RegisterAdminRoutes(router)
		m.VerificationStrategies().RegisterAdminRoutes(router)
	}

	m.HealthHandler().SetRoutes(router.Router, true)
	m.MetricsHandler().SetRoutes(router.Router)
}

func (m *RegistryDefault) RegisterRoutes(public *x.RouterPublic, admin *x.RouterAdmin) {
	m.RegisterAdminRoutes(admin)
	m.RegisterPublicRoutes(public)
}

func NewRegistryDefault() *RegistryDefault {
	return &RegistryDefault{}
}

func (m *RegistryDefault) WithBuildInfo(version, hash, date string) Registry {
	m.buildVersion = version
	m.buildHash = hash
	m.buildDate = date
	return m
}

func (m *RegistryDefault) BuildVersion() string {
	return m.buildVersion
}

func (m *RegistryDefault) BuildDate() string {
	return m.buildDate
}

func (m *RegistryDefault) BuildHash() string {
	return m.buildHash
}

func (m *RegistryDefault) WithLogger(l *logrusx.Logger) Registry {
	m.l = l
	return m
}

func (m *RegistryDefault) LogoutHandler() *logout.Handler {
	if m.selfserviceLogoutHandler == nil {
		m.selfserviceLogoutHandler = logout.NewHandler(m, m.c)
	}
	return m.selfserviceLogoutHandler
}

func (m *RegistryDefault) HealthHandler() *healthx.Handler {
	if m.healthxHandler == nil {
		m.healthxHandler = healthx.NewHandler(m.Writer(), m.BuildVersion(),
			healthx.ReadyCheckers{"database": m.Ping})
	}

	return m.healthxHandler
}

func (m *RegistryDefault) MetricsHandler() *prometheus.Handler {
	if m.metricsHandler == nil {
		m.metricsHandler = prometheus.NewHandler(m.Writer(), m.BuildVersion())
	}

	return m.metricsHandler
}

func (m *RegistryDefault) WithCSRFHandler(c x.CSRFHandler) {
	m.nosurf = c
}

func (m *RegistryDefault) CSRFHandler() x.CSRFHandler {
	if m.nosurf == nil {
		panic("csrf handler is not set")
	}
	return m.nosurf
}

func (m *RegistryDefault) selfServiceStrategies() []interface{} {
	if len(m.selfserviceStrategies) == 0 {
		m.selfserviceStrategies = []interface{}{
			password2.NewStrategy(m, m.c),
			oidc.NewStrategy(m, m.c),
			profile.NewStrategy(m, m.c),
			link.NewStrategy(m, m.c),
		}
	}

	return m.selfserviceStrategies
}

func (m *RegistryDefault) RegistrationStrategies() registration.Strategies {
	if len(m.registrationStrategies) == 0 {
		for _, strategy := range m.selfServiceStrategies() {
			if s, ok := strategy.(registration.Strategy); ok {
				if m.c.SelfServiceStrategy(string(s.ID())).Enabled {
					m.registrationStrategies = append(m.registrationStrategies, s)
				}
			}
		}
	}
	return m.registrationStrategies
}

func (m *RegistryDefault) LoginStrategies() login.Strategies {
	if len(m.loginStrategies) == 0 {
		for _, strategy := range m.selfServiceStrategies() {
			if s, ok := strategy.(login.Strategy); ok {
				if m.c.SelfServiceStrategy(string(s.ID())).Enabled {
					m.loginStrategies = append(m.loginStrategies, s)
				}
			}
		}
	}
	return m.loginStrategies
}

func (m *RegistryDefault) VerificationStrategies() verification.Strategies {
	if len(m.verificationStrategies) == 0 {
		for _, strategy := range m.selfServiceStrategies() {
			if s, ok := strategy.(verification.Strategy); ok {
				if m.c.SelfServiceStrategy(s.VerificationStrategyID()).Enabled {
					m.verificationStrategies = append(m.verificationStrategies, s)
				}
			}
		}
	}
	return m.verificationStrategies
}

func (m *RegistryDefault) ActiveCredentialsCounterStrategies() []identity.ActiveCredentialsCounter {
	if len(m.activeCredentialsCounterStrategies) == 0 {
		for _, strategy := range m.selfServiceStrategies() {
			if s, ok := strategy.(identity.ActiveCredentialsCounter); ok {
				if m.c.SelfServiceStrategy(string(s.ID())).Enabled {
					m.activeCredentialsCounterStrategies = append(m.activeCredentialsCounterStrategies, s)
				}
			}
		}
	}
	return m.activeCredentialsCounterStrategies
}

func (m *RegistryDefault) IdentityValidator() *identity.Validator {
	if m.identityValidator == nil {
		m.identityValidator = identity.NewValidator(m, m.c)
	}
	return m.identityValidator
}

func (m *RegistryDefault) WithConfig(c configuration.Provider) Registry {
	m.c = c
	return m
}

func (m *RegistryDefault) Writer() herodot.Writer {
	if m.writer == nil {
		h := herodot.NewJSONWriter(m.Logger())
		m.writer = h
	}
	return m.writer
}

func (m *RegistryDefault) Logger() *logrusx.Logger {
	if m.l == nil {
		m.l = logrusx.New("ORY Kratos", m.BuildVersion())
	}
	return m.l
}

func (m *RegistryDefault) IdentityHandler() *identity.Handler {
	if m.identityHandler == nil {
		m.identityHandler = identity.NewHandler(m.c, m)
	}
	return m.identityHandler
}

func (m *RegistryDefault) SchemaHandler() *schema.Handler {
	if m.schemaHandler == nil {
		m.schemaHandler = schema.NewHandler(m)
	}
	return m.schemaHandler
}

func (m *RegistryDefault) SessionHandler() *session.Handler {
	if m.sessionHandler == nil {
		m.sessionHandler = session.NewHandler(m)
	}
	return m.sessionHandler
}

func (m *RegistryDefault) Hasher() hash.Hasher {
	if m.passwordHasher == nil {
		m.passwordHasher = hash.NewHasherArgon2(m.c)
	}
	return m.passwordHasher
}

func (m *RegistryDefault) PasswordValidator() password2.Validator {
	if m.passwordValidator == nil {
		m.passwordValidator = password2.NewDefaultPasswordValidatorStrategy(m.c)
	}
	return m.passwordValidator
}

func (m *RegistryDefault) SelfServiceErrorHandler() *errorx.Handler {
	if m.errorHandler == nil {
		m.errorHandler = errorx.NewHandler(m)
	}
	return m.errorHandler
}

func (m *RegistryDefault) CookieManager() sessions.Store {
	if m.sessionsStore == nil {
		cs := sessions.NewCookieStore(m.c.SecretsSession()...)
		cs.Options.Secure = !m.c.IsInsecureDevMode()
		cs.Options.HttpOnly = true
		if m.c.SessionDomain() != "" {
			cs.Options.Domain = m.c.SessionDomain()
		}

		if m.c.SessionPath() != "" {
			cs.Options.Path = m.c.SessionPath()
		}

		if m.c.SessionSameSiteMode() != 0 {
			cs.Options.SameSite = m.c.SessionSameSiteMode()
		}

		cs.Options.MaxAge = 0
		if m.c.SessionPersistentCookie() {
			cs.Options.MaxAge = int(m.c.SessionLifespan().Seconds())
		}
		m.sessionsStore = cs
	}
	return m.sessionsStore
}

func (m *RegistryDefault) ContinuityCookieManager() sessions.Store {
	if m.continuitySessionStore == nil {
		cs := sessions.NewCookieStore(m.c.SecretsSession()...)
		cs.Options.Secure = !m.c.IsInsecureDevMode()
		cs.Options.HttpOnly = true
		cs.Options.SameSite = http.SameSiteLaxMode
		m.continuitySessionStore = cs
	}
	return m.continuitySessionStore
}

func (m *RegistryDefault) Tracer() *tracing.Tracer {
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

func (m *RegistryDefault) SessionManager() session.Manager {
	if m.sessionManager == nil {
		m.sessionManager = session.NewManagerHTTP(m.c, m)
	}
	return m.sessionManager
}

func (m *RegistryDefault) SelfServiceErrorManager() *errorx.Manager {
	if m.errorManager == nil {
		m.errorManager = errorx.NewManager(m, m.c)
	}
	return m.errorManager
}

func (m *RegistryDefault) CanHandle(dsn string) bool {
	return dsn == "memory" ||
		strings.HasPrefix(dsn, "mysql") ||
		strings.HasPrefix(dsn, "sqlite") ||
		strings.HasPrefix(dsn, "sqlite3") ||
		strings.HasPrefix(dsn, "postgres") ||
		strings.HasPrefix(dsn, "postgresql") ||
		strings.HasPrefix(dsn, "cockroach") ||
		strings.HasPrefix(dsn, "cockroachdb") ||
		strings.HasPrefix(dsn, "crdb")
}

func (m *RegistryDefault) Init() error {
	if m.persister != nil {
		panic("RegistryDefault.Init() must not be called more than once.")
	}

	bc := backoff.NewExponentialBackOff()
	bc.MaxElapsedTime = time.Minute * 5
	bc.Reset()
	return errors.WithStack(
		backoff.Retry(func() error {
			pool, idlePool, connMaxLifetime, cleanedDSN := sqlcon.ParseConnectionOptions(m.l, m.c.DSN())
			c, err := pop.NewConnection(&pop.ConnectionDetails{
				URL:             sqlcon.FinalizeDSN(m.l, cleanedDSN),
				IdlePool:        idlePool,
				ConnMaxLifetime: connMaxLifetime,
				Pool:            pool,
			})
			if err != nil {
				m.Logger().WithError(err).Warnf("Unable to connect to database, retrying.")
				return errors.WithStack(err)
			}
			if err := c.Open(); err != nil {
				m.Logger().WithError(err).Warnf("Unable to open database, retrying.")
				return errors.WithStack(err)
			}
			p, err := sql.NewPersister(m, m.c, c)
			if err != nil {
				m.Logger().WithError(err).Warnf("Unable to initialize persister, retrying.")
				return err
			}
			if err := p.Ping(context.Background()); err != nil {
				m.Logger().WithError(err).Warnf("Unable to ping database, retrying.")
				return err
			}
			m.persister = p
			return nil
		}, bc),
	)
}

func (m *RegistryDefault) Courier() *courier.Courier {
	if m.courier == nil {
		m.courier = courier.NewSMTP(m, m.c)
	}
	return m.courier
}

func (m *RegistryDefault) ContinuityManager() continuity.Manager {
	if m.continuityManager == nil {
		m.continuityManager = continuity.NewManagerCookie(m)
	}
	return m.continuityManager
}

func (m *RegistryDefault) ContinuityPersister() continuity.Persister {
	return m.persister
}

func (m *RegistryDefault) IdentityPool() identity.Pool {
	return m.persister
}

func (m *RegistryDefault) PrivilegedIdentityPool() identity.PrivilegedPool {
	return m.persister
}

func (m *RegistryDefault) RegistrationFlowPersister() registration.FlowPersister {
	return m.persister
}

func (m *RegistryDefault) RecoveryFlowPersister() recovery.FlowPersister {
	return m.persister
}

func (m *RegistryDefault) LoginFlowPersister() login.FlowPersister {
	return m.persister
}

func (m *RegistryDefault) SettingsFlowPersister() settings.FlowPersister {
	return m.persister
}

func (m *RegistryDefault) SelfServiceErrorPersister() errorx.Persister {
	return m.persister
}

func (m *RegistryDefault) SessionPersister() session.Persister {
	return m.persister
}

func (m *RegistryDefault) CourierPersister() courier.Persister {
	return m.persister
}

func (m *RegistryDefault) RecoveryTokenPersister() link.RecoveryTokenPersister {
	return m.Persister()
}

func (m *RegistryDefault) VerificationTokenPersister() link.VerificationTokenPersister {
	return m.Persister()
}

func (m *RegistryDefault) Persister() persistence.Persister {
	return m.persister
}

func (m *RegistryDefault) Ping() error {
	return m.persister.Ping(context.Background())
}

func (m *RegistryDefault) WithCSRFTokenGenerator(cg x.CSRFToken) {
	m.csrfTokenGenerator = cg
}

func (m *RegistryDefault) GenerateCSRFToken(r *http.Request) string {
	if m.csrfTokenGenerator == nil {
		m.csrfTokenGenerator = x.DefaultCSRFToken
	}
	return m.csrfTokenGenerator(r)
}

func (m *RegistryDefault) IdentityManager() *identity.Manager {
	if m.identityManager == nil {
		m.identityManager = identity.NewManager(m, m.c)
	}
	return m.identityManager
}

func (m *RegistryDefault) PrometheusManager() *prometheus.MetricsManager {
	if m.pmm == nil {
		m.pmm = prometheus.NewMetricsManager(m.buildVersion, m.buildHash, m.buildDate)
	}
	return m.pmm
}
