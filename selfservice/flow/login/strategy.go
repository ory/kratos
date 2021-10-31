package login

import (
	"context"
	"net/http"

	"github.com/ory/kratos/session"

	"github.com/pkg/errors"

	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/ui/node"
	"github.com/ory/kratos/x"
)

type Strategy interface {
	ID() identity.CredentialsType
	NodeGroup() node.Group
	RegisterLoginRoutes(*x.RouterPublic)
	PopulateLoginMethod(r *http.Request, requestedAAL identity.AuthenticatorAssuranceLevel, sr *Flow) error
	Login(w http.ResponseWriter, r *http.Request, f *Flow, ss *session.Session) (i *identity.Identity, err error)
}

type Strategies []Strategy

func (s Strategies) Strategy(id identity.CredentialsType) (Strategy, error) {
	ids := make([]identity.CredentialsType, len(s))
	for k, ss := range s {
		ids[k] = ss.ID()
		if ss.ID() == id {
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
		ss.RegisterLoginRoutes(r)
	}
}

type StrategyProvider interface {
	AllLoginStrategies() Strategies
	LoginStrategies(ctx context.Context) Strategies
}
