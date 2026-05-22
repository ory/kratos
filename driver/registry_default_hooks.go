// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package driver

import (
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/request"
	"github.com/ory/kratos/selfservice/hook"
)

func (m *RegistryDefault) HookVerifier() *hook.Verifier {
	if m.hookVerifier == nil {
		m.hookVerifier = hook.NewVerifier(m)
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

func (m *RegistryDefault) HookAddressVerifier() *hook.AddressVerifier {
	if m.hookAddressVerifier == nil {
		m.hookAddressVerifier = hook.NewAddressVerifier(m)
	}
	return m.hookAddressVerifier
}

func (m *RegistryDefault) HookShowVerificationUI() *hook.ShowVerificationUIHook {
	if m.hookShowVerificationUI == nil {
		m.hookShowVerificationUI = hook.NewShowVerificationUIHook(m)
	}
	return m.hookShowVerificationUI
}

func (m *RegistryDefault) HookVerifyNewAddress() *hook.VerifyNewAddress {
	if m.hookVerifyNewAddress == nil {
		m.hookVerifyNewAddress = hook.NewVerifyNewAddress(m)
	}
	return m.hookVerifyNewAddress
}

func (m *RegistryDefault) HookNotifyPreviousAddresses(c *hook.NotifyPreviousAddressesConfig) *hook.NotifyPreviousAddresses {
	return hook.NewNotifyPreviousAddresses(m, c)
}

func (m *RegistryDefault) WithHooks(hooks map[string]NewHookFn) {
	m.injectedSelfserviceHooks = hooks
}
func (m *RegistryDefault) WithExtraHandlers(handlers []NewHandler) {
	m.extraHandlerFactories = handlers
}

func getHooks[T any](m *RegistryDefault, credentialsType string, configs []config.SelfServiceHook) ([]T, error) {
	hooks := make([]T, 0, len(configs))

	var addSessionIssuer bool
allHooksLoop:
	for _, hookConfig := range configs {
		switch hookConfig.Name {
		case hook.KeySessionIssuer:
			// The session issuer hook always needs to come last.
			addSessionIssuer = true
		case hook.KeySessionDestroyer:
			if h, ok := any(m.HookSessionDestroyer()).(T); ok {
				hooks = append(hooks, h)
			}
		case hook.KeyWebHook:
			cfg := request.Config{}
			if err := json.Unmarshal(hookConfig.Config, &cfg); err != nil {
				m.l.WithError(err).WithField("raw_config", string(hookConfig.Config)).Error("failed to unmarshal hook configuration, ignoring hook")
				return nil, errors.WithStack(fmt.Errorf("failed to unmarshal webhook configuration for %s: %w", credentialsType, err))
			}
			if h, ok := any(hook.NewWebHook(m, &cfg)).(T); ok {
				hooks = append(hooks, h)
			}
		case hook.KeyRequireVerifiedAddress:
			if h, ok := any(m.HookAddressVerifier()).(T); ok {
				hooks = append(hooks, h)
			}
		case hook.KeyVerificationUI:
			if h, ok := any(m.HookShowVerificationUI()).(T); ok {
				hooks = append(hooks, h)
			}
		case hook.KeyVerifier:
			if h, ok := any(m.HookVerifier()).(T); ok {
				hooks = append(hooks, h)
			}
		case hook.KeyVerifyNewAddress:
			if h, ok := any(m.HookVerifyNewAddress()).(T); ok {
				hooks = append(hooks, h)
			}
		case hook.KeyNotifyPreviousAddresses:
			cfg := &hook.NotifyPreviousAddressesConfig{}
			if len(hookConfig.Config) > 0 {
				if err := json.Unmarshal(hookConfig.Config, cfg); err != nil {
					m.l.WithError(err).WithField("raw_config", string(hookConfig.Config)).Error("failed to unmarshal hook configuration, ignoring hook")
					return nil, errors.WithStack(fmt.Errorf("failed to unmarshal notify_previous_addresses configuration for %s: %w", credentialsType, err))
				}
			}
			if h, ok := any(m.HookNotifyPreviousAddresses(cfg)).(T); ok {
				hooks = append(hooks, h)
			}
		default:
			for name, newHook := range m.injectedSelfserviceHooks {
				if name == hookConfig.Name {
					if h, ok := newHook(hookConfig, m).(T); ok {
						hooks = append(hooks, h)
					}
					continue allHooksLoop
				}
			}
			m.l.
				WithField("for", credentialsType).
				WithField("hook", hookConfig.Name).
				Warn("A configuration for a non-existing hook was found and will be ignored.")
		}
	}
	if addSessionIssuer {
		if h, ok := any(m.HookSessionIssuer()).(T); ok {
			hooks = append(hooks, h)
		}
	}

	return hooks, nil
}
