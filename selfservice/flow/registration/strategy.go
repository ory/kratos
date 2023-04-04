// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package registration

import (
	"context"
	"net/http"

	"github.com/ory/kratos/ui/node"

	"github.com/pkg/errors"

	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/x"
)

type Strategy interface {
	ID() identity.CredentialsType
	NodeGroup() node.UiNodeGroup
	RegisterRegistrationRoutes(*x.RouterPublic)
	PopulateRegistrationMethod(r *http.Request, sr *Flow) error
	Register(w http.ResponseWriter, r *http.Request, f *Flow, i *identity.Identity) (err error)
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
		ss.RegisterRegistrationRoutes(r)
	}
}

type StrategyProvider interface {
	RegistrationStrategies(ctx context.Context) Strategies
	AllRegistrationStrategies() Strategies
}
