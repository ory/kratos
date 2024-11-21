// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package driver

import (
	"context"
	"crypto/sha256"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/lestrrat-go/jwx/jwk"

	"github.com/ory/kratos/selfservice/strategy/idfirst"

	"github.com/cenkalti/backoff"
	"github.com/dgraph-io/ristretto"
	"github.com/gobuffalo/pop/v6"
	"github.com/gorilla/sessions"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/luna-duclos/instrumentedsql"
	"github.com/pkg/errors"
	"go.opentelemetry.io/otel/trace/noop"

	"github.com/ory/herodot"
	"github.com/ory/kratos/cipher"
	"github.com/ory/kratos/continuity"
	"github.com/ory/kratos/courier"
	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/hash"
	"github.com/ory/kratos/hydra"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/persistence"
	"github.com/ory/kratos/persistence/sql"
	"github.com/ory/kratos/schema"
	"github.com/ory/kratos/selfservice/errorx"
	"github.com/ory/kratos/selfservice/flow/login"
	"github.com/ory/kratos/selfservice/flow/logout"
	"github.com/ory/kratos/selfservice/flow/recovery"
	"github.com/ory/kratos/selfservice/flow/registration"
	"github.com/ory/kratos/selfservice/flow/settings"
	"github.com/ory/kratos/selfservice/flow/verification"
	"github.com/ory/kratos/selfservice/hook"
	"github.com/ory/kratos/selfservice/strategy/code"
	"github.com/ory/kratos/selfservice/strategy/link"
	"github.com/ory/kratos/selfservice/strategy/lookup"
	"github.com/ory/kratos/selfservice/strategy/oidc"
	"github.com/ory/kratos/selfservice/strategy/passkey"
	"github.com/ory/kratos/selfservice/strategy/password"
	"github.com/ory/kratos/selfservice/strategy/profile"
	"github.com/ory/kratos/selfservice/strategy/totp"
	"github.com/ory/kratos/selfservice/strategy/webauthn"
	"github.com/ory/kratos/session"
	"github.com/ory/kratos/x"
	"github.com/ory/nosurf"
	"github.com/ory/x/contextx"
	"github.com/ory/x/dbal"
	"github.com/ory/x/healthx"
	"github.com/ory/x/httpx"
	"github.com/ory/x/jsonnetsecure"
	"github.com/ory/x/jwksx"
	"github.com/ory/x/logrusx"
	"github.com/ory/x/otelx"
	otelsql "github.com/ory/x/otelx/sql"
	"github.com/ory/x/popx"
	prometheus "github.com/ory/x/prometheusx"
	"github.com/ory/x/sqlcon"
)

type RegistryDefault struct {
	rwl sync.RWMutex
	l   *logrusx.Logger
	c   *config.Config

	ctxer contextx.Contextualizer

	injectedSelfserviceHooks map[string]func(config.SelfServiceHook) interface{}

	nosurf         nosurf.Handler
	trc            *otelx.Tracer
	pmm            *prometheus.MetricsManager
	writer         herodot.Writer
	healthxHandler *healthx.Handler
	metricsHandler *prometheus.Handler

	persister       persistence.Persister
	migrationStatus popx.MigrationStatuses

	hookVerifier            *hook.Verifier
	hookSessionIssuer       *hook.SessionIssuer
	hookSessionDestroyer    *hook.SessionDestroyer
	hookAddressVerifier     *hook.AddressVerifier
	hookShowVerificationUI  *hook.ShowVerificationUIHook
	hookCodeAddressVerifier *hook.CodeAddressVerifier
	hookTwoStepRegistration *hook.TwoStepRegistration

	identityHandler        *identity.Handler
	identityValidator      *identity.Validator
	identityManager        *identity.Manager
	identitySchemaProvider schema.IdentitySchemaProvider

	courierHandler *courier.Handler

	continuityManager continuity.Manager

	schemaHandler *schema.Handler

	sessionHandler   *session.Handler
	sessionManager   session.Manager
	sessionTokenizer *session.Tokenizer

	passwordHasher    hash.Hasher
	passwordValidator password.Validator

	crypter cipher.Cipher

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
	selfserviceCodeSender *code.Sender

	selfserviceRecoveryErrorHandler *recovery.ErrorHandler
	selfserviceRecoveryHandler      *recovery.Handler
	selfserviceRecoveryExecutor     *recovery.HookExecutor

	selfserviceLogoutHandler *logout.Handler

	selfserviceStrategies            []any
	replacementSelfserviceStrategies []NewStrategy

	hydra hydra.Hydra

	buildVersion string
	buildHash    string
	buildDate    string

	csrfTokenGenerator x.CSRFToken

	jsonnetVMProvider jsonnetsecure.VMProvider
	jsonnetPool       jsonnetsecure.Pool
	jwkFetcher        *jwksx.FetcherNext
}

