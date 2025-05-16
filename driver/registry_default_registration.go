// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package driver

import (
	"context"
	"slices"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/selfservice/flow/registration"
)

func (m *RegistryDefault) PostRegistrationPrePersistHooks(ctx context.Context, credentialsType identity.CredentialsType) ([]registration.PostHookPrePersistExecutor, error) {
	hooks, err := getHooks[registration.PostHookPrePersistExecutor](m, string(credentialsType), m.Config().SelfServiceFlowRegistrationAfterHooks(ctx, string(credentialsType)))
	if err != nil {
		return nil, err
	}

	return hooks, nil
}

func (m *RegistryDefault) PostRegistrationPostPersistHooks(ctx context.Context, credentialsType identity.CredentialsType) ([]registration.PostHookPostPersistExecutor, error) {
	hooks, err := getHooks[registration.PostHookPostPersistExecutor](m, string(credentialsType), m.Config().SelfServiceFlowRegistrationAfterHooks(ctx, string(credentialsType)))
	if err != nil {
		return nil, err
	}
	if len(hooks) == 0 {
		// since we don't want merging hooks defined in a specific strategy and
		// global hooks are added only if no strategy specific hooks are defined
		hooks, err = getHooks[registration.PostHookPostPersistExecutor](m, config.HookGlobal, m.Config().SelfServiceFlowRegistrationAfterHooks(ctx, config.HookGlobal))
		if err != nil {
			return nil, err
		}
	}

	// WARNING - If you remove this, no verification emails / sms will be sent post-registration.
	if m.Config().SelfServiceFlowVerificationEnabled(ctx) {
		hooks = slices.Insert(hooks, 0, registration.PostHookPostPersistExecutor(m.HookVerifier()))
	}

	return hooks, nil
}

func (m *RegistryDefault) PreRegistrationHooks(ctx context.Context) ([]registration.PreHookExecutor, error) {
	return getHooks[registration.PreHookExecutor](m, "", m.Config().SelfServiceFlowRegistrationBeforeHooks(ctx))
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
