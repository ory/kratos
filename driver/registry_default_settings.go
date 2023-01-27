// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package driver

import (
	"context"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/selfservice/flow/settings"
)

func (m *RegistryDefault) PostSettingsPrePersistHooks(ctx context.Context, settingsType string) (b []settings.PostHookPrePersistExecutor) {
	for _, v := range m.getHooks(settingsType, m.Config().SelfServiceFlowSettingsAfterHooks(ctx, settingsType)) {
		if hook, ok := v.(settings.PostHookPrePersistExecutor); ok {
			b = append(b, hook)
		}
	}
	return
}

func (m *RegistryDefault) PreSettingsHooks(ctx context.Context) (b []settings.PreHookExecutor) {
	for _, v := range m.getHooks("", m.Config().SelfServiceFlowSettingsBeforeHooks(ctx)) {
		if hook, ok := v.(settings.PreHookExecutor); ok {
			b = append(b, hook)
		}
	}
	return
}

func (m *RegistryDefault) PostSettingsPostPersistHooks(ctx context.Context, settingsType string) (b []settings.PostHookPostPersistExecutor) {
	initialHookCount := 0
	if m.Config().SelfServiceFlowVerificationEnabled(ctx) {
		b = append(b, m.HookVerifier())
		initialHookCount = 1
	}

	for _, v := range m.getHooks(settingsType, m.Config().SelfServiceFlowSettingsAfterHooks(ctx, settingsType)) {
		if hook, ok := v.(settings.PostHookPostPersistExecutor); ok {
			b = append(b, hook)
		}
	}

	if len(b) == initialHookCount {
		// since we don't want merging hooks defined in a specific strategy and global hooks
		// global hooks are added only if no strategy specific hooks are defined
		for _, v := range m.getHooks(config.HookGlobal, m.Config().SelfServiceFlowSettingsAfterHooks(ctx, config.HookGlobal)) {
			if hook, ok := v.(settings.PostHookPostPersistExecutor); ok {
				b = append(b, hook)
			}
		}
	}

	return
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
