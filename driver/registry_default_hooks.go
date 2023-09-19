// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package driver

import (
	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/selfservice/hook"
)

func (m *RegistryDefault) HookVerifier() *hook.Verifier {
	if m.hookVerifier == nil {
		m.hookVerifier = hook.NewVerifier(m)
	}
	return m.hookVerifier
}

func (m *RegistryDefault) HookCodeAddressVerifier() *hook.CodeAddressVerifier {
	if m.hookCodeAddressVerifier == nil {
		m.hookCodeAddressVerifier = hook.NewCodeAddressVerifier(m)
	}
	return m.hookCodeAddressVerifier
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

func (m *RegistryDefault) HookAddressVerifier() *hook.AddressVerifier {
	if m.hookAddressVerifier == nil {
		m.hookAddressVerifier = hook.NewAddressVerifier()
	}
	return m.hookAddressVerifier
}

func (m *RegistryDefault) HookShowVerificationUI() *hook.ShowVerificationUIHook {
	if m.hookShowVerificationUI == nil {
		m.hookShowVerificationUI = hook.NewShowVerificationUIHook(m)
	}
	return m.hookShowVerificationUI
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
		case hook.KeyWebHook:
			i = append(i, hook.NewWebHook(m, h.Config))
		case hook.KeyAddressVerifier:
			i = append(i, m.HookAddressVerifier())
		case hook.KeyVerificationUI:
			i = append(i, m.HookShowVerificationUI())
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
