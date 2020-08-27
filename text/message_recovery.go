package text

import (
	"fmt"
	"time"

	"github.com/pkg/errors"

	"github.com/ory/herodot"
)

const (
	InfoSelfServiceRecovery           ID = 1060000 + iota // 1060000
	InfoSelfServiceRecoverySuccessful                     // 1060001
	InfoSelfServiceRecoveryEmailSent                      // 1060002
)

const (
	ErrorValidationRecovery                                  ID = 4060000 + iota // 4060000
	ErrorValidationRecoveryRetrySuccess                                          // 4060001
	ErrorValidationRecoveryStateFailure                                          // 4060002
	ErrorValidationRecoveryMissingRecoveryToken                                  // 4060003
	ErrorValidationRecoveryRecoveryTokenInvalidOrAlreadyUsed                     // 4060004
	ErrorValidationRecoveryFlowExpired                                           // 4060005
)

func NewErrorValidationRecoveryFlowExpired(ago time.Duration) *Message {
	return &Message{
		ID:   ErrorValidationRecoveryFlowExpired,
		Text: fmt.Sprintf("The recovery flow expired %.2f minutes ago, please try again.", ago.Minutes()),
		Type: Error,
		Context: context(map[string]interface{}{
			"expired_at": time.Now().Add(ago),
		}),
	}
}

func NewRecoverySuccessful(privilegedSessionExpiresAt time.Time) *Message {
	hasLeft := time.Until(privilegedSessionExpiresAt)
	return &Message{
		ID:   InfoSelfServiceRecoverySuccessful,
		Type: Info,
		Text: fmt.Sprintf("You successfully recovered your account. Please change your password or set up an alternative recovery method (e.g. social sign in) within the next %.2f minutes.", hasLeft.Minutes()),
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

func NewErrorValidationRecoveryMissingRecoveryToken() error {
	return errors.WithStack(herodot.
		ErrBadRequest.
		WithDetail("error_id", ErrorValidationRecoveryMissingRecoveryToken).
		WithReason("A recovery request was made but no recovery token was included in the request, please retry the flow."))
}

func NewErrorValidationRecoveryRecoveryTokenInvalidOrAlreadyUsed() *Message {
	return &Message{
		ID:      ErrorValidationRecoveryRecoveryTokenInvalidOrAlreadyUsed,
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
