// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package driver

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/identity"
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

func (m *RegistryDefault) WithHooks(hooks map[string]func(config.SelfServiceHook) interface{}) {
	m.injectedSelfserviceHooks = hooks
}

func (m *RegistryDefault) WithExtraHandlers(handlers []NewHandlerRegistrar) {
	m.extraHandlerFactories = handlers
}

func isEnumerationSafeType(credentialsType string) bool {
	switch credentialsType {
	case identity.CredentialsTypeOIDC.String(),
		identity.CredentialsTypeCodeAuth.String(),
		config.HookGlobal:
		return true
	default:
		return false
	}
}

func getHooks[T any](m *RegistryDefault, ctx context.Context, credentialsType string, configs []config.SelfServiceHook) ([]T, error) {
	hooks := make([]T, 0, len(configs))

	var addSessionIssuer bool
allHooksLoop:
	for _, h := range configs {
		switch h.Name {
		case hook.KeySessionIssuer:

			if m.Config().SecurityAccountEnumerationMitigate(ctx) && !isEnumerationSafeType(credentialsType) {
				m.l.WithField("for", credentialsType).Error("The 'session' hook is incompatible with account enumeration mitigation")
				return nil, errors.Errorf(
					"the 'session' hook for %s is incompatible with security.account_enumeration.mitigate=true: "+
						"issuing sessions during anti-enumeration flows would leak information about existing accounts. "+
						"Please remove the 'session' hook from selfservice.flows.registration.after.%s.hooks or disable account enumeration mitigation",
					credentialsType, credentialsType)
			}
			// The session issuer hook always needs to come last.
			addSessionIssuer = true
		case hook.KeySessionDestroyer:
			if h, ok := any(m.HookSessionDestroyer()).(T); ok {
				hooks = append(hooks, h)
			}
		case hook.KeyWebHook:
			cfg := request.Config{}
			if err := json.Unmarshal(h.Config, &cfg); err != nil {
				m.l.WithError(err).WithField("raw_config", string(h.Config)).Error("failed to unmarshal hook configuration, ignoring hook")
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
		default:
			for name, m := range m.injectedSelfserviceHooks {
				if name == h.Name {
					if h, ok := m(h).(T); ok {
						hooks = append(hooks, h)
					}
					continue allHooksLoop
				}
			}
			m.l.
				WithField("for", credentialsType).
				WithField("hook", h.Name).
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
