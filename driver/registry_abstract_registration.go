package driver

import (
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/selfservice/flow/registration"
)

func (m *RegistryAbstract) PostRegistrationHooks(credentialsType identity.CredentialsType) []registration.PostHookExecutor {
	a := m.hooksPost(credentialsType, m.c.SelfServiceRegistrationAfterHooks(string(credentialsType)))
	b := make([]registration.PostHookExecutor, len(a))
	for k, v := range a {
		b[k] = v
	}
	return b
}

func (m *RegistryAbstract) PreRegistrationHooks() []registration.PreHookExecutor {
	return []registration.PreHookExecutor{}
}
func (m *RegistryAbstract) RegistrationExecutor() *registration.HookExecutor {
	if m.selfserviceRegistrationExecutor == nil {
		m.selfserviceRegistrationExecutor = registration.NewHookExecutor(m.r, m.c)
	}
	return m.selfserviceRegistrationExecutor
}

func (m *RegistryAbstract) RegistrationHookExecutor() *registration.HookExecutor {
	if m.selfserviceRegistrationExecutor == nil {
		m.selfserviceRegistrationExecutor = registration.NewHookExecutor(m.r, m.c)
	}
	return m.selfserviceRegistrationExecutor
}

func (m *RegistryAbstract) RegistrationErrorHandler() *registration.ErrorHandler {
	if m.seflserviceRegistrationErrorHandler == nil {
		m.seflserviceRegistrationErrorHandler = registration.NewErrorHandler(m.r, m.c)
	}
	return m.seflserviceRegistrationErrorHandler
}

func (m *RegistryAbstract) RegistrationHandler() *registration.Handler {
	if m.selfserviceRegistrationHandler == nil {
		m.selfserviceRegistrationHandler = registration.NewHandler(m.r, m.c)
	}

	return m.selfserviceRegistrationHandler
}

func (m *RegistryAbstract) RegistrationRequestErrorHandler() *registration.ErrorHandler {
	if m.selfserviceRegistrationRequestErrorHandler == nil {
		m.selfserviceRegistrationRequestErrorHandler = registration.NewErrorHandler(m.r, m.c)
	}

	return m.selfserviceRegistrationRequestErrorHandler
}
