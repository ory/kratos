package registration

import (
	"net/http"

	"github.com/pkg/errors"

	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/x"
)

type Strategy interface {
	RegistrationStrategyID() identity.CredentialsType
	RegisterRegistrationRoutes(*x.RouterPublic)
	RegisterPasswordStrengthMeterRoutes(*x.RouterPublic)
	PopulateRegistrationMethod(r *http.Request, sr *Request) error
}

type Strategies []Strategy

func (s Strategies) Strategy(id identity.CredentialsType) (Strategy, error) {
	ids := make([]identity.CredentialsType, len(s))
	for k, ss := range s {
		ids[k] = ss.RegistrationStrategyID()
		if ss.RegistrationStrategyID() == id {
			return ss, nil
		}
	}

	return nil, errors.Errorf(`unable to find strategy for %s have %v`, id, ids)
}

func (s Strategies) MustStrategy(id identity.CredentialsType) Strategy {
	strategy, err := s.Strategy(id)
	if err != nil {
		panic(err)
	}
	return strategy
}

func (s Strategies) RegisterPublicRoutes(r *x.RouterPublic) {
	for _, ss := range s {
		ss.RegisterRegistrationRoutes(r)
		ss.RegisterPasswordStrengthMeterRoutes(r)
	}
}



type StrategyProvider interface {
	RegistrationStrategies() Strategies
}
