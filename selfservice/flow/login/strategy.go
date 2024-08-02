// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package login

import (
	"context"
	"net/http"

	"github.com/pkg/errors"

	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/session"
	"github.com/ory/kratos/ui/node"
	"github.com/ory/kratos/x"
	"github.com/ory/x/sqlxx"
)

type Strategy interface {
	ID() identity.CredentialsType
	NodeGroup() node.UiNodeGroup
	RegisterLoginRoutes(*x.RouterPublic)
	PopulateLoginMethod(r *http.Request, requestedAAL identity.AuthenticatorAssuranceLevel, sr *Flow) error
	Login(w http.ResponseWriter, r *http.Request, f *Flow, sess *session.Session) (i *identity.Identity, err error)
	CompletedAuthenticationMethod(ctx context.Context, methods session.AuthenticationMethods, credentialsConfig sqlxx.JSONRawMessage) (*session.AuthenticationMethod, error)
}

type Strategies []Strategy

type LinkableStrategy interface {
	Link(ctx context.Context, i *identity.Identity, credentials sqlxx.JSONRawMessage) error
}

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

type StrategyFilter func(strategy Strategy) bool

type StrategyProvider interface {
	AllLoginStrategies() Strategies
	LoginStrategies(ctx context.Context, filters ...StrategyFilter) Strategies
}
