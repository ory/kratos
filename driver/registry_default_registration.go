package driver

import (
	"context"

	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/selfservice/flow/registration"
)

func (m *RegistryDefault) PostRegistrationPrePersistHooks(ctx context.Context, credentialsType identity.CredentialsType) (b []registration.PostHookPrePersistExecutor) {
	for _, v := range m.getHooks(string(credentialsType), m.Config(ctx).SelfServiceFlowRegistrationAfterHooks(string(credentialsType))) {
		if hook, ok := v.(registration.PostHookPrePersistExecutor); ok {
			b = append(b, hook)
		}
	}

	return
}

func (m *RegistryDefault) PostRegistrationPostPersistHooks(ctx context.Context, credentialsType identity.CredentialsType) (b []registration.PostHookPostPersistExecutor) {
	if m.Config(ctx).SelfServiceFlowVerificationEnabled() {
		b = append(b, m.HookVerifier())
	}

	for _, v := range m.getHooks(string(credentialsType), m.Config(ctx).SelfServiceFlowRegistrationAfterHooks(string(credentialsType))) {
		if hook, ok := v.(registration.PostHookPostPersistExecutor); ok {
			b = append(b, hook)
		}
	}

	for _, v := range m.getHooks("none", m.Config(ctx).SelfServiceFlowRegistrationAfterWebHooks()) {
		if hook, ok := v.(registration.PostHookPostPersistExecutor); ok {
			b = append(b, hook)
		}
	}

	return
}

func (m *RegistryDefault) PreRegistrationHooks(ctx context.Context) (b []registration.PreHookExecutor) {
	for _, v := range m.getHooks("", m.Config(ctx).SelfServiceFlowRegistrationBeforeHooks()) {
		if hook, ok := v.(registration.PreHookExecutor); ok {
			b = append(b, hook)
		}
	}
	return
}

func (m *RegistryDefault) RegistrationExecutor() *registration.HookExecutor {
	if m.selfserviceRegistrationExecutor == nil {
		m.selfserviceRegistrationExecutor = registration.NewHookExecutor(m)
	}
	return m.selfserviceRegistrationExecutor
}

func (m *RegistryDefault) RegistrationHookExecutor() *registration.HookExecutor {
	if m.selfserviceRegistrationExecutor == nil {
		m.selfserviceRegistrationExecutor = registration.NewHookExecutor(m)
	}
	return m.selfserviceRegistrationExecutor
}

func (m *RegistryDefault) RegistrationErrorHandler() *registration.ErrorHandler {
	if m.seflserviceRegistrationErrorHandler == nil {
		m.seflserviceRegistrationErrorHandler = registration.NewErrorHandler(m)
	}
	return m.seflserviceRegistrationErrorHandler
}

func (m *RegistryDefault) RegistrationHandler() *registration.Handler {
	if m.selfserviceRegistrationHandler == nil {
		m.selfserviceRegistrationHandler = registration.NewHandler(m)
	}

	return m.selfserviceRegistrationHandler
}

func (m *RegistryDefault) RegistrationFlowErrorHandler() *registration.ErrorHandler {
	if m.selfserviceRegistrationRequestErrorHandler == nil {
		m.selfserviceRegistrationRequestErrorHandler = registration.NewErrorHandler(m)
	}

	return m.selfserviceRegistrationRequestErrorHandler
}
