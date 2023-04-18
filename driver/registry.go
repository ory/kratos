// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package driver

import (
	"context"

	"github.com/ory/kratos/selfservice/sessiontokenexchange"
	"github.com/ory/x/contextx"
	"github.com/ory/x/jsonnetsecure"
	"github.com/ory/x/otelx"
	prometheus "github.com/ory/x/prometheusx"

	"github.com/gorilla/sessions"
	"github.com/pkg/errors"

	"github.com/ory/nosurf"

	"github.com/ory/x/logrusx"

	"github.com/ory/kratos/continuity"
	"github.com/ory/kratos/courier"
	"github.com/ory/kratos/hash"
	"github.com/ory/kratos/schema"
	"github.com/ory/kratos/selfservice/flow/recovery"
	"github.com/ory/kratos/selfservice/flow/settings"
	"github.com/ory/kratos/selfservice/flow/verification"
	"github.com/ory/kratos/selfservice/strategy/code"
	"github.com/ory/kratos/selfservice/strategy/link"

	"github.com/ory/x/healthx"

	"github.com/ory/kratos/persistence"
	"github.com/ory/kratos/selfservice/flow/login"
	"github.com/ory/kratos/selfservice/flow/logout"
	"github.com/ory/kratos/selfservice/flow/registration"

	"github.com/ory/kratos/x"

	"github.com/ory/x/dbal"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/selfservice/errorx"
	password2 "github.com/ory/kratos/selfservice/strategy/password"
	"github.com/ory/kratos/session"
)

type Registry interface {
	dbal.Driver

	Init(ctx context.Context, ctxer contextx.Contextualizer, opts ...RegistryOption) error

	WithLogger(l *logrusx.Logger) Registry
	WithJsonnetVMProvider(jsonnetsecure.VMProvider) Registry

	WithCSRFHandler(c nosurf.Handler)
	WithCSRFTokenGenerator(cg x.CSRFToken)

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
	WithConfig(c *config.Config) Registry
	WithContextualizer(ctxer contextx.Contextualizer) Registry

	x.CSRFProvider
	x.WriterProvider
	x.LoggingProvider
	x.HTTPClientProvider
	jsonnetsecure.VMProvider

	continuity.ManagementProvider
	continuity.PersistenceProvider

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
	schema.IdentityTraitsProvider

	password2.ValidationProvider

	session.HandlerProvider
	session.ManagementProvider
	session.PersistenceProvider

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

	sessiontokenexchange.HandlerProvider
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

	x.CSRFTokenGeneratorProvider
}

func NewRegistryFromDSN(ctx context.Context, c *config.Config, l *logrusx.Logger) (Registry, error) {
	driver, err := dbal.GetDriverFor(c.DSN(ctx))
	if err != nil {
		return nil, errors.WithStack(err)
	}

	registry, ok := driver.(Registry)
	if !ok {
		return nil, errors.Errorf("driver of type %T does not implement interface Registry", driver)
	}

	tracer, err := otelx.New("Ory Kratos", l, c.Tracing(ctx))
	if err != nil {
		l.WithError(err).Fatalf("failed to initialize tracer")
		tracer = otelx.NewNoop(l, c.Tracing(ctx))
	}
	registry.SetTracer(tracer)

	return registry.WithLogger(l).WithConfig(c), nil
}

type options struct {
	skipNetworkInit bool
	config          *config.Config
	replaceTracer   func(*otelx.Tracer) *otelx.Tracer
	inspect         func(Registry) error
}

type RegistryOption func(*options)

func SkipNetworkInit(o *options) {
	o.skipNetworkInit = true
}

func WithConfig(config *config.Config) func(o *options) {
	return func(o *options) {
		o.config = config
	}
}

func ReplaceTracer(f func(*otelx.Tracer) *otelx.Tracer) func(o *options) {
	return func(o *options) {
		o.replaceTracer = f
	}
}

func Inspect(f func(reg Registry) error) func(o *options) {
	return func(o *options) {
		o.inspect = f
	}
}

func newOptions(os []RegistryOption) *options {
	o := new(options)
	for _, f := range os {
		f(o)
	}
	return o
}
