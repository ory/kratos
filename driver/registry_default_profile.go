package driver

import "github.com/ory/kratos/selfservice/flow/settings"

func (m *RegistryDefault) PostSettingsPrePersistHooks(settingsType string) (b []settings.PostHookPrePersistExecutor) {
	for _, v := range m.getHooks(settingsType, m.c.SelfServiceSettingsAfterHooks(settingsType)) {
		if hook, ok := v.(settings.PostHookPrePersistExecutor); ok {
			b = append(b, hook)
		}
	}
	return
}
func (m *RegistryDefault) PostSettingsPostPersistHooks(settingsType string) (b []settings.PostHookPostPersistExecutor) {
	for _, v := range m.getHooks(settingsType, m.c.SelfServiceSettingsAfterHooks(settingsType)) {
		if hook, ok := v.(settings.PostHookPostPersistExecutor); ok {
			b = append(b, hook)
		}
	}
	return
}

func (m *RegistryDefault) SettingsHookExecutor() *settings.HookExecutor {
	if m.selfserviceSettingsExecutor == nil {
		m.selfserviceSettingsExecutor = settings.NewHookExecutor(m, m.c)
	}
	return m.selfserviceSettingsExecutor
}
