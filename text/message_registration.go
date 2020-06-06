package text

import (
	"fmt"
	"time"
)

const (
	InfoSelfServiceRegistration ID = 1040000 + iota
)

const (
	ErrorValidationRegistration ID = 4040000 + iota
	ErrorValidationRegistrationRequestExpired
)

func NewErrorValidationRegistrationRequestExpired(ago time.Duration) *Message {
	return &Message{
		ID:   ErrorValidationRegistrationRequestExpired,
		Text: fmt.Sprintf("The registration request expired %.2f minutes ago, please try again.", ago.Minutes()),
		Type: Error,
		Context: context(map[string]interface{}{
			"expired_at": time.Now().Add(ago),
		}),
	}
}
