// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package link

import (
	"github.com/ory/kratos/courier"
	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/schema"
	"github.com/ory/kratos/selfservice/errorx"
	"github.com/ory/kratos/selfservice/flow/recovery"
	"github.com/ory/kratos/selfservice/flow/settings"
	"github.com/ory/kratos/selfservice/flow/verification"
	"github.com/ory/kratos/session"
	"github.com/ory/kratos/ui/container"
	"github.com/ory/kratos/ui/node"
	"github.com/ory/kratos/x"
	"github.com/ory/x/decoderx"
)

var (
	_ recovery.Strategy      = new(Strategy)
	_ recovery.AdminHandler  = new(Strategy)
	_ recovery.PublicHandler = new(Strategy)
)

var (
	_ verification.Strategy      = new(Strategy)
	_ verification.AdminHandler  = new(Strategy)
	_ verification.PublicHandler = new(Strategy)
)

type (
	// FlowMethod contains the configuration for this selfservice strategy.
	FlowMethod struct {
		*container.Container
	}

	strategyDependencies interface {
		x.CSRFProvider
		x.CSRFTokenGeneratorProvider
		x.WriterProvider
		x.LoggingProvider

		config.Provider

		session.HandlerProvider
		session.ManagementProvider
		settings.HandlerProvider
		settings.FlowPersistenceProvider

		identity.ValidationProvider
		identity.ManagementProvider
		identity.PoolProvider
		identity.PrivilegedPoolProvider

		courier.Provider

		errorx.ManagementProvider

		recovery.ErrorHandlerProvider
		recovery.FlowPersistenceProvider
		recovery.StrategyProvider
		recovery.HookExecutorProvider

		verification.ErrorHandlerProvider
		verification.FlowPersistenceProvider
		verification.StrategyProvider
		verification.HookExecutorProvider
		verification.HandlerProvider

		RecoveryTokenPersistenceProvider
		VerificationTokenPersistenceProvider
		SenderProvider

		schema.IdentityTraitsProvider
	}

	Strategy struct {
		d  strategyDependencies
		dx *decoderx.HTTP
	}
)

func NewStrategy(d strategyDependencies) *Strategy {
	return &Strategy{d: d, dx: decoderx.NewHTTP()}
}

func (s *Strategy) NodeGroup() node.UiNodeGroup {
	return node.LinkGroup
}
