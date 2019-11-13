package driver

import (
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/selfservice/flow/login"
)

func (m *RegistryAbstract) LoginHookExecutor() *login.HookExecutor {
	if m.selfserviceLoginExecutor == nil {
		m.selfserviceLoginExecutor = login.NewHookExecutor(m.r, m.c)
	}
	return m.selfserviceLoginExecutor
}

func (m *RegistryAbstract) PreLoginHooks() []login.PreHookExecutor {
	return []login.PreHookExecutor{}
}

func (m *RegistryAbstract) PostLoginHooks(credentialsType identity.CredentialsType) []login.PostHookExecutor {
	a := m.hooksPost(credentialsType, m.c.SelfServiceLoginAfterHooks(string(credentialsType)))
	b := make([]login.PostHookExecutor, len(a))
	for k, v := range a {
		b[k] = v
	}
	return b
}

func (m *RegistryAbstract) LoginHandler() *login.Handler {
	if m.selfserviceLoginHandler == nil {
		m.selfserviceLoginHandler = login.NewHandler(m.r, m.c)
	}

	return m.selfserviceLoginHandler
}

func (m *RegistryAbstract) LoginRequestErrorHandler() *login.ErrorHandler {
	if m.selfserviceLoginRequestErrorHandler == nil {
		m.selfserviceLoginRequestErrorHandler = login.NewErrorHandler(m.r, m.c)
	}

	return m.selfserviceLoginRequestErrorHandler
}
