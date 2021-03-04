package driver

import (
	"context"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/luna-duclos/instrumentedsql"
	"github.com/luna-duclos/instrumentedsql/opentracing"

	"github.com/ory/kratos/corp"

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

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/selfservice/errorx"
	password2 "github.com/ory/kratos/selfservice/strategy/password"
	"github.com/ory/kratos/session"
)

type RegistryDefault struct {
	rwl sync.RWMutex
	l   *logrusx.Logger
	c   *config.Config

	injectedSelfserviceHooks map[string]func(config.SelfServiceHook) interface{}

	nosurf         x.CSRFHandler
	trc            *tracing.Tracer
	pmm            *prometheus.MetricsManager
	writer         herodot.Writer
	healthxHandler *healthx.Handler
	metricsHandler *prometheus.Handler

	persister persistence.Persister

	hookVerifier         *hook.Verifier
	hookSessionIssuer    *hook.SessionIssuer
	hookSessionDestroyer *hook.SessionDestroyer

	identityHandler   *identity.Handler
	identityValidator *identity.Validator
	identityManager   *identity.Manager

	continuityManager continuity.Manager

	schemaHandler *schema.Handler

	sessionHandler *session.Handler
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

	selfserviceVerifyErrorHandler   *verification.ErrorHandler
	selfserviceVerifyManager        *identity.Manager
	selfserviceVerifyHandler        *verification.Handler
	selfserviceVerificationExecutor *verification.HookExecutor

	selfserviceLinkSender *link.Sender

	selfserviceRecoveryErrorHandler *recovery.ErrorHandler
	selfserviceRecoveryHandler      *recovery.Handler
	selfserviceRecoveryExecutor     *recovery.HookExecutor

	selfserviceLogoutHandler *logout.Handler

	selfserviceStrategies []interface{}

	buildVersion string
	buildHash    string
	buildDate    string

	csrfTokenGenerator x.CSRFToken
}

func (m *RegistryDefault) Audit() *logrusx.Logger {
	return m.Logger().WithField("audience", "audit")
}

func (m *RegistryDefault) RegisterPublicRoutes(ctx context.Context, router *x.RouterPublic) {
	m.LoginHandler().RegisterPublicRoutes(router)
	m.RegistrationHandler().RegisterPublicRoutes(router)
	m.LogoutHandler().RegisterPublicRoutes(router)
	m.SettingsHandler().RegisterPublicRoutes(router)
	m.IdentityHandler().RegisterPublicRoutes(router)
	m.AllLoginStrategies().RegisterPublicRoutes(router)
	m.AllSettingsStrategies().RegisterPublicRoutes(router)
	m.AllRegistrationStrategies().RegisterPublicRoutes(router)
	m.SessionHandler().RegisterPublicRoutes(router)
	m.SelfServiceErrorHandler().RegisterPublicRoutes(router)
	m.SchemaHandler().RegisterPublicRoutes(router)

	m.AllRecoveryStrategies().RegisterPublicRoutes(router)
	m.RecoveryHandler().RegisterPublicRoutes(router)

	m.VerificationHandler().RegisterPublicRoutes(router)
	m.AllVerificationStrategies().RegisterPublicRoutes(router)

	m.HealthHandler(ctx).SetHealthRoutes(router.Router, false)
}

func (m *RegistryDefault) RegisterAdminRoutes(ctx context.Context, router *x.RouterAdmin) {
	m.RegistrationHandler().RegisterAdminRoutes(router)
	m.LoginHandler().RegisterAdminRoutes(router)
	m.SchemaHandler().RegisterAdminRoutes(router)
	m.SettingsHandler().RegisterAdminRoutes(router)
	m.IdentityHandler().RegisterAdminRoutes(router)
	m.SessionHandler().RegisterAdminRoutes(router)
	m.SelfServiceErrorHandler().RegisterAdminRoutes(router)

	m.RecoveryHandler().RegisterAdminRoutes(router)
	m.AllRecoveryStrategies().RegisterAdminRoutes(router)

	m.VerificationHandler().RegisterAdminRoutes(router)
	m.AllVerificationStrategies().RegisterAdminRoutes(router)

	m.HealthHandler(ctx).SetHealthRoutes(router.Router, true)
	m.HealthHandler(ctx).SetVersionRoutes(router.Router)
	m.MetricsHandler().SetRoutes(router.Router)
}

func (m *RegistryDefault) RegisterRoutes(ctx context.Context, public *x.RouterPublic, admin *x.RouterAdmin) {
	m.RegisterAdminRoutes(ctx, admin)
	m.RegisterPublicRoutes(ctx, public)
}

func NewRegistryDefault() *RegistryDefault {
	return &RegistryDefault{}
}

func (m *RegistryDefault) WithLogger(l *logrusx.Logger) Registry {
	m.l = l
	return m
}

func (m *RegistryDefault) LogoutHandler() *logout.Handler {
	if m.selfserviceLogoutHandler == nil {
		m.selfserviceLogoutHandler = logout.NewHandler(m)
	}
	return m.selfserviceLogoutHandler
}

