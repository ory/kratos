// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package login

import (
	"context"
	"net/http"

	"github.com/pkg/errors"

	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/session"
	"github.com/ory/kratos/ui/node"
	"github.com/ory/x/sqlxx"
)

type Strategy interface {
	ID() identity.CredentialsType
	NodeGroup() node.UiNodeGroup
	// Login attempts to log the identity in.
	//
	// A strategy may complete the flow itself (for example by sending a login
	// code and rendering the next step). In that case it returns completed ==
	// true with a nil error: this is a control signal, not a failure, so
	// callers must stop processing the flow without treating it as an error.
	Login(w http.ResponseWriter, r *http.Request, f *Flow, sess *session.Session) (i *identity.Identity, completed bool, err error)
	CompletedAuthenticationMethod(ctx context.Context) session.AuthenticationMethod
}

type Strategies []Strategy

type LinkableStrategy interface {
	Link(ctx context.Context, i *identity.Identity, credentials sqlxx.JSONRawMessage) error
	CompletedLogin(sess *session.Session, data *flow.DuplicateCredentialsData) error
	SetDuplicateCredentials(f flow.InternalContexter, duplicateIdentifier string, credentials identity.Credentials, provider string) error
}

// FastLoginStrategy is implemented by strategies that can complete a login
// during flow creation (the "fast" path). The returned completed bool is a
// control signal: when true (with a nil error) the strategy has taken over the
// response and the caller must stop processing the flow without recording an
// error.
type FastLoginStrategy interface {
	FastLogin1FA(w http.ResponseWriter, r *http.Request, f *Flow, sess *session.Session) (completed bool, err error)
	FastLogin2FA(w http.ResponseWriter, r *http.Request, f *Flow, sess *session.Session) (completed bool, err error)
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

type StrategyFilter func(strategy Strategy) bool

type StrategyProvider interface {
	AllLoginStrategies() Strategies
	LoginStrategies(ctx context.Context, filters ...StrategyFilter) Strategies
}
