// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package text

import (
	"fmt"
	"time"
)

func NewErrorValidationRecoveryFlowExpired(expiredAt time.Time) *Message {
	return &Message{
		ID:   ErrorValidationRecoveryFlowExpired,
		Text: fmt.Sprintf("The recovery flow expired %.2f minutes ago, please try again.", (-Until(expiredAt)).Minutes()),
		Type: Error,
		Context: context(map[string]interface{}{
			"expired_at": expiredAt,
		}),
	}
}

func NewRecoverySuccessful(privilegedSessionExpiresAt time.Time) *Message {
	hasLeft := Until(privilegedSessionExpiresAt)
	return &Message{
		ID:   InfoSelfServiceRecoverySuccessful,
		Type: Success,
		Text: fmt.Sprintf("You successfully recovered your account. Please change your password or set up an alternative login method (e.g. social sign in) within the next %.2f minutes.", hasLeft.Minutes()),
		Context: context(map[string]interface{}{
			"privilegedSessionExpiresAt": privilegedSessionExpiresAt,
		}),
	}
}

func NewRecoveryEmailSent() *Message {
	return &Message{
		ID:      InfoSelfServiceRecoveryEmailSent,
		Type:    Info,
		Text:    "An email containing a recovery link has been sent to the email address you provided. If you have not received an email, check the spelling of the address and make sure to use the address you registered with.",
		Context: context(nil),
	}
}

func NewRecoveryEmailWithCodeSent() *Message {
	return &Message{
		ID:      InfoSelfServiceRecoveryEmailWithCodeSent,
		Type:    Info,
		Text:    "An email containing a recovery code has been sent to the email address you provided. If you have not received an email, check the spelling of the address and make sure to use the address you registered with.",
		Context: context(nil),
	}
}

func NewErrorValidationRecoveryTokenInvalidOrAlreadyUsed() *Message {
	return &Message{
		ID:      ErrorValidationRecoveryTokenInvalidOrAlreadyUsed,
		Text:    "The recovery token is invalid or has already been used. Please retry the flow.",
		Type:    Error,
		Context: context(nil),
	}
}

func NewErrorValidationRecoveryCodeInvalidOrAlreadyUsed() *Message {
	return &Message{
		ID:      ErrorValidationRecoveryCodeInvalidOrAlreadyUsed,
		Text:    "The recovery code is invalid or has already been used. Please try again.",
		Type:    Error,
		Context: context(nil),
	}
}

func NewErrorValidationRecoveryRetrySuccess() *Message {
	return &Message{
		ID:      ErrorValidationRecoveryRetrySuccess,
		Text:    "The request was already completed successfully and can not be retried.",
		Type:    Error,
		Context: context(nil),
	}
}

func NewErrorValidationRecoveryStateFailure() *Message {
	return &Message{
		ID:      ErrorValidationRecoveryStateFailure,
		Text:    "The recovery flow reached a failure state and must be retried.",
		Type:    Error,
		Context: context(nil),
	}
}
