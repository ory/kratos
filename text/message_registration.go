// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package text

import (
	"fmt"
	"time"
)

func NewInfoRegistration() *Message {
	return &Message{
		ID:      InfoSelfServiceRegistration,
		Text:    "Sign up",
		Type:    Info,
		Context: context(map[string]interface{}{}),
	}
}

func NewInfoRegistrationWith(provider string) *Message {
	return &Message{
		ID:   InfoSelfServiceRegistrationWith,
		Text: fmt.Sprintf("Sign up with %s", provider),
		Type: Info,
		Context: context(map[string]interface{}{
			"provider": provider,
		}),
	}
}

func NewInfoRegistrationContinue() *Message {
	return &Message{
		ID:   InfoSelfServiceRegistrationContinue,
		Text: "Continue",
		Type: Info,
	}
}

func NewErrorValidationRegistrationFlowExpired(expiredAt time.Time) *Message {
	return &Message{
		ID:   ErrorValidationRegistrationFlowExpired,
		Text: fmt.Sprintf("The registration flow expired %.2f minutes ago, please try again.", (-Until(expiredAt)).Minutes()),
		Type: Error,
		Context: context(map[string]interface{}{
			"expired_at": expiredAt,
		}),
	}
}

func NewInfoSelfServiceRegistrationRegisterWebAuthn() *Message {
	return &Message{
		ID:   InfoSelfServiceRegistrationRegisterWebAuthn,
		Text: "Sign up with security key",
		Type: Info,
	}
}
