// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package driver

import (
	"context"

	"github.com/pkg/errors"

	"github.com/ory/herodot"
	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/selfservice/flow/recovery"
	"github.com/ory/kratos/selfservice/strategy/code"
)

func (m *RegistryDefault) RecoveryFlowErrorHandler() *recovery.ErrorHandler {
	if m.selfserviceRecoveryErrorHandler == nil {
		m.selfserviceRecoveryErrorHandler = recovery.NewErrorHandler(m)
	}

	return m.selfserviceRecoveryErrorHandler
}

func (m *RegistryDefault) RecoveryHandler() *recovery.Handler {
	if m.selfserviceRecoveryHandler == nil {
		m.selfserviceRecoveryHandler = recovery.NewHandler(m)
	}

	return m.selfserviceRecoveryHandler
}

func (m *RegistryDefault) RecoveryStrategies(ctx context.Context) (recoveryStrategies recovery.Strategies) {
	for _, strategy := range m.selfServiceStrategies() {
		if s, ok := strategy.(recovery.Strategy); ok {
			if m.Config().SelfServiceStrategy(ctx, s.RecoveryStrategyID()).Enabled {
				recoveryStrategies = append(recoveryStrategies, s)
			}
		}
	}
	return
}

// GetActiveRecoveryStrategy returns the currently active recovery strategy
// If no recovery strategy has been set, an error is returned
func (m *RegistryDefault) GetActiveRecoveryStrategy(ctx context.Context) (recovery.Strategy, error) {
	as := m.Config().SelfServiceFlowRecoveryUse(ctx)
	s, err := m.RecoveryStrategies(ctx).Strategy(as)
	if err != nil {
		return nil, errors.WithStack(herodot.ErrBadRequest.
			WithReasonf("You attempted recovery using %s, which is not enabled or does not exist. An administrator needs to enable this recovery method.", as))
	}
	return s, nil
}

func (m *RegistryDefault) AllRecoveryStrategies() (recoveryStrategies recovery.Strategies) {
	for _, strategy := range m.selfServiceStrategies() {
		if s, ok := strategy.(recovery.Strategy); ok {
			recoveryStrategies = append(recoveryStrategies, s)
		}
	}
	return
}

func (m *RegistryDefault) RecoveryExecutor() *recovery.HookExecutor {
	if m.selfserviceRecoveryExecutor == nil {
		m.selfserviceRecoveryExecutor = recovery.NewHookExecutor(m)
	}
	return m.selfserviceRecoveryExecutor
}

func (m *RegistryDefault) PreRecoveryHooks(ctx context.Context) ([]recovery.PreHookExecutor, error) {
	return getHooks[recovery.PreHookExecutor](m, "", m.Config().SelfServiceFlowRecoveryBeforeHooks(ctx))
}

func (m *RegistryDefault) PostRecoveryHooks(ctx context.Context) ([]recovery.PostHookExecutor, error) {
	return getHooks[recovery.PostHookExecutor](m, config.HookGlobal, m.Config().SelfServiceFlowRecoveryAfterHooks(ctx, config.HookGlobal))
}

func (m *RegistryDefault) CodeSender() *code.Sender {
	if m.selfserviceCodeSender == nil {
		m.selfserviceCodeSender = code.NewSender(m)
	}

	return m.selfserviceCodeSender
}
