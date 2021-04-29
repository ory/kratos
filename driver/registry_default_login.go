package driver

import (
	"context"

	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/selfservice/flow/login"
)

func (m *RegistryDefault) LoginHookExecutor() *login.HookExecutor {
	if m.selfserviceLoginExecutor == nil {
		m.selfserviceLoginExecutor = login.NewHookExecutor(m)
	}
	return m.selfserviceLoginExecutor
}

func (m *RegistryDefault) PreLoginHooks(ctx context.Context) (b []login.PreHookExecutor) {
	for _, v := range m.getHooks("", m.Config(ctx).SelfServiceFlowLoginBeforeHooks()) {
		if hook, ok := v.(login.PreHookExecutor); ok {
			b = append(b, hook)
		}
	}
	return
}

func (m *RegistryDefault) PostLoginHooks(ctx context.Context, credentialsType identity.CredentialsType) (b []login.PostHookExecutor) {
	for _, v := range m.getHooks(string(credentialsType), m.Config(ctx).SelfServiceFlowLoginAfterHooks(string(credentialsType))) {
		if hook, ok := v.(login.PostHookExecutor); ok {
			b = append(b, hook)
		}
	}

	if len(b) == 0 {
		// since we don't want merging hooks defined in a specific strategy and global hooks
		// global hooks are added only if no strategy specific hooks are defined
		for _, v := range m.getHooks("global", m.Config(ctx).SelfServiceFlowLoginAfterHooks("global")) {
			if hook, ok := v.(login.PostHookExecutor); ok {
				b = append(b, hook)
			}
		}
	}
	return
}

func (m *RegistryDefault) LoginHandler() *login.Handler {
	if m.selfserviceLoginHandler == nil {
		m.selfserviceLoginHandler = login.NewHandler(m)
	}

	return m.selfserviceLoginHandler
}

func (m *RegistryDefault) LoginFlowErrorHandler() *login.ErrorHandler {
	if m.selfserviceLoginRequestErrorHandler == nil {
		m.selfserviceLoginRequestErrorHandler = login.NewFlowErrorHandler(m)
	}

	return m.selfserviceLoginRequestErrorHandler
}
