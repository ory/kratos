package driver

import (
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/selfservice/flow/login"
)

func (m *RegistryDefault) LoginHookExecutor() *login.HookExecutor {
	if m.selfserviceLoginExecutor == nil {
		m.selfserviceLoginExecutor = login.NewHookExecutor(m, m.c)
	}
	return m.selfserviceLoginExecutor
}

func (m *RegistryDefault) PreLoginHooks() (b []login.PreHookExecutor) {
	for _, v := range m.getHooks("", m.c.SelfServiceFlowLoginBeforeHooks()) {
		if hook, ok := v.(login.PreHookExecutor); ok {
			b = append(b, hook)
		}
	}
	return
}

func (m *RegistryDefault) PostLoginHooks(credentialsType identity.CredentialsType) (b []login.PostHookExecutor) {
	for _, v := range m.getHooks(string(credentialsType), m.c.SelfServiceFlowLoginAfterHooks(string(credentialsType))) {
		if hook, ok := v.(login.PostHookExecutor); ok {
			b = append(b, hook)
		}
	}
	return
}

func (m *RegistryDefault) LoginHandler() *login.Handler {
	if m.selfserviceLoginHandler == nil {
		m.selfserviceLoginHandler = login.NewHandler(m, m.c)
	}

	return m.selfserviceLoginHandler
}

func (m *RegistryDefault) LoginRequestErrorHandler() *login.ErrorHandler {
	if m.selfserviceLoginRequestErrorHandler == nil {
		m.selfserviceLoginRequestErrorHandler = login.NewErrorHandler(m, m.c)
	}

	return m.selfserviceLoginRequestErrorHandler
}
