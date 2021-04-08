package verification

import (
	"context"
	"net/http"

	"github.com/pkg/errors"

	"github.com/ory/kratos/x"
)

const (
	StrategyVerificationLinkName = "link"
)

type (
	Strategy interface {
		VerificationStrategyID() string
		PopulateVerificationMethod(*http.Request, *Flow) error
	}
	AdminHandler interface {
		RegisterAdminVerificationRoutes(admin *x.RouterAdmin)
	}
	PublicHandler interface {
		RegisterPublicVerificationRoutes(public *x.RouterPublic)
	}
	Strategies       []Strategy
	StrategyProvider interface {
		VerificationStrategies(ctx context.Context) Strategies
	}
)

func (s Strategies) Strategy(id string) (Strategy, error) {
	ids := make([]string, len(s))
	for k, ss := range s {
		ids[k] = ss.VerificationStrategyID()
		if ss.VerificationStrategyID() == id {
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
		if h, ok := ss.(PublicHandler); ok {
			h.RegisterPublicVerificationRoutes(r)
		}
	}
}

func (s Strategies) RegisterAdminRoutes(r *x.RouterAdmin) {
	for _, ss := range s {
		if h, ok := ss.(AdminHandler); ok {
			h.RegisterAdminVerificationRoutes(r)
		}
	}
}
