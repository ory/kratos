// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package recovery

import (
	"context"
	"net/http"

	"github.com/pkg/errors"

	"github.com/ory/herodot"
	"github.com/ory/kratos/ui/node"

	"github.com/ory/kratos/x"
)

//swagger:enum RecoveryMethod
type RecoveryMethod string

const (
	RecoveryStrategyLink RecoveryMethod = "link"
	RecoveryStrategyCode RecoveryMethod = "code"
)

type (
	Strategy interface {
		RecoveryStrategyID() string
		NodeGroup() node.UiNodeGroup
		PopulateRecoveryMethod(*http.Request, *Flow) error
		Recover(w http.ResponseWriter, r *http.Request, f *Flow) (err error)
	}
	AdminHandler interface {
		RegisterAdminRecoveryRoutes(admin *x.RouterAdmin)
	}
	PublicHandler interface {
		RegisterPublicRecoveryRoutes(public *x.RouterPublic)
	}
	Strategies       []Strategy
	StrategyProvider interface {
		AllRecoveryStrategies() Strategies
		RecoveryStrategies(ctx context.Context) Strategies
		GetActiveRecoveryStrategy(ctx context.Context) (Strategy, error)
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

	return nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("unable to find strategy for %s have %v", id, ids))
}

func (s Strategies) RegisterPublicRoutes(r *x.RouterPublic) {
	for _, ss := range s {
		if h, ok := ss.(PublicHandler); ok {
			h.RegisterPublicRecoveryRoutes(r)
		}
	}
}

func (s Strategies) RegisterAdminRoutes(r *x.RouterAdmin) {
	for _, ss := range s {
		if h, ok := ss.(AdminHandler); ok {
			h.RegisterAdminRecoveryRoutes(r)
		}
	}
}
