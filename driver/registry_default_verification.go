// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package driver

import (
	"context"

	"github.com/pkg/errors"

	"github.com/ory/herodot"
	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/selfservice/flow/verification"
	"github.com/ory/kratos/selfservice/strategy/link"
)

func (m *RegistryDefault) VerificationFlowPersister() verification.FlowPersister {
	return m.persister
}

func (m *RegistryDefault) VerificationFlowErrorHandler() *verification.ErrorHandler {
	if m.selfserviceVerifyErrorHandler == nil {
		m.selfserviceVerifyErrorHandler = verification.NewErrorHandler(m)
	}

	return m.selfserviceVerifyErrorHandler
}

func (m *RegistryDefault) VerificationManager() *identity.Manager {
	if m.selfserviceVerifyManager == nil {
		m.selfserviceVerifyManager = identity.NewManager(m)
	}

	return m.selfserviceVerifyManager
}

func (m *RegistryDefault) VerificationHandler() *verification.Handler {
	if m.selfserviceVerifyHandler == nil {
		m.selfserviceVerifyHandler = verification.NewHandler(m)
	}

	return m.selfserviceVerifyHandler
}

func (m *RegistryDefault) LinkSender() *link.Sender {
	if m.selfserviceLinkSender == nil {
		m.selfserviceLinkSender = link.NewSender(m)
	}

	return m.selfserviceLinkSender
}

// GetActiveVerificationStrategy returns the currently active verification strategy
// If no verification strategy has been set, an error is returned
func (m *RegistryDefault) GetActiveVerificationStrategy(ctx context.Context) (verification.Strategy, error) {
	as := m.Config().SelfServiceFlowVerificationUse(ctx)
	s, err := m.VerificationStrategies(ctx).Strategy(as)
	if err != nil {
		return nil, errors.WithStack(herodot.ErrBadRequest.
			WithReasonf("The active verification strategy %s is not enabled. Please enable it in the configuration.", as))
	}
	return s, nil
}

func (m *RegistryDefault) VerificationStrategies(ctx context.Context) (verificationStrategies verification.Strategies) {
	for _, strategy := range m.selfServiceStrategies() {
		if s, ok := strategy.(verification.Strategy); ok {
			if m.Config().SelfServiceStrategy(ctx, s.VerificationStrategyID()).Enabled {
				verificationStrategies = append(verificationStrategies, s)
			}
		}
	}
	return
}

func (m *RegistryDefault) AllVerificationStrategies() (recoveryStrategies verification.Strategies) {
	for _, strategy := range m.selfServiceStrategies() {
		if s, ok := strategy.(verification.Strategy); ok {
			recoveryStrategies = append(recoveryStrategies, s)
		}
	}

	return
}

func (m *RegistryDefault) VerificationExecutor() *verification.HookExecutor {
	if m.selfserviceVerificationExecutor == nil {
		m.selfserviceVerificationExecutor = verification.NewHookExecutor(m)
	}
	return m.selfserviceVerificationExecutor
}

func (m *RegistryDefault) PreVerificationHooks(ctx context.Context) (b []verification.PreHookExecutor) {
	for _, v := range m.getHooks("", m.Config().SelfServiceFlowVerificationBeforeHooks(ctx)) {
		if hook, ok := v.(verification.PreHookExecutor); ok {
			b = append(b, hook)
		}
	}
	return
}

func (m *RegistryDefault) PostVerificationHooks(ctx context.Context) (b []verification.PostHookExecutor) {
	for _, v := range m.getHooks(config.HookGlobal, m.Config().SelfServiceFlowVerificationAfterHooks(ctx, config.HookGlobal)) {
		if hook, ok := v.(verification.PostHookExecutor); ok {
			b = append(b, hook)
		}
	}

	return
}
