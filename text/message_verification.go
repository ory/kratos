package text

import (
	"fmt"
	"time"
)

const (
	InfoSelfServiceVerification ID = 1070000 + iota
)

const (
	ErrorValidationVerification ID = 4070000 + iota
	ErrorValidationVerificationTokenInvalidOrAlreadyUsed
	ErrorValidationVerificationRequestExpired
)

func NewErrorValidationVerificationTokenInvalidOrAlreadyUsed() *Message {
	return &Message{
		ID:       ErrorValidationVerificationTokenInvalidOrAlreadyUsed,
		Text:     "The verification code has expired or was otherwise invalid. Please request another code.",
		Type:     Error,
		Context:  context(nil),
		I18nText: "The verification code has expired or was otherwise invalid. Please request another code.",
		I18nData: context(nil),
	}
}

func NewErrorValidationVerificationRequestExpired(ago time.Duration) *Message {
	return &Message{
		ID:   ErrorValidationVerificationRequestExpired,
		Text: fmt.Sprintf("The verification request expired %.2f minutes ago, please try again.", ago.Minutes()),
		Type: Error,
		Context: context(map[string]interface{}{
			"expired_at": time.Now().Add(ago),
		}),
		I18nText: "The verification request expired {ago} minutes ago, please try again.",
		I18nData: context(map[string]interface{}{
			"ago": ago,
		}),
	}
}
