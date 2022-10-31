package driver

import (
	"context"

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
	activeRecoveryStrategy := m.Config().SelfServiceFlowRecoveryUse(ctx)
	return m.RecoveryStrategies(ctx).Strategy(activeRecoveryStrategy)
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

func (m *RegistryDefault) PreRecoveryHooks(ctx context.Context) (b []recovery.PreHookExecutor) {
	for _, v := range m.getHooks("", m.Config().SelfServiceFlowRecoveryBeforeHooks(ctx)) {
		if hook, ok := v.(recovery.PreHookExecutor); ok {
			b = append(b, hook)
		}
	}
	return
}

func (m *RegistryDefault) PostRecoveryHooks(ctx context.Context) (b []recovery.PostHookExecutor) {
	for _, v := range m.getHooks(config.HookGlobal, m.Config().SelfServiceFlowRecoveryAfterHooks(ctx, config.HookGlobal)) {
		if hook, ok := v.(recovery.PostHookExecutor); ok {
			b = append(b, hook)
		}
	}

	return
}

func (m *RegistryDefault) RecoveryCodeSender() *code.RecoveryCodeSender {
	if m.selfserviceCodeSender == nil {
		m.selfserviceCodeSender = code.NewSender(m)
	}

	return m.selfserviceCodeSender
}
