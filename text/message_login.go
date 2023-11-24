// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package text

import (
	"fmt"
	"time"
)

func NewInfoLoginReAuth() *Message {
	return &Message{
		ID:   InfoSelfServiceLoginReAuth,
		Type: Info,
		Text: "Please confirm this action by verifying that it is you.",
	}
}

func NewInfoLoginMFA() *Message {
	return &Message{
		ID:   InfoSelfServiceLoginMFA,
		Type: Info,
		Text: "Please complete the second authentication challenge.",
	}
}

func NewInfoLoginWebAuthnPasswordless() *Message {
	return &Message{
		ID:   InfoSelfServiceLoginWebAuthnPasswordless,
		Type: Info,
		Text: "Prepare your WebAuthn device (e.g. security key, biometrics scanner, ...) and press continue.",
	}
}

func NewInfoLoginTOTPLabel() *Message {
	return &Message{
		ID:   InfoSelfServiceLoginTOTPLabel,
		Type: Info,
		Text: "Authentication code",
	}
}

func NewInfoLoginLookupLabel() *Message {
	return &Message{
		ID:   InfoLoginLookupLabel,
		Type: Info,
		Text: "Backup recovery code",
	}
}

func NewInfoLogin() *Message {
	return &Message{
		ID:   InfoSelfServiceLogin,
		Text: "Sign in",
		Type: Info,
	}
}

func NewInfoLoginLinkMessage(dupIdentifier, provider, newLoginURL string) *Message {
	return &Message{
		ID:   InfoSelfServiceLoginLink,
		Type: Info,
		Text: fmt.Sprintf(
			"Signing in will link your account to %q at provider %q. If you do not wish to link that account, please start a new login flow.",
			dupIdentifier,
			provider,
		),
		Context: context(map[string]any{
			"duplicateIdentifier": dupIdentifier,
			"provider":            provider,
			"newLoginUrl":         newLoginURL,
		}),
	}
}

func NewInfoLoginAndLink() *Message {
	return &Message{
		ID:   InfoSelfServiceLoginAndLink,
		Text: "Sign in and link",
		Type: Info,
	}
}

func NewInfoLoginTOTP() *Message {
	return &Message{
		ID:   InfoLoginTOTP,
		Text: "Use Authenticator",
		Type: Info,
	}
}

func NewInfoLoginLookup() *Message {
	return &Message{
		ID:   InfoLoginLookup,
		Text: "Use backup recovery code",
		Type: Info,
	}
}

func NewInfoLoginVerify() *Message {
	return &Message{
		ID:   InfoSelfServiceLoginVerify,
		Text: "Verify",
		Type: Info,
	}
}

func NewInfoLoginWith(provider string) *Message {
	return &Message{
		ID:   InfoSelfServiceLoginWith,
		Text: fmt.Sprintf("Sign in with %s", provider),
		Type: Info,
		Context: context(map[string]any{
			"provider": provider,
		}),
	}
}

func NewInfoLoginWithAndLink(provider string) *Message {

	return &Message{
		ID:   InfoSelfServiceLoginWithAndLink,
		Text: fmt.Sprintf("Sign in with %s and link credential", provider),
		Type: Info,
		Context: context(map[string]any{
			"provider": provider,
		}),
	}
}

func NewErrorValidationLoginFlowExpired(expiredAt time.Time) *Message {
	return &Message{
		ID:   ErrorValidationLoginFlowExpired,
		Text: fmt.Sprintf("The login flow expired %.2f minutes ago, please try again.", Since(expiredAt).Minutes()),
		Type: Error,
		Context: context(map[string]any{
			"expired_at":      expiredAt,
			"expired_at_unix": expiredAt.Unix(),
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

func NewInfoSelfServiceLoginWebAuthn() *Message {
	return &Message{
		ID:   InfoSelfServiceLoginWebAuthn,
		Text: "Use security key",
		Type: Info,
	}
}

func NewInfoSelfServiceContinueLoginWebAuthn() *Message {
	return &Message{
		ID:   InfoSelfServiceLoginContinueWebAuthn,
		Text: "Continue with security key",
		Type: Info,
	}
}

func NewInfoSelfServiceLoginContinue() *Message {
	return &Message{
		ID:   InfoSelfServiceLoginContinue,
		Text: "Continue",
		Type: Info,
	}
}

func NewLoginEmailWithCodeSent() *Message {
	return &Message{
		ID:   InfoSelfServiceLoginEmailWithCodeSent,
		Type: Info,
		Text: "An email containing a code has been sent to the email address you provided. If you have not received an email, check the spelling of the address and retry the login.",
	}
}

func NewErrorValidationLoginCodeInvalidOrAlreadyUsed() *Message {
	return &Message{
		ID:   ErrorValidationLoginCodeInvalidOrAlreadyUsed,
		Text: "The login code is invalid or has already been used. Please try again.",
		Type: Error,
	}
}

func NewErrorValidationLoginRetrySuccessful() *Message {
	return &Message{
		ID:   ErrorValidationLoginRetrySuccess,
		Type: Error,
		Text: "The request was already completed successfully and can not be retried.",
	}
}

func NewInfoSelfServiceLoginCode() *Message {
	return &Message{
		ID:   InfoSelfServiceLoginCode,
		Type: Info,
		Text: "Sign in with code",
	}
}

func NewErrorValidationLoginLinkedCredentialsDoNotMatch() *Message {
	return &Message{
		ID:   ErrorValidationLoginLinkedCredentialsDoNotMatch,
		Text: "Linked credentials do not match.",
		Type: Error,
	}
}
