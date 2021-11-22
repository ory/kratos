package text

import (
	"fmt"
	"time"
)

func NewErrorValidationVerificationFlowExpired(ago time.Duration) *Message {
	return &Message{
		ID:   ErrorValidationVerificationFlowExpired,
		Text: fmt.Sprintf("The verification flow expired %.2f minutes ago, please try again.", ago.Minutes()),
		Type: Error,
		Context: context(map[string]interface{}{
			"expired_at": Now().UTC().Add(ago),
		}),
	}
}

func NewInfoSelfServiceVerificationSuccessful() *Message {
	return &Message{
		ID:   InfoSelfServiceVerificationSuccessful,
		Type: Info,
		Text: "You successfully verified your email address.",
	}
}

func NewVerificationEmailSent() *Message {
	return &Message{
		ID:      InfoSelfServiceVerificationEmailSent,
		Type:    Info,
		Text:    "An email containing a verification link has been sent to the email address you provided.",
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
