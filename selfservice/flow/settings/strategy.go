package settings

import (
	"net/http"

	"github.com/pkg/errors"

	"github.com/ory/kratos/session"
	"github.com/ory/kratos/x"
)

type Strategy interface {
	SettingsStrategyID() string
	RegisterSettingsRoutes(*x.RouterPublic)
	PopulateSettingsMethod(*http.Request, *session.Session, *Request) error
}

type Strategies []Strategy

func (s Strategies) Strategy(id string) (Strategy, error) {
	ids := make([]string, len(s))
	for k, ss := range s {
		ids[k] = ss.SettingsStrategyID()
		if ss.SettingsStrategyID() == id {
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
		ss.RegisterSettingsRoutes(r)
	}
}

type StrategyProvider interface {
	SettingsStrategies() Strategies
}
