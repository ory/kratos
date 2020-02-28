package password

import (
	"github.com/justinas/nosurf"
	"gopkg.in/go-playground/validator.v9"

	"github.com/ory/kratos/selfservice/flow/login"
	"github.com/ory/kratos/selfservice/flow/registration"
	"github.com/ory/kratos/selfservice/form"

	"github.com/ory/kratos/driver/configuration"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/selfservice/errorx"
	"github.com/ory/kratos/session"
	"github.com/ory/kratos/x"
)

var _ login.Strategy = new(Strategy)
var _ registration.Strategy = new(Strategy)

type registrationStrategyDependencies interface {
	x.LoggingProvider
	x.WriterProvider
	errorx.ManagementProvider
	ValidationProvider
	HashProvider

	registration.HandlerProvider
	registration.HooksProvider
	registration.ErrorHandlerProvider
	registration.HookExecutorProvider
	registration.RequestPersistenceProvider

	login.HooksProvider
	login.ErrorHandlerProvider
	login.HookExecutorProvider
	login.RequestPersistenceProvider
	login.HandlerProvider

	identity.PrivilegedPoolProvider
	identity.ValidationProvider

	session.HandlerProvider
	session.ManagementProvider
}

type Strategy struct {
	c  configuration.Provider
	d  registrationStrategyDependencies
	v  *validator.Validate
	cg form.CSRFGenerator
}

func NewStrategy(
	d registrationStrategyDependencies,
	c configuration.Provider,
) *Strategy {
	return &Strategy{
		c:  c,
		d:  d,
		v:  validator.New(),
		cg: nosurf.Token,
	}
}

func (s *Strategy) WithTokenGenerator(g form.CSRFGenerator) {
	s.cg = g
}

func (s *Strategy) ID() identity.CredentialsType {
	return identity.CredentialsTypePassword
}

func (s *Strategy) RegistrationStrategyID() identity.CredentialsType {
	return s.ID()
}

func (s *Strategy) LoginStrategyID() identity.CredentialsType {
	return s.ID()
}
