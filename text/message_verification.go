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
		Text: fmt.Sprintf("The verification flow expired %.2f minutes ago, please try again.", (-Until(expiredAt)).Minutes()),
		Type: Error,
		Context: context(map[string]interface{}{
			"expired_at": expiredAt,
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

func NewVerificationEmailSent() *Message {
	return &Message{
		ID:      InfoSelfServiceVerificationEmailSent,
		Type:    Info,
		Text:    "An email containing a verification link has been sent to the email address you provided. If you have not received an email, check the spelling of the address and make sure to use the address you registered with.",
		Context: context(nil),
	}
}

func NewErrorValidationVerificationTokenInvalidOrAlreadyUsed() *Message {
	return &Message{
		ID:      ErrorValidationVerificationTokenInvalidOrAlreadyUsed,
		Text:    "The verification token is invalid or has already been used. Please retry the flow.",
		Type:    Error,
		Context: context(nil),
	}
}

func NewErrorValidationVerificationRetrySuccess() *Message {
	return &Message{
		ID:      ErrorValidationVerificationRetrySuccess,
		Text:    "The request was already completed successfully and can not be retried.",
		Type:    Error,
		Context: context(nil),
	}
}

func NewErrorValidationVerificationStateFailure() *Message {
	return &Message{
		ID:      ErrorValidationVerificationStateFailure,
		Text:    "The verification flow reached a failure state and must be retried.",
		Type:    Error,
		Context: context(nil),
	}
}

func NewErrorValidationVerificationCodeInvalidOrAlreadyUsed() *Message {
	return &Message{
		ID:      ErrorValidationVerificationCodeInvalidOrAlreadyUsed,
		Text:    "The verification code is invalid or has already been used. Please try again.",
		Type:    Error,
		Context: context(nil),
	}
}

func NewVerificationEmailWithCodeSent() *Message {
	return &Message{
		ID:      InfoSelfServiceVerificationEmailWithCodeSent,
		Type:    Info,
		Text:    "An email containing a verification code has been sent to the email address you provided. If you have not received an email, check the spelling of the address and make sure to use the address you registered with.",
		Context: context(nil),
	}
}
