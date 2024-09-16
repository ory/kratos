// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package text

import (
	"fmt"
	"time"
)

func NewInfoRegistration() *Message {
	return &Message{
		ID:   InfoSelfServiceRegistration,
		Text: "Sign up",
		Type: Info,
	}
}

func NewInfoRegistrationWith(provider string, providerID string) *Message {
	return &Message{
		ID:   InfoSelfServiceRegistrationWith,
		Text: fmt.Sprintf("Sign up with %s", provider),
		Type: Info,
		Context: context(map[string]any{
			"provider":    provider,
			"provider_id": providerID,
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

func NewInfoRegistrationBack() *Message {
	return &Message{
		ID:   InfoSelfServiceRegistrationBack,
		Text: "Back",
		Type: Info,
	}
}

func NewInfoSelfServiceChooseCredentials() *Message {
	return &Message{
		ID:   InfoSelfServiceRegistrationChooseCredentials,
		Text: "Please choose a credential to authenticate yourself with.",
		Type: Info,
	}
}

func NewErrorValidationRegistrationFlowExpired(expiredAt time.Time) *Message {
	return &Message{
		ID:   ErrorValidationRegistrationFlowExpired,
		Text: fmt.Sprintf("The registration flow expired %.2f minutes ago, please try again.", Since(expiredAt).Minutes()),
		Type: Error,
		Context: context(map[string]any{
			"expired_at":      expiredAt,
			"expired_at_unix": expiredAt.Unix(),
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

func NewInfoSelfServiceRegistrationRegisterPasskey() *Message {
	return &Message{
		ID:   InfoSelfServiceRegistrationRegisterPasskey,
		Text: "Sign up with passkey",
		Type: Info,
	}
}

func NewRegistrationEmailWithCodeSent() *Message {
	return &Message{
		ID:   InfoSelfServiceRegistrationEmailWithCodeSent,
		Type: Info,
		Text: "A code has been sent to the address(es) you provided. If you have not received a message, check the spelling of the address and retry the registration.",
	}
}

func NewErrorValidationRegistrationCodeInvalidOrAlreadyUsed() *Message {
	return &Message{
		ID:   ErrorValidationRegistrationCodeInvalidOrAlreadyUsed,
		Text: "The registration code is invalid or has already been used. Please try again.",
		Type: Error,
	}
}

func NewErrorValidationRegistrationRetrySuccessful() *Message {
	return &Message{
		ID:   ErrorValidateionRegistrationRetrySuccess,
		Type: Error,
		Text: "The request was already completed successfully and can not be retried.",
	}
}

func NewInfoSelfServiceRegistrationRegisterCode() *Message {
	return &Message{
		ID:   InfoSelfServiceRegistrationRegisterCode,
		Text: "Send sign up code",
		Type: Info,
	}
}