func (m *RegistryDefault) JsonnetVM(ctx context.Context) (jsonnetsecure.VM, error) {
	if m.jsonnetVMProvider == nil {
		m.jsonnetVMProvider = &jsonnetsecure.DefaultProvider{Subcommand: "jsonnet", Pool: m.jsonnetPool}
	}
	return m.jsonnetVMProvider.JsonnetVM(ctx)
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
	m.CourierHandler().RegisterPublicRoutes(router)
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
	m.LogoutHandler().RegisterAdminRoutes(router)
	m.SchemaHandler().RegisterAdminRoutes(router)
	m.SettingsHandler().RegisterAdminRoutes(router)
	m.IdentityHandler().RegisterAdminRoutes(router)
	m.CourierHandler().RegisterAdminRoutes(router)
	m.SelfServiceErrorHandler().RegisterAdminRoutes(router)

	m.RecoveryHandler().RegisterAdminRoutes(router)
	m.AllRecoveryStrategies().RegisterAdminRoutes(router)
	m.SessionHandler().RegisterAdminRoutes(router)

	m.VerificationHandler().RegisterAdminRoutes(router)
	m.AllVerificationStrategies().RegisterAdminRoutes(router)

	m.HealthHandler(ctx).SetHealthRoutes(router, true)
	m.HealthHandler(ctx).SetVersionRoutes(router)
	m.MetricsHandler().SetRoutes(router)

	config.NewConfigHashHandler(m, router)
}

func (m *RegistryDefault) RegisterRoutes(ctx context.Context, public *x.RouterPublic, admin *x.RouterAdmin) {
	m.RegisterAdminRoutes(ctx, admin)
	m.RegisterPublicRoutes(ctx, public)
}

func NewRegistryDefault() *RegistryDefault {
	return &RegistryDefault{
		trc: otelx.NewNoop(nil, new(otelx.Config)),
	}
}

func (m *RegistryDefault) WithLogger(l *logrusx.Logger) Registry {
	m.l = l
	return m
}

