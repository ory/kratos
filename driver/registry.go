// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package driver

import (
	"context"
	"io/fs"

	"github.com/ory/pop/v6"
	"github.com/ory/x/configx"
	"github.com/ory/x/servicelocatorx"

	"github.com/gorilla/sessions"

	"github.com/ory/kratos/cipher"
	"github.com/ory/kratos/continuity"
	"github.com/ory/kratos/courier"
	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/hash"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/persistence"
	"github.com/ory/kratos/schema"
	"github.com/ory/kratos/selfservice/errorx"
	"github.com/ory/kratos/selfservice/flow/login"
	"github.com/ory/kratos/selfservice/flow/logout"
	"github.com/ory/kratos/selfservice/flow/recovery"
	"github.com/ory/kratos/selfservice/flow/registration"
	"github.com/ory/kratos/selfservice/flow/settings"
	"github.com/ory/kratos/selfservice/flow/verification"
	"github.com/ory/kratos/selfservice/sessiontokenexchange"
	"github.com/ory/kratos/selfservice/strategy/code"
	"github.com/ory/kratos/selfservice/strategy/link"
	password2 "github.com/ory/kratos/selfservice/strategy/password"
	"github.com/ory/kratos/session"
	"github.com/ory/kratos/x"
	"github.com/ory/kratos/x/nosurfx"
	"github.com/ory/nosurf"
	"github.com/ory/x/contextx"
	"github.com/ory/x/dbal"
	"github.com/ory/x/healthx"
	"github.com/ory/x/jsonnetsecure"
	"github.com/ory/x/logrusx"
	"github.com/ory/x/otelx"
	"github.com/ory/x/popx"
	prometheus "github.com/ory/x/prometheusx"
)

type Registry interface {
	dbal.Driver

	Init(ctx context.Context, ctxer contextx.Contextualizer, opts ...RegistryOption) error

	SetLogger(l *logrusx.Logger)
	SetJSONNetVMProvider(jsonnetsecure.VMProvider)

	WithCSRFHandler(c nosurf.Handler)
	WithCSRFTokenGenerator(cg nosurfx.CSRFToken)

	MetricsHandler() *prometheus.Handler
	HealthHandler(ctx context.Context) *healthx.Handler
	CookieManager(ctx context.Context) sessions.StoreExact
	ContinuityCookieManager(ctx context.Context) sessions.StoreExact

	RegisterRoutes(ctx context.Context, public *x.RouterPublic, admin *x.RouterAdmin)
	RegisterPublicRoutes(ctx context.Context, public *x.RouterPublic)
	RegisterAdminRoutes(ctx context.Context, admin *x.RouterAdmin)
	PrometheusManager() *prometheus.MetricsManager
	Tracer(context.Context) *otelx.Tracer
	SetTracer(*otelx.Tracer)

	config.Provider
	CourierConfig() config.CourierConfigs
	SetConfig(c *config.Config)
	SetContextualizer(ctxer contextx.Contextualizer)

	nosurfx.CSRFProvider
	x.WriterProvider
	x.LoggingProvider
	x.HTTPClientProvider
	jsonnetsecure.VMProvider

	continuity.ManagementProvider
	continuity.PersistenceProvider

	cipher.Provider

	courier.Provider

	persistence.Provider

	errorx.ManagementProvider
	errorx.HandlerProvider
	errorx.PersistenceProvider

	hash.HashProvider

	identity.HandlerProvider
	identity.ValidationProvider
	identity.PoolProvider
	identity.PrivilegedPoolProvider
	identity.ManagementProvider
	identity.ActiveCredentialsCounterStrategyProvider

	courier.HandlerProvider
	courier.PersistenceProvider

	schema.HandlerProvider
	schema.IdentitySchemaProvider

	password2.ValidationProvider

	session.HandlerProvider
	session.ManagementProvider
	session.PersistenceProvider
	session.TokenizerProvider

	settings.HandlerProvider
	settings.ErrorHandlerProvider
	settings.FlowPersistenceProvider
	settings.StrategyProvider

	login.FlowPersistenceProvider
	login.ErrorHandlerProvider
	login.HooksProvider
	login.HookExecutorProvider
	login.HandlerProvider
	login.StrategyProvider

	logout.HandlerProvider

	registration.FlowPersistenceProvider
	registration.ErrorHandlerProvider
	registration.HooksProvider
	registration.HookExecutorProvider
	registration.HandlerProvider
	registration.StrategyProvider

	verification.FlowPersistenceProvider
	verification.ErrorHandlerProvider
	verification.HandlerProvider
	verification.StrategyProvider

	sessiontokenexchange.PersistenceProvider

	link.SenderProvider
	link.VerificationTokenPersistenceProvider
	link.RecoveryTokenPersistenceProvider

	code.SenderProvider
	code.RecoveryCodePersistenceProvider

	recovery.FlowPersistenceProvider
	recovery.ErrorHandlerProvider
	recovery.HandlerProvider
	recovery.StrategyProvider

	nosurfx.CSRFTokenGeneratorProvider
}

