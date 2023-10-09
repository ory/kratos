// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package verification

import (
	"context"
	"net/http"

	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/ui/node"

	"github.com/pkg/errors"

	"github.com/ory/kratos/x"
)

//swagger:enum VerificationStrategy
type VerificationStrategy string

const (
	VerificationStrategyLink VerificationStrategy = "link"
	VerificationStrategyCode VerificationStrategy = "code"
)

type (
	Strategy interface {
		VerificationStrategyID() string
		NodeGroup() node.UiNodeGroup
		PopulateVerificationMethod(*http.Request, *Flow) error
		Verify(w http.ResponseWriter, r *http.Request, f *Flow) (err error)
		SendVerificationEmail(context.Context, *Flow, *identity.Identity, *identity.VerifiableAddress) error
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
		AllVerificationStrategies() Strategies
		GetActiveVerificationStrategy(context.Context) (Strategy, error)
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
