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

func NewRegistrationEmailWithCodeSent() *Message {
	return &Message{
		ID:      InfoSelfServiceRegistrationEmailWithCodeSent,
		Type:    Info,
		Text:    "An email containing a code has been sent to the email address you provided. If you have not received an email, check the spelling of the address and retry the registration.",
		Context: context(nil),
	}
}

func NewErrorValidationRegistrationCodeInvalidOrAlreadyUsed() *Message {
	return &Message{
		ID:      ErrorValidationRegistrationCodeInvalidOrAlreadyUsed,
		Text:    "The registration code is invalid or has already been used. Please try again.",
		Type:    Error,
		Context: context(nil),
	}
}

func NewErrorValidationRegistrationRetrySuccessful() *Message {
	return &Message{
		ID:      ErrorValidateionRegistrationRetrySuccess,
		Type:    Error,
		Text:    "The request was already completed successfully and can not be retried.",
		Context: context(nil),
	}
}

func NewInfoSelfServiceRegistrationRegisterCode() *Message {
	return &Message{
		ID:   InfoSelfServiceRegistrationRegisterCode,
		Text: "Sign up with code",
		Type: Info,
	}
}