func NewRegistryFromDSN(ctx context.Context, c *config.Config, l *logrusx.Logger) (*RegistryDefault, error) {
	reg := NewRegistryDefault()

	tracer, err := otelx.New("Ory Kratos", l, c.Tracing(ctx))
	if err != nil {
		l.WithError(err).Fatalf("failed to initialize tracer")
		tracer = otelx.NewNoop(l, c.Tracing(ctx))
	}
	reg.SetTracer(tracer)
	reg.SetLogger(l)
	reg.SetConfig(c)

	return reg, nil
}

type options struct {
	skipNetworkInit               bool
	config                        *config.Config
	configOptions                 []configx.OptionModifier
	replaceTracer                 func(*otelx.Tracer) *otelx.Tracer
	replaceIdentitySchemaProvider func(Registry) schema.IdentitySchemaProvider
	inspect                       func(Registry) error
	extraMigrations               []fs.FS
	extraGoMigrations             popx.Migrations
	replacementStrategies         []NewStrategy
	extraHooks                    map[string]func(config.SelfServiceHook) any
	extraHandlers                 []NewHandlerRegistrar
	disableMigrationLogging       bool
	jsonnetPool                   jsonnetsecure.Pool
	serviceLocatorOptions         []servicelocatorx.Option
	dbOpts                        []func(details *pop.ConnectionDetails)
}

type RegistryOption func(*options)

// WithDBOptions adds database connection options that will be applied to the
// underlying connection.
func WithDBOptions(opts ...func(details *pop.ConnectionDetails)) RegistryOption {
	return func(o *options) {
		o.dbOpts = append(o.dbOpts, opts...)
	}
}

func SkipNetworkInit(o *options) {
	o.skipNetworkInit = true
}

func WithJsonnetPool(pool jsonnetsecure.Pool) RegistryOption {
	return func(o *options) {
		o.jsonnetPool = pool
	}
}

func WithConfig(config *config.Config) RegistryOption {
	return func(o *options) {
		o.config = config
	}
}

func WithConfigOptions(opts ...configx.OptionModifier) RegistryOption {
	return func(o *options) {
		o.configOptions = append(o.configOptions, opts...)
	}
}

func WithIdentitySchemaProvider(f func(r Registry) schema.IdentitySchemaProvider) RegistryOption {
	return func(o *options) {
		o.replaceIdentitySchemaProvider = f
	}
}

func ReplaceTracer(f func(*otelx.Tracer) *otelx.Tracer) RegistryOption {
	return func(o *options) {
		o.replaceTracer = f
	}
}

type NewStrategy func(deps any) any

// WithReplaceStrategies adds a strategy to the registry. This is useful if you want to
// add a custom strategy to the registry. Default strategies with the same
// name/ID will be overwritten.
func WithReplaceStrategies(s ...NewStrategy) RegistryOption {
	return func(o *options) {
		o.replacementStrategies = append(o.replacementStrategies, s...)
	}
}

func WithExtraHooks(hooks map[string]func(config.SelfServiceHook) any) RegistryOption {
	return func(o *options) {
		o.extraHooks = hooks
	}
}

type NewHandlerRegistrar func(deps any) x.HandlerRegistrar

func WithExtraHandlers(handlers ...NewHandlerRegistrar) RegistryOption {
	return func(o *options) {
		o.extraHandlers = handlers
	}
}

func Inspect(f func(reg Registry) error) RegistryOption {
	return func(o *options) {
		o.inspect = f
	}
}

func WithExtraMigrations(m ...fs.FS) RegistryOption {
	return func(o *options) {
		o.extraMigrations = append(o.extraMigrations, m...)
	}
}

func WithExtraGoMigrations(m ...popx.Migration) RegistryOption {
	return func(o *options) {
		o.extraGoMigrations = append(o.extraGoMigrations, m...)
	}
}

func WithDisabledMigrationLogging() RegistryOption {
	return func(o *options) {
		o.disableMigrationLogging = true
	}
}

func WithServiceLocatorOptions(opts ...servicelocatorx.Option) RegistryOption {
	return func(o *options) {
		o.serviceLocatorOptions = append(o.serviceLocatorOptions, opts...)
	}
}

func newOptions(os []RegistryOption) *options {
	o := new(options)
	for _, f := range os {
		f(o)
	}
	return o
}
