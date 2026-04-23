// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package recovery

import (
	"context"
	"net/http"

	"github.com/pkg/errors"

	"github.com/ory/herodot"
	"github.com/ory/kratos/ui/node"
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
		IsPrimary() bool
		NodeGroup() node.UiNodeGroup
		PopulateRecoveryMethod(*http.Request, *Flow) error
		Recover(w http.ResponseWriter, r *http.Request, f *Flow) (err error)
	}
	Strategies       []Strategy
	StrategyProvider interface {
		AllRecoveryStrategies() Strategies
		RecoveryStrategies(ctx context.Context) Strategies
		GetActiveRecoveryStrategies(ctx context.Context) (active Strategies, primary Strategy, err error)
	}
)

func (s Strategies) ActiveStrategies(id string) (active Strategies, primary Strategy, err error) {
	ids := make([]string, len(s))
	for k, ss := range s {
		ids[k] = ss.RecoveryStrategyID()
		if ss.RecoveryStrategyID() == id || !ss.IsPrimary() {
			active = append(active, ss)
			if ss.IsPrimary() {
				primary = ss
			}
		}
	}

	if primary == nil {
		return nil, nil, errors.WithStack(herodot.ErrInternalServerError().WithReasonf("unable to find strategy for %s have %v", id, ids))
	}

	return active, primary, nil
}
