package password

import (
	"github.com/go-playground/form"
	"github.com/justinas/nosurf"
	"gopkg.in/go-playground/validator.v9"

	"github.com/ory/hive/driver/configuration"
	"github.com/ory/hive/errorx"
	"github.com/ory/hive/identity"
	"github.com/ory/hive/selfservice"
	"github.com/ory/hive/x"
)

var _ selfservice.Strategy = new(Strategy)

type registrationStrategyDependencies interface {
	x.LoggingProvider
	x.WriterProvider
	errorx.ManagementProvider
	selfservice.RegistrationRequestManagementProvider
	identity.ValidationProvider
	ValidationProvider
	HashProvider
	selfservice.RegistrationExecutionProvider
	selfservice.PostRegistrationHookProvider

	selfservice.LoginRequestManagementProvider
	identity.PoolProvider
	selfservice.LoginExecutionProvider
	selfservice.PostLoginHookProvider
}

type Strategy struct {
	c   configuration.Provider
	d   registrationStrategyDependencies
	dc  *form.Decoder
	v   *validator.Validate
	dec *RegistrationFormDecoder

	cg csrfGenerator
}

func NewStrategy(
	d registrationStrategyDependencies,
	c configuration.Provider,
) *Strategy {
	return &Strategy{
		c:   c,
		d:   d,
		dc:  form.NewDecoder(),
		v:   validator.New(),
		cg:  nosurf.Token,
		dec: NewRegistrationFormDecoder(),
	}
}

func (s *Strategy) WithTokenGenerator(g csrfGenerator) *Strategy {
	s.cg = g
	return s
}

func (s *Strategy) ID() identity.CredentialsType {
	return CredentialsType
}

func (s *Strategy) SetRoutes(r *x.RouterPublic) {
	s.setRegistrationRoutes(r)
	s.setLoginRoutes(r)
}
