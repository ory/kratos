package text

import (
	"fmt"
	"time"
)

func NewErrorValidationRecoveryFlowExpired(ago time.Duration) *Message {
	return &Message{
		ID:   ErrorValidationRecoveryFlowExpired,
		Text: fmt.Sprintf("The recovery flow expired %.2f minutes ago, please try again.", ago.Minutes()),
		Type: Error,
		Context: context(map[string]interface{}{
			"expired_at": Now().UTC().Add(ago),
		}),
	}
}

func NewRecoverySuccessful(privilegedSessionExpiresAt time.Time) *Message {
	hasLeft := Until(privilegedSessionExpiresAt)
	return &Message{
		ID:   InfoSelfServiceRecoverySuccessful,
		Type: Info,
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
		Text:    "An email containing a recovery link has been sent to the email address you provided.",
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
