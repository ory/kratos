package driver

import (
	"github.com/ory/kratos/selfservice/flow/profile"
)

func (m *RegistryDefault) PostProfileManagementHooks(credentialsType string) []profile.PostHookExecutor {
	a := m.getHooks(credentialsType, m.c.SelfServiceProfileManagementAfterHooks(credentialsType))

	var b []profile.PostHookExecutor
	for _, v := range a {
		if hook, ok := v.(profile.PostHookExecutor); ok {
			b = append(b, hook)
		}
	}

	return b
}

func (m *RegistryDefault) ProfileManagementExecutor() *profile.HookExecutor {
	if m.selfserviceProfileManagementExecutor == nil {
		m.selfserviceProfileManagementExecutor = profile.NewHookExecutor(m, m.c)
	}
	return m.selfserviceProfileManagementExecutor
}
