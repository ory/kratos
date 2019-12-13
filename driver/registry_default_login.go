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

func (m *RegistryDefault) PreLoginHooks() []login.PreHookExecutor {
	return []login.PreHookExecutor{}
}

func (m *RegistryDefault) PostLoginHooks(credentialsType identity.CredentialsType) []login.PostHookExecutor {
	return m.hooksPostLogin(credentialsType, m.c.SelfServiceLoginAfterHooks(string(credentialsType)))
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
