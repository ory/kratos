package driver

import (
	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/selfservice/hook"
)

func (m *RegistryDefault) HookVerifier() *hook.Verifier {
	if m.hookVerifier == nil {
		m.hookVerifier = hook.NewVerifier(m, m.c)
	}
	return m.hookVerifier
}

func (m *RegistryDefault) HookSessionIssuer() *hook.SessionIssuer {
	if m.hookSessionIssuer == nil {
		m.hookSessionIssuer = hook.NewSessionIssuer(m)
	}
	return m.hookSessionIssuer
}

func (m *RegistryDefault) HookSessionDestroyer() *hook.SessionDestroyer {
	if m.hookSessionDestroyer == nil {
		m.hookSessionDestroyer = hook.NewSessionDestroyer(m)
	}
	return m.hookSessionDestroyer
}

func (m *RegistryDefault) WithHooks(hooks map[string]func(config.SelfServiceHook) interface{}) {
	m.injectedSelfserviceHooks = hooks
}

func (m *RegistryDefault) getHooks(credentialsType string, configs []config.SelfServiceHook) (i []interface{}) {
	for _, h := range configs {
		switch h.Name {
		case hook.KeySessionIssuer:
			i = append(i, m.HookSessionIssuer())
		case hook.KeySessionDestroyer:
			i = append(i, m.HookSessionDestroyer())
		default:
			var found bool
			for name, m := range m.injectedSelfserviceHooks {
				if name == h.Name {
					i = append(i, m(h))
					found = true
					break
				}
			}
			if found {
				continue
			}
			m.l.
				WithField("for", credentialsType).
				WithField("hook", h.Name).
				Errorf("A unknown hook was requested and can therefore not be used")
		}
	}

	return i
}
