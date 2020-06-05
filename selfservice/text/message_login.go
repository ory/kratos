package text

import (
	"fmt"
	"time"
)

const (
	InfoSelfServiceLogin ID = 1010000 + iota
)

const (
	ErrorValidationLogin ID = 4010000 + iota
	ErrorValidationLoginRequestExpired
)

func NewErrorValidationLoginRequestExpired(ago time.Duration) *Message {
	return &Message{
		ID:   ErrorValidationLoginRequestExpired,
		Text: fmt.Sprintf("The login request expired %.2f minutes ago, please try again.", ago.Minutes()),
		Type: Error,
		Context: context(map[string]interface{}{
			"expired_at": time.Now().Add(ago),
		}),
	}
}
