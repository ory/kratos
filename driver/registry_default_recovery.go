package driver

import (
	"context"

	"github.com/ory/kratos/selfservice/flow/recovery"
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
			if m.Config(ctx).SelfServiceStrategy(s.RecoveryStrategyID()).Enabled {
				recoveryStrategies = append(recoveryStrategies, s)
			}
		}
	}
	return
}

func (m *RegistryDefault) AllRecoveryStrategies() (recoveryStrategies recovery.Strategies) {
	for _, strategy := range m.selfServiceStrategies() {
		if s, ok := strategy.(recovery.Strategy); ok {
			recoveryStrategies = append(recoveryStrategies, s)
		}
	}
	return
}
