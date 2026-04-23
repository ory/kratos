// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package verification

import (
	"context"
	"net/http"

	"github.com/ory/herodot"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/ui/node"

	"github.com/pkg/errors"
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
	}
	PrimaryStrategy interface {
		Strategy
		SendVerificationCode(context.Context, *Flow, *identity.Identity, identity.VerifiableAddressLike) error
	}
	Strategies       []Strategy
	StrategyProvider interface {
		VerificationStrategies(ctx context.Context) Strategies
		AllVerificationStrategies() Strategies
		GetActiveVerificationStrategies(context.Context) (active Strategies, primary PrimaryStrategy, err error)
	}
)

func (s Strategies) ActiveStrategies(id string) (active Strategies, primary PrimaryStrategy, err error) {
	ids := make([]string, len(s))
	activeStrategies := Strategies{}
	for k, ss := range s {
		ids[k] = ss.VerificationStrategyID()
		if ps, isPrimary := ss.(PrimaryStrategy); ss.VerificationStrategyID() == id || !isPrimary {
			activeStrategies = append(activeStrategies, ss)
			if isPrimary {
				primary = ps
			}
		}
	}

	if primary == nil {
		return nil, nil, errors.WithStack(herodot.ErrInternalServerError().WithReasonf("unable to find strategy for %s have %v", id, ids))
	}

	return activeStrategies, primary, nil
}

func (s Strategies) MustStrategy(id string) Strategy {
	_, strategy, err := s.ActiveStrategies(id)
	if err != nil {
		panic(err)
	}
	return strategy
}
