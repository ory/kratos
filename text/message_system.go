// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package text

func NewErrorSystemGeneric(reason string) *Message {
	return &Message{
		ID:   ErrorSystemGeneric,
		Text: reason,
		Type: Error,
		Context: context(map[string]any{
			"reason": reason,
		}),
	}
}

func NewErrorSystemNoAuthenticationMethodsAvailable() *Message {
	return &Message{
		ID:   ErrorSystemNoAuthenticationMethodsAvailable,
		Text: "No authentication methods are available. Please contact the system administrator.",
		Type: Error,
	}
}

func NewErrorSystemOrganizationNoSSOProvidersAvailable() *Message {
	return &Message{
		ID:   ErrorSystemOrganizationNoSSOProvidersAvailable,
		Text: "Your organization requires SSO authentication, but no SSO provider is configured. Please contact the system administrator.",
		Type: Error,
	}
}

func NewCaptchaContainerMessage() *Message {
	return &Message{
		ID:   InfoNodeLabelCaptcha,
		Text: "Please complete the captcha challenge to continue.",
		Type: Info,
	}
}
