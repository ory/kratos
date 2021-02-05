package driver

import (
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

func (m *RegistryDefault) RecoveryStrategies() recovery.Strategies {
	if len(m.recoveryStrategies) == 0 {
		for _, strategy := range m.selfServiceStrategies() {
			if s, ok := strategy.(recovery.Strategy); ok {
				if m.c.SelfServiceStrategy(s.RecoveryStrategyID()).Enabled {
					m.recoveryStrategies = append(m.recoveryStrategies, s)
				}
			}
		}
	}
	return m.recoveryStrategies
}

func (m *RegistryDefault) AllRecoveryStrategies() recovery.Strategies {
	var recoveryStrategies []recovery.Strategy
	for _, strategy := range m.selfServiceStrategies() {
		if s, ok := strategy.(recovery.Strategy); ok {
			recoveryStrategies = append(recoveryStrategies, s)
		}
	}
	return recoveryStrategies
}
