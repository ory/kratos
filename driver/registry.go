package driver

import (
	"context"
	"net/url"
	"strings"

	"github.com/gorilla/sessions"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/ory/kratos/continuity"
	"github.com/ory/kratos/courier"
	"github.com/ory/kratos/schema"
	"github.com/ory/kratos/selfservice/flow/settings"
	"github.com/ory/kratos/selfservice/flow/verify"

	"github.com/ory/x/healthx"

	"github.com/ory/kratos/persistence"
	"github.com/ory/kratos/selfservice/flow/login"
	"github.com/ory/kratos/selfservice/flow/logout"
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

	WithCSRFHandler(c x.CSRFHandler)
	WithCSRFTokenGenerator(cg x.CSRFToken)

	HealthHandler() *healthx.Handler
	CookieManager() sessions.Store

	x.CSRFProvider
	x.WriterProvider
	x.LoggingProvider

	continuity.ManagementProvider
	continuity.PersistenceProvider

	courier.Provider

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

	settings.HandlerProvider
	settings.ErrorHandlerProvider
	settings.RequestPersistenceProvider
	settings.StrategyProvider

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
	verify.HandlerProvider

	x.CSRFTokenGeneratorProvider
}

func NewRegistry(c configuration.Provider) (Registry, error) {
	dsn := c.DSN()
	driver, err := dbal.GetDriverFor(dsn)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	registry, ok := driver.(Registry)
	if !ok {
		return nil, errors.Errorf("driver of type %T does not implement interface Registry", driver)
	}

	// if dsn is memory we have to run the migrations on every start
	if urlParts := strings.SplitN(dsn, "?", 1); len(urlParts) == 2 && strings.HasPrefix(dsn, "sqlite://") {
		queryVals, err := url.ParseQuery(urlParts[1])
		if err != nil {
			return nil, errors.WithMessage(errors.WithStack(err), "unable to parse the DSN url")
		}
		if queryVals.Get("mode") == "memory" {
			registry.Logger().Print("Kratos is running migrations on every startup as DSN is memory.\n")
			registry.Logger().Print("This means your data is lost when Kratos terminates.\n")
			if err := registry.Persister().MigrateUp(context.Background()); err != nil {
				return nil, err
			}
		}
	}
	return registry, nil
}