func (m *RegistryDefault) WithJsonnetVMProvider(p jsonnetsecure.VMProvider) Registry {
	m.jsonnetVMProvider = p
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
				"database": func(r *http.Request) error {
					return m.PingContext(r.Context())
				},
				"migrations": func(r *http.Request) error {
					if m.migrationStatus != nil && !m.migrationStatus.HasPending() {
						return nil
					}

					status, err := m.Persister().MigrationStatus(r.Context())
					if err != nil {
						return err
					}

					if status.HasPending() {
						return errors.Errorf("migrations have not yet been fully applied")
					}

					m.migrationStatus = status
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

func (m *RegistryDefault) WithCSRFHandler(c nosurf.Handler) {
	m.nosurf = c
}

func (m *RegistryDefault) CSRFHandler() nosurf.Handler {
	if m.nosurf == nil {
		panic("csrf handler is not set")
	}
	return m.nosurf
}

func (m *RegistryDefault) Config() *config.Config {
	if m.c == nil {
		panic("configuration not set")
	}
	return m.c
}

func (m *RegistryDefault) CourierConfig() config.CourierConfigs {
	return m.Config()
}

func (m *RegistryDefault) selfServiceStrategies() []any {
	if len(m.selfserviceStrategies) == 0 {
		if m.replacementSelfserviceStrategies != nil {
			// Construct self-service strategies from the replacements
			for _, newStrategy := range m.replacementSelfserviceStrategies {
				m.selfserviceStrategies = append(m.selfserviceStrategies, newStrategy(m))
			}
		} else {
			// Construct the default list of strategies
			m.selfserviceStrategies = []any{
				password.NewStrategy(m),
				oidc.NewStrategy(m),
				profile.NewStrategy(m),
				code.NewStrategy(m),
				link.NewStrategy(m),
				totp.NewStrategy(m),
				passkey.NewStrategy(m),
				webauthn.NewStrategy(m),
				lookup.NewStrategy(m),
				idfirst.NewStrategy(m),
			}
		}
	}

	return m.selfserviceStrategies
}

func (m *RegistryDefault) strategyRegistrationEnabled(ctx context.Context, id string) bool {
	if id == "profile" {
		return m.Config().SelfServiceFlowRegistrationTwoSteps(ctx)
	}
	return m.Config().SelfServiceStrategy(ctx, id).Enabled
}

func (m *RegistryDefault) strategyLoginEnabled(ctx context.Context, id string) bool {
	return m.Config().SelfServiceStrategy(ctx, id).Enabled
}

func (m *RegistryDefault) RegistrationStrategies(ctx context.Context, filters ...registration.StrategyFilter) (registrationStrategies registration.Strategies) {
nextStrategy:
	for _, strategy := range m.selfServiceStrategies() {
		if s, ok := strategy.(registration.Strategy); ok {
			for _, filter := range filters {
				if !filter(s) {
					continue nextStrategy
				}
			}
			if m.strategyRegistrationEnabled(ctx, s.ID().String()) {
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

func (m *RegistryDefault) LoginStrategies(ctx context.Context, filters ...login.StrategyFilter) (loginStrategies login.Strategies) {
nextStrategy:
	for _, strategy := range m.selfServiceStrategies() {
		if s, ok := strategy.(login.Strategy); ok {
			for _, filter := range filters {
				if !filter(s) {
					continue nextStrategy
				}
			}

			if m.strategyLoginEnabled(ctx, s.ID().String()) {
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

func (m *RegistryDefault) ActiveCredentialsCounterStrategies(_ context.Context) (activeCredentialsCounterStrategies []identity.ActiveCredentialsCounter) {
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

// WithSelfserviceStrategies is only available in testing and overrides the
// selfservice strategies with the given ones.
func (m *RegistryDefault) WithSelfserviceStrategies(t testing.TB, strategies []any) Registry {
	if t == nil {
		panic("Passing selfservice strategies is only supported in testing")
	}
	m.selfserviceStrategies = strategies
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

func (m *RegistryDefault) CourierHandler() *courier.Handler {
	if m.courierHandler == nil {
		m.courierHandler = courier.NewHandler(m)
	}
	return m.courierHandler
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

func (m *RegistryDefault) Cipher(ctx context.Context) cipher.Cipher {
	if m.crypter == nil {
		switch m.c.CipherAlgorithm(ctx) {
		case "xchacha20-poly1305":
			m.crypter = cipher.NewCryptChaCha20(m.Config())
		case "aes":
			m.crypter = cipher.NewCryptAES(m.Config())
		default:
			m.crypter = cipher.NewNoop()
			m.l.Logger.Warning("No encryption configuration found. The default algorithm (noop) will be used, resulting in sensitive data being stored in plaintext")
		}
	}
	return m.crypter
}

func (m *RegistryDefault) Hasher(ctx context.Context) hash.Hasher {
	if m.passwordHasher == nil {
		if m.c.HasherPasswordHashingAlgorithm(ctx) == "bcrypt" {
			m.passwordHasher = hash.NewHasherBcrypt(m)
		} else {
			m.passwordHasher = hash.NewHasherArgon2(m)
		}
	}
	return m.passwordHasher
}

func (m *RegistryDefault) PasswordValidator() password.Validator {
	if m.passwordValidator == nil {
		var err error
		m.passwordValidator, err = password.NewDefaultPasswordValidatorStrategy(m)
		if err != nil {
			m.Logger().WithError(err).Fatal("could not initialize DefaultPasswordValidator")
		}
	}
	return m.passwordValidator
}

func (m *RegistryDefault) SelfServiceErrorHandler() *errorx.Handler {
	if m.errorHandler == nil {
		m.errorHandler = errorx.NewHandler(m)
	}
	return m.errorHandler
}

func (m *RegistryDefault) CookieManager(ctx context.Context) sessions.StoreExact {
	var keys [][]byte
	for _, k := range m.Config().SecretsSession(ctx) {
		encrypt := sha256.Sum256(k)
		keys = append(keys, k, encrypt[:])
	}

	cs := sessions.NewCookieStore(keys...)
	cs.Options.Secure = m.Config().SessionCookieSecure(ctx)
	cs.Options.HttpOnly = true

	if domain := m.Config().SessionDomain(ctx); domain != "" {
		cs.Options.Domain = domain
	}

	if path := m.Config().SessionPath(ctx); path != "" {
		cs.Options.Path = path
	}

	if sameSite := m.Config().SessionSameSiteMode(ctx); sameSite != 0 {
		cs.Options.SameSite = sameSite
	}

	cs.Options.MaxAge = 0
	if m.Config().SessionPersistentCookie(ctx) {
		cs.Options.MaxAge = int(m.Config().SessionLifespan(ctx).Seconds())
		cs.MaxAge(cs.Options.MaxAge)
	}
	return cs
}

func (m *RegistryDefault) ContinuityCookieManager(ctx context.Context) sessions.StoreExact {
	// To support hot reloading, this can not be instantiated only once.
	cs := sessions.NewCookieStore(m.Config().SecretsSession(ctx)...)
	cs.Options.Secure = m.Config().CookieSecure(ctx)
	cs.Options.HttpOnly = true
	cs.Options.SameSite = http.SameSiteLaxMode
	return cs
}

func (m *RegistryDefault) Tracer(ctx context.Context) *otelx.Tracer {
	if m.trc == nil {
		return otelx.NewNoop(m.l, m.Config().Tracing(ctx))
	}
	return m.trc
}

func (m *RegistryDefault) SetTracer(t *otelx.Tracer) {
	m.trc = t
}

func (m *RegistryDefault) SessionManager() session.Manager {
	if m.sessionManager == nil {
		m.sessionManager = session.NewManagerHTTP(m)
	}
	return m.sessionManager
}

func (m *RegistryDefault) Hydra() hydra.Hydra {
	if m.hydra == nil {
		m.hydra = hydra.NewDefaultHydra(m)
	}
	return m.hydra
}

func (m *RegistryDefault) WithHydra(h hydra.Hydra) Registry {
	m.hydra = h
	return m
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

func (m *RegistryDefault) Init(ctx context.Context, ctxer contextx.Contextualizer, opts ...RegistryOption) error {
	if m.persister != nil {
		// The DSN connection can not be hot-reloaded!
		panic("RegistryDefault.Init() must not be called more than once.")
	}

	o := newOptions(opts)

	m.jsonnetPool = o.jsonnetPool

	var instrumentedDriverOpts []instrumentedsql.Opt
	if m.Tracer(ctx).IsLoaded() {
		instrumentedDriverOpts = []instrumentedsql.Opt{
			instrumentedsql.WithTracer(otelsql.NewTracer()),
			instrumentedsql.WithOpsExcluded(instrumentedsql.OpSQLRowsNext),
			instrumentedsql.WithOmitArgs(), // don't risk leaking PII or secrets
		}
	}

	if o.replaceTracer != nil {
		m.trc = o.replaceTracer(m.trc)
	}

	if o.replacementStrategies != nil {
		m.replacementSelfserviceStrategies = o.replacementStrategies
	}

	if o.extraHooks != nil {
		m.WithHooks(o.extraHooks)
	}

	if o.replaceIdentitySchemaProvider != nil {
		m.identitySchemaProvider = o.replaceIdentitySchemaProvider(m)
	}

	bc := backoff.NewExponentialBackOff()
	bc.MaxElapsedTime = time.Minute * 5
	bc.Reset()
	err := backoff.Retry(func() error {
		m.WithContextualizer(ctxer)

		pool, idlePool, connMaxLifetime, connMaxIdleTime, cleanedDSN := sqlcon.ParseConnectionOptions(m.l, m.Config().DSN(ctx))
		m.Logger().
			WithField("pool", pool).
			WithField("idlePool", idlePool).
			WithField("connMaxLifetime", connMaxLifetime).
			Debug("Connecting to SQL Database")
		c, err := pop.NewConnection(&pop.ConnectionDetails{
			URL:                       sqlcon.FinalizeDSN(m.l, cleanedDSN),
			IdlePool:                  idlePool,
			ConnMaxLifetime:           connMaxLifetime,
			ConnMaxIdleTime:           connMaxIdleTime,
			Pool:                      pool,
			UseInstrumentedDriver:     m.Tracer(ctx).IsLoaded(),
			InstrumentedDriverOptions: instrumentedDriverOpts,
		})
		if err != nil {
			m.Logger().WithError(err).Warnf("Unable to connect to database, retrying.")
			return errors.WithStack(err)
		}
		if err := c.Open(); err != nil {
			m.Logger().WithError(err).Warnf("Unable to open database, retrying.")
			return errors.WithStack(err)
		}
		p, err := sql.NewPersister(ctx, m, c,
			sql.WithExtraMigrations(o.extraMigrations...),
			sql.WithExtraGoMigrations(o.extraGoMigrations...),
			sql.WithDisabledLogging(o.disableMigrationLogging))
		if err != nil {
			m.Logger().WithError(err).Warnf("Unable to initialize persister, retrying.")
			return err
		}

		ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		if err := c.Store.SQLDB().PingContext(ctx); err != nil {
			m.Logger().WithError(err).Warnf("Unable to ping database, retrying.")
			return err
		}

		// if dsn is memory we have to run the migrations on every start
		if dbal.IsMemorySQLite(m.Config().DSN(ctx)) || m.Config().DSN(ctx) == "memory" {
			m.Logger().Infoln("Ory Kratos is running migrations on every startup as DSN is memory. This means your data is lost when Kratos terminates.")
			if err := p.MigrateUp(ctx); err != nil {
				m.Logger().WithError(err).Warnf("Unable to run migrations, retrying.")
				return err
			}
		}

		if o.skipNetworkInit {
			m.persister = p
			return nil
		}

		net, err := p.DetermineNetwork(ctx)
		if err != nil {
			m.Logger().WithError(err).Warnf("Unable to determine network, retrying.")
			return err
		}

		m.persister = p.WithNetworkID(net.ID)
		return nil
	}, bc)
	if err != nil {
		return err
	}

	if o.inspect != nil {
		if err := o.inspect(m); err != nil {
			return errors.WithStack(err)
		}
	}
	return nil
}

func (m *RegistryDefault) SetPersister(p persistence.Persister) {
	m.persister = p
}

func (m *RegistryDefault) Courier(ctx context.Context) (courier.Courier, error) {
	return courier.NewCourier(ctx, m)
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

func (m *RegistryDefault) RecoveryCodePersister() code.RecoveryCodePersister {
	return m.Persister()
}

func (m *RegistryDefault) VerificationTokenPersister() link.VerificationTokenPersister {
	return m.Persister()
}

func (m *RegistryDefault) VerificationCodePersister() code.VerificationCodePersister {
	return m.Persister()
}

func (m *RegistryDefault) RegistrationCodePersister() code.RegistrationCodePersister {
	return m.Persister()
}

func (m *RegistryDefault) LoginCodePersister() code.LoginCodePersister {
	return m.Persister()
}

func (m *RegistryDefault) TransactionalPersisterProvider() x.TransactionalPersister {
	return m.Persister()
}

func (m *RegistryDefault) Persister() persistence.Persister {
	return m.persister
}

func (m *RegistryDefault) PingContext(ctx context.Context) error {
	return m.persister.Ping(ctx)
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
		m.identityManager = identity.NewManager(m)
	}
	return m.identityManager
}

func (m *RegistryDefault) PrometheusManager() *prometheus.MetricsManager {
	m.rwl.Lock()
	defer m.rwl.Unlock()
	if m.pmm == nil {
		m.pmm = prometheus.NewMetricsManagerWithPrefix("kratos", prometheus.HTTPMetrics, m.buildVersion, m.buildHash, m.buildDate)
	}
	return m.pmm
}

func (m *RegistryDefault) HTTPClient(_ context.Context, opts ...httpx.ResilientOptions) *retryablehttp.Client {
	opts = append(opts,
		httpx.ResilientClientWithLogger(m.Logger()),
		httpx.ResilientClientWithMaxRetry(2),
		httpx.ResilientClientWithConnectionTimeout(30*time.Second),
		httpx.ResilientClientWithTracer(noop.NewTracerProvider().Tracer("Ory Kratos")), // will use the tracer from a context if available
	)

	// One of the few exceptions, this usually should not be hot reloaded.
	if m.Config().ClientHTTPNoPrivateIPRanges(contextx.RootContext) {
		opts = append(
			opts,
			httpx.ResilientClientDisallowInternalIPs(),
			// One of the few exceptions, this usually should not be hot reloaded.
			httpx.ResilientClientAllowInternalIPRequestsTo(m.Config().ClientHTTPPrivateIPExceptionURLs(contextx.RootContext)...),
		)
	}
	return httpx.NewResilientClient(opts...)
}

func (m *RegistryDefault) WithContextualizer(ctxer contextx.Contextualizer) Registry {
	m.ctxer = ctxer
	return m
}

func (m *RegistryDefault) Contextualizer() contextx.Contextualizer {
	if m.ctxer == nil {
		panic("registry Contextualizer not set")
	}
	return m.ctxer
}

func (m *RegistryDefault) JWKSFetcher() *jwksx.FetcherNext {
	if m.jwkFetcher == nil {
		maxItems := int64(10000000)
		cache, _ := ristretto.NewCache(&ristretto.Config[[]byte, jwk.Set]{
			NumCounters:        maxItems * 10,
			MaxCost:            maxItems,
			BufferItems:        64,
			Metrics:            true,
			IgnoreInternalCost: true,
			Cost: func(value jwk.Set) int64 {
				return 1
			},
		})

		m.jwkFetcher = jwksx.NewFetcherNext(cache)
	}
	return m.jwkFetcher
}

func (m *RegistryDefault) SessionTokenizer() *session.Tokenizer {
	if m.sessionTokenizer == nil {
		m.sessionTokenizer = session.NewTokenizer(m)
	}
	return m.sessionTokenizer
}