func (m *RegistryDefault) HealthHandler(_ context.Context) *healthx.Handler {
	if m.healthxHandler == nil {
		m.healthxHandler = healthx.NewHandler(m.Writer(), config.Version,
			healthx.ReadyCheckers{
				"database": func(_ *http.Request) error {
					return m.Ping()
				},
				"migrations": func(r *http.Request) error {
					status, err := m.Persister().MigrationStatus(r.Context())
					if err != nil {
						return err
					}

					if status.HasPending() {
						return errors.Errorf("migrations have not yet been fully applied")
					}

					return nil
				},
			})
	}

	return m.healthxHandler
}

func (m *RegistryDefault) MetricsHandler() *prometheus.Handler {
	if m.metricsHandler == nil {
		m.metricsHandler = prometheus.NewHandler(m.Writer(), config.Version)
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

func (m *RegistryDefault) Config(ctx context.Context) *config.Config {
	if m.c == nil {
		panic("configuration not set")
	}
	return corp.ContextualizeConfig(ctx, m.c)
}

func (m *RegistryDefault) selfServiceStrategies() []interface{} {
	if len(m.selfserviceStrategies) == 0 {
		m.selfserviceStrategies = []interface{}{
			password2.NewStrategy(m),
			oidc.NewStrategy(m),
			profile.NewStrategy(m),
			link.NewStrategy(m),
		}
	}

	return m.selfserviceStrategies
}

func (m *RegistryDefault) RegistrationStrategies(ctx context.Context) (registrationStrategies registration.Strategies) {
	for _, strategy := range m.selfServiceStrategies() {
		if s, ok := strategy.(registration.Strategy); ok {
			if m.Config(ctx).SelfServiceStrategy(string(s.ID())).Enabled {
				registrationStrategies = append(registrationStrategies, s)
			}
		}
	}
	return
}

func (m *RegistryDefault) AllRegistrationStrategies() registration.Strategies {
	var registrationStrategies []registration.Strategy

	for _, strategy := range m.selfServiceStrategies() {
		if s, ok := strategy.(registration.Strategy); ok {
			registrationStrategies = append(registrationStrategies, s)
		}
	}
	return registrationStrategies
}

func (m *RegistryDefault) LoginStrategies(ctx context.Context) (loginStrategies login.Strategies) {
	for _, strategy := range m.selfServiceStrategies() {
		if s, ok := strategy.(login.Strategy); ok {
			if m.Config(ctx).SelfServiceStrategy(string(s.ID())).Enabled {
				loginStrategies = append(loginStrategies, s)
			}
		}
	}
	return
}

func (m *RegistryDefault) AllLoginStrategies() login.Strategies {
	var loginStrategies []login.Strategy
	for _, strategy := range m.selfServiceStrategies() {
		if s, ok := strategy.(login.Strategy); ok {
			loginStrategies = append(loginStrategies, s)
		}
	}
	return loginStrategies
}

func (m *RegistryDefault) ActiveCredentialsCounterStrategies(ctx context.Context) (activeCredentialsCounterStrategies []identity.ActiveCredentialsCounter) {
	for _, strategy := range m.selfServiceStrategies() {
		if s, ok := strategy.(identity.ActiveCredentialsCounter); ok {
			activeCredentialsCounterStrategies = append(activeCredentialsCounterStrategies, s)
		}
	}
	return
}

func (m *RegistryDefault) IdentityValidator() *identity.Validator {
	if m.identityValidator == nil {
		m.identityValidator = identity.NewValidator(m)
	}
	return m.identityValidator
}

func (m *RegistryDefault) WithConfig(c *config.Config) Registry {
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
		m.l = logrusx.New("Ory Kratos", config.Version)
	}
	return m.l
}

func (m *RegistryDefault) IdentityHandler() *identity.Handler {
	if m.identityHandler == nil {
		m.identityHandler = identity.NewHandler(m)
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
		if m.c.HasherPasswordHashingAlgorithm() == "bcrypt" {
			m.passwordHasher = hash.NewHasherBcrypt(m)
		} else {
			m.passwordHasher = hash.NewHasherArgon2(m)
		}
	}
	return m.passwordHasher
}

func (m *RegistryDefault) PasswordValidator() password2.Validator {
	if m.passwordValidator == nil {
		m.passwordValidator = password2.NewDefaultPasswordValidatorStrategy(m)
	}
	return m.passwordValidator
}

func (m *RegistryDefault) SelfServiceErrorHandler() *errorx.Handler {
	if m.errorHandler == nil {
		m.errorHandler = errorx.NewHandler(m)
	}
	return m.errorHandler
}

func (m *RegistryDefault) CookieManager(ctx context.Context) sessions.Store {
	cs := sessions.NewCookieStore(m.Config(ctx).SecretsSession()...)
	cs.Options.Secure = !m.Config(ctx).IsInsecureDevMode()
	cs.Options.HttpOnly = true

	if domain := m.Config(ctx).SessionDomain(); domain != "" {
		cs.Options.Domain = domain
	}

	if path := m.Config(ctx).SessionPath(); path != "" {
		cs.Options.Path = path
	}

	if sameSite := m.Config(ctx).SessionSameSiteMode(); sameSite != 0 {
		cs.Options.SameSite = sameSite
	}

	cs.Options.MaxAge = 0
	if m.Config(ctx).SessionPersistentCookie() {
		cs.Options.MaxAge = int(m.Config(ctx).SessionLifespan().Seconds())
	}
	return cs
}

func (m *RegistryDefault) ContinuityCookieManager(ctx context.Context) sessions.Store {
	// To support hot reloading, this can not be instantiated only once.
	cs := sessions.NewCookieStore(m.Config(ctx).SecretsSession()...)
	cs.Options.Secure = !m.Config(ctx).IsInsecureDevMode()
	cs.Options.HttpOnly = true
	cs.Options.SameSite = http.SameSiteLaxMode
	return cs
}

func (m *RegistryDefault) Tracer(ctx context.Context) *tracing.Tracer {
	if m.trc == nil {
		// Tracing is initialized only once so it can not be hot reloaded or context-aware.
		t, err := tracing.New(m.l, m.Config(ctx).Tracing())
		if err != nil {
			m.Logger().WithError(err).Fatalf("Unable to initialize Tracer.")
		}
		m.trc = t
	}

	return m.trc
}

func (m *RegistryDefault) SessionManager() session.Manager {
	if m.sessionManager == nil {
		m.sessionManager = session.NewManagerHTTP(m)
	}
	return m.sessionManager
}

func (m *RegistryDefault) SelfServiceErrorManager() *errorx.Manager {
	if m.errorManager == nil {
		m.errorManager = errorx.NewManager(m)
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

func (m *RegistryDefault) Init(ctx context.Context) error {
	if m.persister != nil {
		// The DSN connection can not be hot-reloaded!
		panic("RegistryDefault.Init() must not be called more than once.")
	}

	bc := backoff.NewExponentialBackOff()
	bc.MaxElapsedTime = time.Minute * 5
	bc.Reset()
	return errors.WithStack(
		backoff.Retry(func() error {
			var opts []instrumentedsql.Opt
			if m.Tracer(ctx).IsLoaded() {
				opts = []instrumentedsql.Opt{
					instrumentedsql.WithTracer(opentracing.NewTracer(true)),
					instrumentedsql.WithOmitArgs(),
				}
			}

			pool, idlePool, connMaxLifetime, cleanedDSN := sqlcon.ParseConnectionOptions(m.l, m.Config(ctx).DSN())
			m.Logger().
				WithField("pool", pool).
				WithField("idlePool", idlePool).
				WithField("connMaxLifetime", connMaxLifetime).
				Debug("Connecting to SQL Database")
			c, err := pop.NewConnection(&pop.ConnectionDetails{
				URL:                       sqlcon.FinalizeDSN(m.l, cleanedDSN),
				IdlePool:                  idlePool,
				ConnMaxLifetime:           connMaxLifetime,
				Pool:                      pool,
				UseInstrumentedDriver:     m.Tracer(ctx).IsLoaded(),
				InstrumentedDriverOptions: opts,
			})
			if err != nil {
				m.Logger().WithError(err).Warnf("Unable to connect to database, retrying.")
				return errors.WithStack(err)
			}
			if err := c.Open(); err != nil {
				m.Logger().WithError(err).Warnf("Unable to open database, retrying.")
				return errors.WithStack(err)
			}
			p, err := sql.NewPersister(ctx, m, c)
			if err != nil {
				m.Logger().WithError(err).Warnf("Unable to initialize persister, retrying.")
				return err
			}

			if err := p.Ping(); err != nil {
				m.Logger().WithError(err).Warnf("Unable to ping database, retrying.")
				return err
			}

			net, err := p.DetermineNetwork(ctx)
			if err != nil {
				m.Logger().WithError(err).Warnf("Unable to determine network, retrying.")
				return err
			}

			// if dsn is memory we have to run the migrations on every start
			if dbal.IsMemorySQLite(m.Config(ctx).DSN()) || m.Config(ctx).DSN() == dbal.SQLiteInMemory || m.Config(ctx).DSN() == dbal.SQLiteSharedInMemory || m.Config(ctx).DSN() == "memory" {
				m.Logger().Infoln("Ory Kratos is running migrations on every startup as DSN is memory. This means your data is lost when Kratos terminates.")
				if err := p.MigrateUp(ctx); err != nil {
					m.Logger().WithError(err).Warnf("Unable to run migrations, retrying.")
					return err
				}
			}

			m.persister = p.WithNetworkID(net.ID)
			return nil
		}, bc),
	)
}

func (m *RegistryDefault) Courier(ctx context.Context) *courier.Courier {
	return courier.NewSMTP(m, m.Config(ctx))
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
	return m.persister.Ping()
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
		m.identityManager = identity.NewManager(m)
	}
	return m.identityManager
}

func (m *RegistryDefault) PrometheusManager() *prometheus.MetricsManager {
	m.rwl.Lock()
	defer m.rwl.Unlock()
	if m.pmm == nil {
		m.pmm = prometheus.NewMetricsManager(m.buildVersion, m.buildHash, m.buildDate)
	}
	return m.pmm
}
