// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package text

import (
	"fmt"
	"strings"

	"github.com/ory/x/stringslice"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func NewValidationErrorGeneric(reason string) *Message {
	return &Message{
		ID:   ErrorValidationGeneric,
		Text: reason,
		Type: Error,
		Context: context(map[string]any{
			"reason": reason,
		}),
	}
}

func NewValidationErrorRequired(missing string) *Message {
	return &Message{
		ID:   ErrorValidationRequired,
		Text: fmt.Sprintf("Property %s is missing.", missing),
		Type: Error,
		Context: context(map[string]any{
			"property": missing,
		}),
	}
}

func NewErrorValidationMinLength(minLength, actualLength int) *Message {
	return &Message{
		ID:   ErrorValidationMinLength,
		Text: fmt.Sprintf("length must be >= %d, but got %d", minLength, actualLength),
		Type: Error,
		Context: context(map[string]any{
			"min_length":    minLength,
			"actual_length": actualLength,
		}),
	}
}

func NewErrorValidationMaxLength(maxLength, actualLength int) *Message {
	return &Message{
		ID:   ErrorValidationMaxLength,
		Text: fmt.Sprintf("length must be <= %d, but got %d", maxLength, actualLength),
		Type: Error,
		Context: context(map[string]any{
			"max_length":    maxLength,
			"actual_length": actualLength,
		}),
	}
}

func NewErrorValidationInvalidFormat(pattern string) *Message {
	return &Message{
		ID:   ErrorValidationInvalidFormat,
		Text: fmt.Sprintf("does not match pattern %q", pattern),
		Type: Error,
		Context: context(map[string]any{
			"pattern": pattern,
		}),
	}
}

func NewErrorValidationMinimum(minimum, actual float64) *Message {
	return &Message{
		ID:   ErrorValidationMinimum,
		Text: fmt.Sprintf("must be >= %v but found %v", minimum, actual),
		Type: Error,
		Context: context(map[string]any{
			"minimum": minimum,
			"actual":  actual,
		}),
	}
}

func NewErrorValidationExclusiveMinimum(minimum, actual float64) *Message {
	return &Message{
		ID:   ErrorValidationExclusiveMinimum,
		Text: fmt.Sprintf("must be > %v but found %v", minimum, actual),
		Type: Error,
		Context: context(map[string]any{
			"minimum": minimum,
			"actual":  actual,
		}),
	}
}

func NewErrorValidationMaximum(maximum, actual float64) *Message {
	return &Message{
		ID:   ErrorValidationMaximum,
		Text: fmt.Sprintf("must be <= %v but found %v", maximum, actual),
		Type: Error,
		Context: context(map[string]any{
			"maximum": maximum,
			"actual":  actual,
		}),
	}
}

func NewErrorValidationExclusiveMaximum(maximum, actual float64) *Message {
	return &Message{
		ID:   ErrorValidationExclusiveMaximum,
		Text: fmt.Sprintf("must be < %v but found %v", maximum, actual),
		Type: Error,
		Context: context(map[string]any{
			"maximum": maximum,
			"actual":  actual,
		}),
	}
}

func NewErrorValidationMultipleOf(base, actual float64) *Message {
	return &Message{
		ID:   ErrorValidationMultipleOf,
		Text: fmt.Sprintf("%v not multipleOf %v", actual, base),
		Type: Error,
		Context: context(map[string]any{
			"base":   base,
			"actual": actual,
		}),
	}
}

func NewErrorValidationMaxItems(maxItems, actualItems int) *Message {
	return &Message{
		ID:   ErrorValidationMaxItems,
		Text: fmt.Sprintf("maximum %d items allowed, but found %d items", maxItems, actualItems),
		Type: Error,
		Context: context(map[string]any{
			"max_items":    maxItems,
			"actual_items": actualItems,
		}),
	}
}

func NewErrorValidationMinItems(minItems, actualItems int) *Message {
	return &Message{
		ID:   ErrorValidationMinItems,
		Text: fmt.Sprintf("minimum %d items allowed, but found %d items", minItems, actualItems),
		Type: Error,
		Context: context(map[string]any{
			"min_items":    minItems,
			"actual_items": actualItems,
		}),
	}
}

func NewErrorValidationUniqueItems(indexA, indexB int) *Message {
	return &Message{
		ID:   ErrorValidationUniqueItems,
		Text: fmt.Sprintf("items at index %d and %d are equal", indexA, indexB),
		Type: Error,
		Context: context(map[string]any{
			"index_a": indexA,
			"index_b": indexB,
		}),
	}
}

func NewErrorValidationWrongType(allowedTypes []string, actualType string) *Message {
	return &Message{
		ID:   ErrorValidationWrongType,
		Text: fmt.Sprintf("expected %s, but got %s", strings.Join(allowedTypes, " or "), actualType),
		Type: Error,
		Context: context(map[string]any{
			"allowed_types": allowedTypes,
			"actual_type":   actualType,
		}),
	}
}

func NewErrorValidationConst(expected any) *Message {
	return &Message{
		ID:   ErrorValidationConst,
		Text: fmt.Sprintf("must be equal to constant %v", expected),
		Type: Error,
		Context: context(map[string]any{
			"expected": expected,
		}),
	}
}

func NewErrorValidationConstGeneric() *Message {
	return &Message{
		ID:   ErrorValidationConstGeneric,
		Text: "const failed",
		Type: Error,
	}
}

func NewErrorValidationPasswordPolicyViolationGeneric(reason string) *Message {
	return &Message{
		ID:   ErrorValidationPasswordPolicyViolationGeneric,
		Text: fmt.Sprintf("The password can not be used because %s.", reason),
		Type: Error,
		Context: context(map[string]any{
			"reason": reason,
		}),
	}
}

func NewErrorValidationPasswordIdentifierTooSimilar() *Message {
	return &Message{
		ID:   ErrorValidationPasswordIdentifierTooSimilar,
		Text: "The password can not be used because it is too similar to the identifier.",
		Type: Error,
	}
}

func NewErrorValidationPasswordMinLength(minLength, actualLength int) *Message {
	return &Message{
		ID:   ErrorValidationPasswordMinLength,
		Text: fmt.Sprintf("The password must be at least %d characters long, but got %d.", minLength, actualLength),
		Type: Error,
		Context: context(map[string]any{
			"min_length":    minLength,
			"actual_length": actualLength,
		}),
	}
}

func NewErrorValidationPasswordMaxLength(maxLength, actualLength int) *Message {
	return &Message{
		ID:   ErrorValidationPasswordMaxLength,
		Text: fmt.Sprintf("The password must be at most %d characters long, but got %d.", maxLength, actualLength),
		Type: Error,
		Context: context(map[string]any{
			"max_length":    maxLength,
			"actual_length": actualLength,
		}),
	}
}

func NewErrorValidationPasswordNewSameAsOld() *Message {
	return &Message{
		ID:   ErrorValidationPasswordNewSameAsOld,
		Text: "The new password must be different from the old password.",
		Type: Error,
	}
}

func NewErrorValidationPasswordTooManyBreaches(breaches int64) *Message {
	return &Message{
		ID:   ErrorValidationPasswordTooManyBreaches,
		Text: "The password has been found in data breaches and must no longer be used.",
		Type: Error,
		Context: context(map[string]any{
			"breaches": breaches,
		}),
	}
}

func NewErrorValidationInvalidCredentials() *Message {
	return &Message{
		ID:   ErrorValidationInvalidCredentials,
		Text: "The provided credentials are invalid, check for spelling mistakes in your password or username, email address, or phone number.",
		Type: Error,
	}
}

func NewErrorValidationAccountNotFound() *Message {
	return &Message{
		ID:   ErrorValidationAccountNotFound,
		Text: "This account does not exist or has no login method configured.",
		Type: Error,
	}
}

func NewErrorValidationDuplicateCredentials() *Message {
	return &Message{
		ID:   ErrorValidationDuplicateCredentials,
		Text: "An account with the same identifier (email, phone, username, ...) exists already.",
		Type: Error,
	}
}

func NewErrorValidationDuplicateCredentialsWithHints(availableCredentialTypes []string, availableOIDCProviders []string, credentialIdentifierHint string) *Message {
	identifier := credentialIdentifierHint
	if identifier == "" {
		identifier = "an email, phone, or username"
	}
	oidcProviders := make([]string, 0, len(availableOIDCProviders))
	for _, provider := range availableOIDCProviders {
		oidcProviders = append(oidcProviders, cases.Title(language.English).String(provider))
	}

	reason := fmt.Sprintf("You tried signing in with %s which is already in use by another account.", identifier)
	if len(availableCredentialTypes) > 0 {
		humanReadable := make([]string, 0, len(availableCredentialTypes))
		for _, cred := range availableCredentialTypes {
			switch cred {
			case "password":
				humanReadable = append(humanReadable, "your password")
			case "oidc", "saml":
				humanReadable = append(humanReadable, "social sign in")
			case "webauthn":
				humanReadable = append(humanReadable, "your passkey or a security key")
			case "passkey":
				humanReadable = append(humanReadable, "your passkey")
			}
		}
		if len(humanReadable) == 0 {
			// show at least some hint
			// also our example message generation tool runs into this case
			humanReadable = append(humanReadable, availableCredentialTypes...)
		}

		humanReadable = stringslice.Unique(humanReadable)

		// Final format: "You can sign in using foo, bar, or baz."
		if len(humanReadable) > 1 {
			humanReadable[len(humanReadable)-1] = "or " + humanReadable[len(humanReadable)-1]
		}
		if len(humanReadable) > 0 {
			reason += fmt.Sprintf(" You can sign in using %s.", strings.Join(humanReadable, ", "))
		}
	}
	if len(oidcProviders) > 0 {
		reason += fmt.Sprintf(" You can sign in using one of the following social sign in providers: %s.", strings.Join(oidcProviders, ", "))
	}

	return &Message{
		ID:   ErrorValidationDuplicateCredentialsWithHints,
		Text: reason,
		Type: Error,
		Context: context(map[string]any{
			"available_credential_types": availableCredentialTypes,
			"available_oidc_providers":   availableOIDCProviders,
			"credential_identifier_hint": credentialIdentifierHint,
		}),
	}
}

func NewErrorValidationDuplicateCredentialsOnOIDCLink() *Message {
	return &Message{
		ID: ErrorValidationDuplicateCredentialsOnOIDCLink,
		Text: "An account with the same identifier (email, phone, username, ...) exists already. " +
			"Please sign in to your existing account to link your social profile.",
		Type: Error,
	}
}

func NewErrorValidationTOTPVerifierWrong() *Message {
	return &Message{
		ID:   ErrorValidationTOTPVerifierWrong,
		Text: "The provided authentication code is invalid, please try again.",
		Type: Error,
	}
}

func NewErrorValidationLookupAlreadyUsed() *Message {
	return &Message{
		ID:   ErrorValidationLookupAlreadyUsed,
		Text: "This backup recovery code has already been used.",
		Type: Error,
	}
}

func NewErrorValidationLookupInvalid() *Message {
	return &Message{
		ID:   ErrorValidationLookupInvalid,
		Text: "The backup recovery code is not valid.",
		Type: Error,
	}
}

func NewErrorValidationIdentifierMissing() *Message {
	return &Message{
		ID:   ErrorValidationIdentifierMissing,
		Text: "Could not find any login identifiers. Did you forget to set them? This could also be caused by a server misconfiguration.",
		Type: Error,
	}
}

func NewErrorValidationAddressNotVerified() *Message {
	return &Message{
		ID:   ErrorValidationAddressNotVerified,
		Text: "Account not active yet. Did you forget to verify your email address?",
		Type: Error,
	}
}

func NewErrorValidationNoTOTPDevice() *Message {
	return &Message{
		ID:   ErrorValidationNoTOTPDevice,
		Text: "You have no TOTP device set up.",
		Type: Error,
	}
}

func NewErrorValidationNoLookup() *Message {
	return &Message{
		ID:   ErrorValidationNoLookup,
		Text: "You have no backup recovery codes set up.",
		Type: Error,
	}
}

func NewErrorValidationNoWebAuthnDevice() *Message {
	return &Message{
		ID:   ErrorValidationNoWebAuthnDevice,
		Text: "You have no WebAuthn device set up.",
		Type: Error,
	}
}

func NewErrorValidationSuchNoWebAuthnUser() *Message {
	return &Message{
		ID:   ErrorValidationSuchNoWebAuthnUser,
		Text: "This account does not exist or has no security key set up.",
		Type: Error,
	}
}

func NewErrorValidationNoCodeUser() *Message {
	return &Message{
		ID:   ErrorValidationNoCodeUser,
		Text: "This account does not exist or has not setup sign in with code.",
		Type: Error,
	}
}

func NewErrorValidationTraitsMismatch() *Message {
	return &Message{
		ID:   ErrorValidationTraitsMismatch,
		Text: "The provided traits do not match the traits previously associated with this flow.",
		Type: Error,
	}
}

func NewErrorCaptchaFailed() *Message {
	return &Message{
		ID:   ErrorValidationCaptchaError,
		Text: "Captcha verification failed, please try again.",
		Type: Error,
	}
}
