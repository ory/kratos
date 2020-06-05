package recovery

import (
	"net/http"

	"github.com/pkg/errors"

	"github.com/ory/kratos/x"
)

const (
	StrategyRecoveryTokenName = "token"
)

type (
	Strategy interface {
		RecoveryStrategyID() string
		RegisterRecoveryRoutes(*x.RouterPublic)
		PopulateRecoveryMethod(*http.Request, *Request) error
	}
	Strategies       []Strategy
	StrategyProvider interface {
		RecoveryStrategies() Strategies
	}
)

func (s Strategies) Strategy(id string) (Strategy, error) {
	ids := make([]string, len(s))
	for k, ss := range s {
		ids[k] = ss.RecoveryStrategyID()
		if ss.RecoveryStrategyID() == id {
			return ss, nil
		}
	}

	return nil, errors.Errorf(`unable to find strategy for %s have %v`, id, ids)
}

func (s Strategies) MustStrategy(id string) Strategy {
	strategy, err := s.Strategy(id)
	if err != nil {
		panic(err)
	}
	return strategy
}

func (s Strategies) RegisterPublicRoutes(r *x.RouterPublic) {
	for _, ss := range s {
		ss.RegisterRecoveryRoutes(r)
	}
}
