package text

import (
	"fmt"
	"time"
)

const (
	InfoSelfServiceLogin ID = 1010000 + iota
)

const (
	ErrorValidationLogin                       ID = 4010000 + iota // 4010000
	ErrorValidationLoginFlowExpired                                // 4010001
	ErrorValidationLoginNoStrategyFound                            // 4010002
	ErrorValidationRegistrationNoStrategyFound                     // 4010003
	ErrorValidationSettingsNoStrategyFound                         // 4010004
)

func NewErrorValidationLoginFlowExpired(ago time.Duration) *Message {
	return &Message{
		ID:   ErrorValidationLoginFlowExpired,
		Text: fmt.Sprintf("The login flow expired %.2f minutes ago, please try again.", ago.Minutes()),
		Type: Error,
		Context: context(map[string]interface{}{
			"expired_at": time.Now().Add(ago),
		}),
	}
}

func NewErrorValidationLoginNoStrategyFound() *Message {
	return &Message{
		ID:   ErrorValidationLoginNoStrategyFound,
		Text: "Could not find a strategy to log you in with. Did you fill out the form correctly?",
		Type: Error,
	}
}

func NewErrorValidationRegistrationNoStrategyFound() *Message {
	return &Message{
		ID:   ErrorValidationRegistrationNoStrategyFound,
		Text: "Could not find a strategy to sign you up with. Did you fill out the form correctly?",
		Type: Error,
	}
}

func NewErrorValidationSettingsNoStrategyFound() *Message {
	return &Message{
		ID:   ErrorValidationSettingsNoStrategyFound,
		Text: "Could not find a strategy to update your settings. Did you fill out the form correctly?",
		Type: Error,
	}
}
