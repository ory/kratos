package text

import (
	"fmt"
	"time"
)

const (
	InfoSelfServiceLoginRoot ID = 1010000 + iota // 1010000
	InfoSelfServiceLogin                         // 1010001
	InfoSelfServiceLoginWith                     // 1010002
)

const (
	ErrorValidationLogin                       ID = 4010000 + iota // 4010000
	ErrorValidationLoginFlowExpired                                // 4010001
	ErrorValidationLoginNoStrategyFound                            // 4010002
	ErrorValidationRegistrationNoStrategyFound                     // 4010003
	ErrorValidationSettingsNoStrategyFound                         // 4010004
	ErrorValidationRecoveryNoStrategyFound                         // 4010005
	ErrorValidationVerificationNoStrategyFound                     // 4010006
)

func NewInfoLogin() *Message {
	return &Message{
		ID:      InfoSelfServiceLogin,
		Text:    "Sign in",
		Type:    Info,
		Context: context(map[string]interface{}{}),
	}
}

func NewInfoLoginWith(provider string) *Message {
	return &Message{
		ID:   InfoSelfServiceLoginWith,
		Text: fmt.Sprintf("Sign in with %s", provider),
		Type: Info,
		Context: context(map[string]interface{}{
			"provider": provider,
		}),
	}
}

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

func NewErrorValidationRecoveryNoStrategyFound() *Message {
	return &Message{
		ID:   ErrorValidationRecoveryNoStrategyFound,
		Text: "Could not find a strategy to recover your account with. Did you fill out the form correctly?",
		Type: Error,
	}
}

func NewErrorValidationVerificationNoStrategyFound() *Message {
	return &Message{
		ID:   ErrorValidationVerificationNoStrategyFound,
		Text: "Could not find a strategy to verify your account with. Did you fill out the form correctly?",
		Type: Error,
	}
}
