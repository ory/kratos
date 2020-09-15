package text

import (
	"fmt"
	"time"
)

const (
	InfoSelfServiceVerification           ID = 1070000 + iota
	InfoSelfServiceVerificationSuccessful    // 1060001
	InfoSelfServiceVerificationEmailSent     // 1060002
)

const (
	ErrorValidationVerification ID = 4070000 + iota
	ErrorValidationVerificationTokenInvalidOrAlreadyUsed
	ErrorValidationVerificationRetrySuccess
	ErrorValidationVerificationStateFailure
	ErrorValidationVerificationMissingVerificationToken
	ErrorValidationVerificationFlowExpired
)

func NewErrorValidationVerificationFlowExpired(ago time.Duration) *Message {
	return &Message{
		ID:   ErrorValidationVerificationFlowExpired,
		Text: fmt.Sprintf("The verification flow expired %.2f minutes ago, please try again.", ago.Minutes()),
		Type: Error,
		Context: context(map[string]interface{}{
			"expired_at": time.Now().Add(ago),
		}),
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
