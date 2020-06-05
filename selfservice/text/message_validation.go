package text

import (
	"fmt"
)

const (
	ErrorValidation ID = 4000000 + iota
	ErrorValidationGeneric
	ErrorValidationRequired
	ErrorValidationMinLength
	ErrorValidationInvalidFormat
	ErrorValidationPasswordPolicyViolation
	ErrorValidationInvalidCredentials
	ErrorValidationDuplicateCredentials
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
