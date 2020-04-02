package driver

import "github.com/ory/kratos/selfservice/flow/settings"

func (m *RegistryDefault) PostSettingsHooks(credentialsType string) []settings.PostHookExecutor {
	a := m.getHooks(credentialsType, m.c.SelfServiceSettingsAfterHooks(credentialsType))

	var b []settings.PostHookExecutor
	for _, v := range a {
		if hook, ok := v.(settings.PostHookExecutor); ok {
			b = append(b, hook)
		}
	}

	return b
}

func (m *RegistryDefault) SettingsExecutor() *settings.HookExecutor {
	if m.selfserviceSettingsExecutor == nil {
		m.selfserviceSettingsExecutor = settings.NewHookExecutor(m, m.c)
	}
	return m.selfserviceSettingsExecutor
}
