// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package driver

import (
	"context"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/selfservice/flow/login"
)

func (m *RegistryDefault) LoginHookExecutor() *login.HookExecutor {
	if m.selfserviceLoginExecutor == nil {
		m.selfserviceLoginExecutor = login.NewHookExecutor(m)
	}
	return m.selfserviceLoginExecutor
}

func (m *RegistryDefault) PreLoginHooks(ctx context.Context) ([]login.PreHookExecutor, error) {
	return getHooks[login.PreHookExecutor](m, "", m.Config().SelfServiceFlowLoginBeforeHooks(ctx))
}

func (m *RegistryDefault) PostLoginHooks(ctx context.Context, credentialsType identity.CredentialsType) ([]login.PostHookExecutor, error) {
	hooks, err := getHooks[login.PostHookExecutor](m, string(credentialsType), m.Config().SelfServiceFlowLoginAfterHooks(ctx, string(credentialsType)))
	if err != nil {
		return nil, err
	}
	if len(hooks) > 0 {
		return hooks, nil
	}

	// since we don't want merging hooks defined in a specific strategy and global hooks
	// global hooks are added only if no strategy specific hooks are defined
	return getHooks[login.PostHookExecutor](m, config.HookGlobal, m.Config().SelfServiceFlowLoginAfterHooks(ctx, config.HookGlobal))
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
