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

func NewErrorValidationMinLength(expected, actual int) *Message {
	return &Message{
		ID:   ErrorValidationMinLength,
		Text: fmt.Sprintf("Length must be >= %d, but got %d.", expected, actual),
		Type: Error,
		Context: context(map[string]interface{}{
			"expected_length": expected,
			"actual_length":   actual,
		}),
	}
}

func NewErrorValidationInvalidFormat(format, value string) *Message {
	return &Message{
		ID:   ErrorValidationInvalidFormat,
		Text: fmt.Sprintf("%q is not valid %q", value, format),
		Type: Error,
		Context: context(map[string]interface{}{
			"expected_format": format,
			"actual_value":    value,
		}),
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
		ID:      ErrorValidationLookupAlreadyUsed,
		Text:    "The backup recovery code is not valid.",
		Type:    Error,
		Context: context(nil),
	}
}

func NewErrorValidationIdentifierMissing() *Message {
	return &Message{
		ID:   ErrorValidationIdentifierMissing,
		Text: "Could not find any login identifiers. Did you forget to set them?",
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
