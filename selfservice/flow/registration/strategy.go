// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package registration

import (
	"context"
	"net/http"

	"github.com/ory/kratos/ui/node"

	"github.com/pkg/errors"

	"github.com/ory/kratos/identity"
)

type Strategy interface {
	ID() identity.CredentialsType
	NodeGroup() node.UiNodeGroup
	// Register attempts to register the identity.
	//
	// A strategy may complete the flow itself (for example by redirecting to an
	// OIDC provider). In that case it returns completed == true with a nil
	// error: this is a control signal, not a failure, so callers must stop
	// processing the flow without treating it as an error.
	Register(w http.ResponseWriter, r *http.Request, f *Flow, i *identity.Identity) (completed bool, err error)
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

type StrategyFilter func(strategy Strategy) bool

type StrategyProvider interface {
	RegistrationStrategies(ctx context.Context, filters ...StrategyFilter) Strategies
	AllRegistrationStrategies() Strategies
}
