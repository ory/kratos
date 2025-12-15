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

func NewInfoLoginLinkMessage(dupIdentifier, provider, newLoginURL string, availableCredentials, availableProviders []string) *Message {
	return &Message{
		ID:   InfoSelfServiceLoginLink,
		Type: Info,
		Text: fmt.Sprintf(
			"You tried to sign in with %q, but that email is already used by another account. Sign in to your account with one of the options below to add your account %[1]q at %q as another way to sign in.",
			dupIdentifier,
			provider,
		),
		Context: context(map[string]any{
			"duplicateIdentifier":        dupIdentifier,
			"provider":                   provider,
			"newLoginUrl":                newLoginURL,
			"duplicate_identifier":       dupIdentifier,
			"new_login_url":              newLoginURL,
			"available_credential_types": availableCredentials,
			"available_providers":        availableProviders,
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

func NewInfoLoginPassword() *Message {
	return &Message{
		ID:   InfoSelfServiceLoginPassword,
		Text: "Sign in with password",
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

func NewInfoLoginWith(provider string, providerId string) *Message {
	return &Message{
		ID:   InfoSelfServiceLoginWith,
		Text: fmt.Sprintf("Sign in with %s", provider),
		Type: Info,
		Context: context(map[string]any{
			"provider":    provider,
			"provider_id": providerId,
		}),
	}
}

func NewInfoLoginWithAndLink(provider string) *Message {
	return &Message{
		ID:   InfoSelfServiceLoginWithAndLink,
		Text: fmt.Sprintf("Confirm with %s", provider),
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
		Text: "Sign in with hardware key",
		Type: Info,
	}
}

func NewInfoSelfServiceLoginPasskey() *Message {
	return &Message{
		ID:   InfoSelfServiceLoginPasskey,
		Text: "Sign in with passkey",
		Type: Info,
	}
}

func NewInfoSelfServiceContinueLoginWebAuthn() *Message {
	return &Message{
		ID:   InfoSelfServiceLoginContinueWebAuthn,
		Text: "Sign in with hardware key",
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

func NewLoginCodeSent() *Message {
	return &Message{
		ID:   InfoSelfServiceLoginCodeSent,
		Type: Info,
		Text: "A code was sent to the address you provided. If you didn't receive it, please check the spelling of the address and try again.",
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
		Text: "Send sign in code",
	}
}

func NewErrorValidationLoginLinkedCredentialsDoNotMatch() *Message {
	return &Message{
		ID:   ErrorValidationLoginLinkedCredentialsDoNotMatch,
		Text: "Linked credentials do not match.",
		Type: Error,
	}
}

func NewErrorValidationAddressUnknown() *Message {
	return &Message{
		ID:   ErrorValidationLoginAddressUnknown,
		Text: "The address you entered does not match any known addresses in the current account.",
		Type: Error,
	}
}

func NewInfoSelfServiceLoginCodeMFA() *Message {
	return &Message{
		ID:   InfoSelfServiceLoginCodeMFA,
		Type: Info,
		Text: "Request code to continue",
	}
}

func NewInfoSelfServiceLoginAAL2CodeAddress(channel string, to string) *Message {
	return &Message{
		ID:   InfoSelfServiceLoginAAL2CodeAddress,
		Type: Info,
		Text: fmt.Sprintf("Send code to %s", to),
		Context: context(map[string]any{
			"address": to,
			"channel": channel,
		}),
	}
}
