package text

import (
	"fmt"
	"time"
)

const (
	InfoSelfServiceRegistrationRoot ID = 1040000 + iota // 1040000
	InfoSelfServiceRegistration                         // 1040001
	InfoSelfServiceRegistrationWith                     // 1040002
	InfoRegistrationContinue                            // 1040003
)

const (
	ErrorValidationRegistration ID = 4040000 + iota
	ErrorValidationRegistrationFlowExpired
)

func NewInfoRegistration() *Message {
	return &Message{
		ID:      InfoSelfServiceRegistration,
		Text:    "Sign up",
		Type:    Info,
		Context: context(map[string]interface{}{}),
	}
}

func NewInfoRegistrationWith(provider string) *Message {
	return &Message{
		ID:   InfoSelfServiceRegistrationWith,
		Text: fmt.Sprintf("Sign up with %s", provider),
		Type: Info,
		Context: context(map[string]interface{}{
			"provider": provider,
		}),
	}
}

func NewInfoRegistrationContinue() *Message {
	return &Message{
		ID:   InfoRegistrationContinue,
		Text: "Continue",
		Type: Info,
	}
}

func NewErrorValidationRegistrationFlowExpired(ago time.Duration) *Message {
	return &Message{
		ID:   ErrorValidationRegistrationFlowExpired,
		Text: fmt.Sprintf("The registration flow expired %.2f minutes ago, please try again.", ago.Minutes()),
		Type: Error,
		Context: context(map[string]interface{}{
			"expired_at": time.Now().Add(ago),
		}),
	}
}
