package driver

import (
	"github.com/go-errors/errors"
	"github.com/gorilla/sessions"
	"github.com/justinas/nosurf"
	"github.com/sirupsen/logrus"

	"github.com/ory/kratos/schema"
	"github.com/ory/kratos/selfservice/flow/verify"

	"github.com/ory/x/healthx"

	"github.com/ory/kratos/persistence"
	"github.com/ory/kratos/selfservice/flow/login"
	"github.com/ory/kratos/selfservice/flow/logout"
	"github.com/ory/kratos/selfservice/flow/profile"
	"github.com/ory/kratos/selfservice/flow/registration"

	"github.com/ory/kratos/x"

	"github.com/ory/x/dbal"

	"github.com/ory/kratos/driver/configuration"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/selfservice/errorx"
	password2 "github.com/ory/kratos/selfservice/strategy/password"
	"github.com/ory/kratos/session"
)

type Registry interface {
	dbal.Driver

	Init() error

	WithConfig(c configuration.Provider) Registry
	WithLogger(l logrus.FieldLogger) Registry

	BuildVersion() string
	BuildDate() string
	BuildHash() string
	WithBuildInfo(version, hash, date string) Registry

	WithCSRFHandler(c *nosurf.CSRFHandler)
	CSRFHandler() *nosurf.CSRFHandler
	HealthHandler() *healthx.Handler
	CookieManager() sessions.Store

	x.WriterProvider
	x.LoggingProvider

	persistence.Provider

	errorx.ManagementProvider
	errorx.HandlerProvider
	errorx.PersistenceProvider

	identity.HandlerProvider
	identity.ValidationProvider
	identity.PoolProvider
	identity.PrivilegedPoolProvider
	identity.ManagementProvider

	schema.HandlerProvider

	password2.ValidationProvider
	password2.HashProvider

	session.HandlerProvider
	session.ManagementProvider
	session.PersistenceProvider

	profile.HandlerProvider
	profile.ErrorHandlerProvider
	profile.RequestPersistenceProvider

	login.RequestPersistenceProvider
	login.ErrorHandlerProvider
	login.HooksProvider
	login.HookExecutorProvider
	login.HandlerProvider
	login.StrategyProvider

	logout.HandlerProvider

	registration.RequestPersistenceProvider
	registration.ErrorHandlerProvider
	registration.HooksProvider
	registration.HookExecutorProvider
	registration.HandlerProvider
	registration.StrategyProvider

	verify.PersistenceProvider
	verify.ErrorHandlerProvider
	verify.SenderProvider

	x.CSRFTokenGeneratorProvider
}

type selfServiceStrategy interface {
	login.Strategy
	registration.Strategy
}

func NewRegistry(c configuration.Provider) (Registry, error) {
	driver, err := dbal.GetDriverFor(c.DSN())
	if err != nil {
		return nil, err
	}

	registry, ok := driver.(Registry)
	if !ok {
		return nil, errors.Errorf("driver of type %T does not implement interface Registry", driver)
	}

	return registry, nil
}
