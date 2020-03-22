package profile

import (
	"net/http"

	"github.com/pkg/errors"

	"github.com/ory/kratos/session"
	"github.com/ory/kratos/x"
)

type Strategy interface {
	ProfileManagementStrategyID() string
	RegisterProfileManagementRoutes(*x.RouterPublic)
	PopulateProfileManagementMethod(*http.Request, *session.Session, *Request) error
}

type Strategies []Strategy

func (s Strategies) Strategy(id string) (Strategy, error) {
	ids := make([]string, len(s))
	for k, ss := range s {
		ids[k] = ss.ProfileManagementStrategyID()
		if ss.ProfileManagementStrategyID() == id {
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
		ss.RegisterProfileManagementRoutes(r)
	}
}

type StrategyProvider interface {
	ProfileManagementStrategies() Strategies
}
