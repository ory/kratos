package driver

import (
	"net/http"

	"github.com/go-errors/errors"
	"github.com/gorilla/sessions"
	"github.com/justinas/nosurf"
	"github.com/sirupsen/logrus"

	"github.com/ory/herodot"

	"github.com/ory/hive-cloud/hive/selfservice"

	"github.com/ory/x/dbal"

	"github.com/ory/hive-cloud/hive/driver/configuration"
	"github.com/ory/hive-cloud/hive/errorx"
	"github.com/ory/hive-cloud/hive/identity"
	password2 "github.com/ory/hive-cloud/hive/selfservice/password"
	"github.com/ory/hive-cloud/hive/session"
)

type Registry interface {
	dbal.Driver

	WithConfig(c configuration.Provider) Registry
	WithLogger(l logrus.FieldLogger) Registry

	Logger() logrus.FieldLogger
	Writer() herodot.Writer

	ErrorManager() errorx.Manager
	ErrorHandler() *errorx.Handler

	IdentityHandler() *identity.Handler
	IdentityPool() identity.Pool

	PasswordHasher() password2.Hasher
	PasswordValidator() password2.Validator

	SessionHandler() *session.Handler
	SessionManager() session.Manager

	StrategyHandler() *selfservice.StrategyHandler
	SelfServiceStrategies() []selfservice.Strategy

	CookieManager() sessions.Store

	NoSurf(http.Handler) *nosurf.CSRFHandler
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
