// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package text

import (
	"fmt"
	"time"
)

func NewErrorValidationVerificationFlowExpired(expiredAt time.Time) *Message {
	return &Message{
		ID:   ErrorValidationVerificationFlowExpired,
		Text: fmt.Sprintf("The verification flow expired %.2f minutes ago, please try again.", Since(expiredAt).Minutes()),
		Type: Error,
		Context: context(map[string]any{
			"expired_at":      expiredAt,
			"expired_at_unix": expiredAt.Unix(),
		}),
	}
}

func NewInfoSelfServiceVerificationSuccessful() *Message {
	return &Message{
		ID:   InfoSelfServiceVerificationSuccessful,
		Type: Success,
		Text: "You successfully verified your email address.",
	}
}

func NewInfoSelfServiceVerificationPhoneSuccessful() *Message {
	return &Message{
		ID:   InfoSelfServiceVerificationPhoneSuccessful,
		Type: Success,
		Text: "You successfully verified your phone number.",
	}
}

func NewVerificationEmailSent() *Message {
	return &Message{
		ID:   InfoSelfServiceVerificationEmailSent,
		Type: Info,
		Text: "An email containing a verification link has been sent to the email address you provided. If you have not received an email, check the spelling of the address and make sure to use the address you registered with.",
	}
}

func NewErrorValidationVerificationTokenInvalidOrAlreadyUsed() *Message {
	return &Message{
		ID:   ErrorValidationVerificationTokenInvalidOrAlreadyUsed,
		Text: "The verification token is invalid or has already been used. Please retry the flow.",
		Type: Error,
	}
}

func NewErrorValidationVerificationRetrySuccess() *Message {
	return &Message{
		ID:   ErrorValidationVerificationRetrySuccess,
		Text: "The request was already completed successfully and can not be retried.",
		Type: Error,
	}
}

func NewErrorValidationVerificationStateFailure() *Message {
	return &Message{
		ID:   ErrorValidationVerificationStateFailure,
		Text: "The verification flow reached a failure state and must be retried.",
		Type: Error,
	}
}

func NewErrorValidationVerificationCodeInvalidOrAlreadyUsed() *Message {
	return &Message{
		ID:   ErrorValidationVerificationCodeInvalidOrAlreadyUsed,
		Text: "The verification code is invalid or has already been used. Please try again.",
		Type: Error,
	}
}

func NewVerificationEmailWithCodeSent() *Message {
	return &Message{
		ID:   InfoSelfServiceVerificationEmailWithCodeSent,
		Type: Info,
		Text: "An email containing a verification code has been sent to the email address you provided. If you have not received an email, check the spelling of the address and make sure to use the address you registered with.",
	}
}

func NewVerificationPhoneWithCodeSent() *Message {
	return &Message{
		ID:   InfoSelfServiceVerificationPhoneWithCodeSent,
		Type: Info,
		Text: "A text message containing a verification code has been sent to the phone number you provided. If you have not received a text message, check the spelling of the number and make sure to use the number you registered with.",
	}
}
