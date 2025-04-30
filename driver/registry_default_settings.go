// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package driver

import (
	"context"
	"slices"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/selfservice/flow/settings"
)

func (m *RegistryDefault) PostSettingsPrePersistHooks(ctx context.Context, settingsType string) ([]settings.PostHookPrePersistExecutor, error) {
	return getHooks[settings.PostHookPrePersistExecutor](m, settingsType, m.Config().SelfServiceFlowSettingsAfterHooks(ctx, settingsType))
}

func (m *RegistryDefault) PreSettingsHooks(ctx context.Context) ([]settings.PreHookExecutor, error) {
	return getHooks[settings.PreHookExecutor](m, "", m.Config().SelfServiceFlowSettingsBeforeHooks(ctx))
}

func (m *RegistryDefault) PostSettingsPostPersistHooks(ctx context.Context, settingsType string) ([]settings.PostHookPostPersistExecutor, error) {
	hooks, err := getHooks[settings.PostHookPostPersistExecutor](m, settingsType, m.Config().SelfServiceFlowSettingsAfterHooks(ctx, settingsType))
	if err != nil {
		return nil, err
	}
	if len(hooks) == 0 {
		// since we don't want merging hooks defined in a specific strategy and
		// global hooks are added only if no strategy specific hooks are defined
		hooks, err = getHooks[settings.PostHookPostPersistExecutor](m, config.HookGlobal, m.Config().SelfServiceFlowSettingsAfterHooks(ctx, config.HookGlobal))
		if err != nil {
			return nil, err
		}
	}

	if m.Config().SelfServiceFlowVerificationEnabled(ctx) {
		hooks = slices.Insert(hooks, 0, settings.PostHookPostPersistExecutor(m.HookVerifier()))
	}

	return hooks, nil
}

func (m *RegistryDefault) SettingsHookExecutor() *settings.HookExecutor {
	if m.selfserviceSettingsExecutor == nil {
		m.selfserviceSettingsExecutor = settings.NewHookExecutor(m)
	}
	return m.selfserviceSettingsExecutor
}

func (m *RegistryDefault) SettingsHandler() *settings.Handler {
	if m.selfserviceSettingsHandler == nil {
		m.selfserviceSettingsHandler = settings.NewHandler(m)
	}
	return m.selfserviceSettingsHandler
}

func (m *RegistryDefault) SettingsFlowErrorHandler() *settings.ErrorHandler {
	if m.selfserviceSettingsErrorHandler == nil {
		m.selfserviceSettingsErrorHandler = settings.NewErrorHandler(m)
	}
	return m.selfserviceSettingsErrorHandler
}

func (m *RegistryDefault) SettingsStrategies(ctx context.Context) (profileStrategies settings.Strategies) {
	for _, strategy := range m.selfServiceStrategies() {
		if s, ok := strategy.(settings.Strategy); ok {
			if m.Config().SelfServiceStrategy(ctx, s.SettingsStrategyID()).Enabled {
				profileStrategies = append(profileStrategies, s)
			}
		}
	}
	return
}

func (m *RegistryDefault) AllSettingsStrategies() settings.Strategies {
	var profileStrategies []settings.Strategy
	for _, strategy := range m.selfServiceStrategies() {
		if s, ok := strategy.(settings.Strategy); ok {
			profileStrategies = append(profileStrategies, s)
		}
	}
	return profileStrategies
}
