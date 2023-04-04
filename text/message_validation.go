// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package text

import (
	"fmt"
)

func NewValidationErrorGeneric(reason string) *Message {
	return &Message{
		ID:      ErrorValidationGeneric,
		Text:    reason,
		Type:    Error,
		Context: nil,
	}
}

func NewValidationErrorRequired(missing string) *Message {
	return &Message{
		ID:   ErrorValidationRequired,
		Text: fmt.Sprintf("Property %s is missing.", missing),
		Type: Error,
		Context: context(map[string]interface{}{
			"property": missing,
		}),
	}
}

func NewErrorValidationMinLength(reason string) *Message {
	return &Message{
		ID:      ErrorValidationMinLength,
		Text:    reason,
		Type:    Error,
		Context: context(nil),
	}
}

func NewErrorValidationMaxLength(reason string) *Message {
	return &Message{
		ID:      ErrorValidationMaxLength,
		Text:    reason,
		Type:    Error,
		Context: context(nil),
	}
}

func NewErrorValidationInvalidFormat(reason string) *Message {
	return &Message{
		ID:      ErrorValidationInvalidFormat,
		Text:    reason,
		Type:    Error,
		Context: context(nil),
	}
}

func NewErrorValidationMinimum(reason string) *Message {
	return &Message{
		ID:      ErrorValidationMinimum,
		Text:    reason,
		Type:    Error,
		Context: context(nil),
	}
}

func NewErrorValidationExclusiveMinimum(reason string) *Message {
	return &Message{
		ID:      ErrorValidationExclusiveMinimum,
		Text:    reason,
		Type:    Error,
		Context: context(nil),
	}
}

func NewErrorValidationMaximum(reason string) *Message {
	return &Message{
		ID:      ErrorValidationMaximum,
		Text:    reason,
		Type:    Error,
		Context: context(nil),
	}
}

func NewErrorValidationExclusiveMaximum(reason string) *Message {
	return &Message{
		ID:      ErrorValidationExclusiveMaximum,
		Text:    reason,
		Type:    Error,
		Context: context(nil),
	}
}

func NewErrorValidationMultipleOf(reason string) *Message {
	return &Message{
		ID:      ErrorValidationMultipleOf,
		Text:    reason,
		Type:    Error,
		Context: context(nil),
	}
}

func NewErrorValidationMaxItems(reason string) *Message {
	return &Message{
		ID:      ErrorValidationMaxItems,
		Text:    reason,
		Type:    Error,
		Context: context(nil),
	}
}

func NewErrorValidationMinItems(reason string) *Message {
	return &Message{
		ID:      ErrorValidationMinItems,
		Text:    reason,
		Type:    Error,
		Context: context(nil),
	}
}

func NewErrorValidationUniqueItems(reason string) *Message {
	return &Message{
		ID:      ErrorValidationUniqueItems,
		Text:    reason,
		Type:    Error,
		Context: context(nil),
	}
}

func NewErrorValidationWrongType(reason string) *Message {
	return &Message{
		ID:      ErrorValidationWrongType,
		Text:    reason,
		Type:    Error,
		Context: context(nil),
	}
}

func NewErrorValidationPasswordPolicyViolation(reason string) *Message {
	return &Message{
		ID:   ErrorValidationPasswordPolicyViolation,
		Text: fmt.Sprintf("The password can not be used because %s.", reason),
		Type: Error,
		Context: context(map[string]interface{}{
			"reason": reason,
		}),
	}
}

func NewErrorValidationInvalidCredentials() *Message {
	return &Message{
		ID:      ErrorValidationInvalidCredentials,
		Text:    "The provided credentials are invalid, check for spelling mistakes in your password or username, email address, or phone number.",
		Type:    Error,
		Context: context(nil),
	}
}

func NewErrorValidationDuplicateCredentials() *Message {
	return &Message{
		ID:      ErrorValidationDuplicateCredentials,
		Text:    "An account with the same identifier (email, phone, username, ...) exists already.",
		Type:    Error,
		Context: context(nil),
	}
}

func NewErrorValidationDuplicateCredentialsOnOIDCLink() *Message {
	return &Message{
		ID:      ErrorValidationDuplicateCredentialsOnOIDCLink,
		Text:    "An account with the same identifier (email, phone, username, ...) exists already. Please sign in to your existing account and link your social profile in the settings page.",
		Type:    Error,
		Context: context(nil),
	}
}

func NewErrorValidationTOTPVerifierWrong() *Message {
	return &Message{
		ID:      ErrorValidationTOTPVerifierWrong,
		Text:    "The provided authentication code is invalid, please try again.",
		Type:    Error,
		Context: context(nil),
	}
}

func NewErrorValidationLookupAlreadyUsed() *Message {
	return &Message{
		ID:      ErrorValidationLookupAlreadyUsed,
		Text:    "This backup recovery code has already been used.",
		Type:    Error,
		Context: context(nil),
	}
}

func NewErrorValidationLookupInvalid() *Message {
	return &Message{
		ID:      ErrorValidationLookupInvalid,
		Text:    "The backup recovery code is not valid.",
		Type:    Error,
		Context: context(nil),
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
		ID:      ErrorValidationNoTOTPDevice,
		Text:    "You have no TOTP device set up.",
		Type:    Error,
		Context: context(nil),
	}
}

func NewErrorValidationNoLookup() *Message {
	return &Message{
		ID:      ErrorValidationNoLookup,
		Text:    "You have no backup recovery codes set up.",
		Type:    Error,
		Context: context(nil),
	}
}

func NewErrorValidationNoWebAuthnDevice() *Message {
	return &Message{
		ID:      ErrorValidationNoWebAuthnDevice,
		Text:    "You have no WebAuthn device set up.",
		Type:    Error,
		Context: context(nil),
	}
}

func NewErrorValidationSuchNoWebAuthnUser() *Message {
	return &Message{
		ID:      ErrorValidationSuchNoWebAuthnUser,
		Text:    "This account does not exist or has no security key set up.",
		Type:    Error,
		Context: context(nil),
	}
}
