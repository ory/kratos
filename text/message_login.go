package text

import (
	"fmt"
	"time"
)

const (
	InfoSelfServiceLogin ID = 1010000 + iota
)

const (
	ErrorValidationLogin               ID = 4010000 + iota // 4010000
	ErrorValidationLoginRequestExpired                     // 4010001
)

func NewErrorValidationLoginRequestExpired(ago time.Duration) *Message {
	return &Message{
		ID:   ErrorValidationLoginRequestExpired,
		Text: fmt.Sprintf("The login request expired %.2f minutes ago, please try again.", ago.Minutes()),
		Type: Error,
		Context: context(map[string]interface{}{
			"expired_at": time.Now().Add(ago),
		}),
		I18nText: "The login request expired {ago} minutes ago, please try again.",
		I18nData: context(map[string]interface{}{
			"ago": ago.Minutes(),
		}),
	}
}
