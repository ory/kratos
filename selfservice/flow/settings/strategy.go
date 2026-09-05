// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package settings

import (
	"context"
	"net/http"
	"reflect"

	"github.com/ory/kratos/session"

	"github.com/ory/kratos/ui/node"

	"github.com/pkg/errors"

	"github.com/ory/kratos/identity"
)

const (
	StrategyProfile = "profile"
)

var pkgName = reflect.TypeOf(Strategies{}).PkgPath()

type Strategy interface {
	SettingsStrategyID() string
	NodeGroup() node.UiNodeGroup
	PopulateSettingsMethod(context.Context, *http.Request, *identity.Identity, *Flow) error
	// Settings attempts to update the identity's settings.
	//
	// A strategy may complete the flow itself (for example by redirecting to an
	// OIDC provider for account linking). In that case it returns completed ==
	// true with a nil error: this is a control signal, not a failure, so callers
	// must stop processing without treating it as an error.
	Settings(ctx context.Context, w http.ResponseWriter, r *http.Request, f *Flow, s *session.Session) (*UpdateContext, bool, error)
}

type Strategies []Strategy

func (s Strategies) Strategy(id string) (Strategy, error) {
	ids := make([]string, len(s))
	for k, ss := range s {
		ids[k] = ss.SettingsStrategyID()
		if ss.SettingsStrategyID() == id {
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

// StrategyFilter selects which Settings strategies to include when populating a flow.
type StrategyFilter func(strategy Strategy) bool

type StrategyProvider interface {
	SettingsStrategies(ctx context.Context, filters ...StrategyFilter) Strategies
	AllSettingsStrategies() Strategies
}
