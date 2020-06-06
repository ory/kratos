package schema

import (
	"fmt"

	"github.com/pkg/errors"

	"github.com/ory/jsonschema/v3"

	"github.com/ory/kratos/text"
)

type ValidationError struct {
	*jsonschema.ValidationError
	Messages text.Messages
}

func NewMinLengthError(instancePtr string, expected, actual int) error {
	return errors.WithStack(&ValidationError{
		ValidationError: &jsonschema.ValidationError{
			Message:     fmt.Sprintf("length must be >= %d, but got %d", expected, actual),
			InstancePtr: instancePtr,
		},
		Messages: new(text.Messages).Add(text.NewErrorValidationMinLength(expected, actual)),
	})
}

func NewRequiredError(missingPtr, missingFieldName string) error {
	return errors.WithStack(&ValidationError{
		ValidationError: &jsonschema.ValidationError{
			Message:     fmt.Sprintf("missing properties: %s", missingFieldName),
			InstancePtr: missingPtr,
			Context: &jsonschema.ValidationErrorContextRequired{
				Missing: []string{missingFieldName},
			},
		},
		Messages: new(text.Messages).Add(text.NewValidationErrorRequired(missingFieldName)),
	})
}

func NewInvalidFormatError(instancePtr, format, value string) error {
	return errors.WithStack(&ValidationError{
		ValidationError: &jsonschema.ValidationError{
			Message:     fmt.Sprintf("%q is not valid %q", value, format),
			InstancePtr: instancePtr,
		},
		Messages: new(text.Messages).Add(text.NewErrorValidationInvalidFormat(value, format)),
	})
}

type ValidationErrorContextPasswordPolicyViolation struct {
	Reason string
}

func (r *ValidationErrorContextPasswordPolicyViolation) AddContext(_, _ string) {}

func (r *ValidationErrorContextPasswordPolicyViolation) FinishInstanceContext() {}

func NewPasswordPolicyViolationError(instancePtr string, reason string) error {
	return errors.WithStack(&ValidationError{
		ValidationError: &jsonschema.ValidationError{
			Message:     fmt.Sprintf("the password does not fulfill the password policy because: %s", reason),
			InstancePtr: instancePtr,
			Context: &ValidationErrorContextPasswordPolicyViolation{
				Reason: reason,
			},
		},
		Messages: new(text.Messages).Add(text.NewErrorValidationPasswordPolicyViolation(reason)),
	})
}

type ValidationErrorContextInvalidCredentialsError struct{}

func (r *ValidationErrorContextInvalidCredentialsError) AddContext(_, _ string) {}

func (r *ValidationErrorContextInvalidCredentialsError) FinishInstanceContext() {}

func NewInvalidCredentialsError() error {
	return errors.WithStack(&ValidationError{
		ValidationError: &jsonschema.ValidationError{
			Message:     `the provided credentials are invalid, check for spelling mistakes in your password or username, email address, or phone number`,
			InstancePtr: "#/",
			Context:     &ValidationErrorContextPasswordPolicyViolation{},
		},
		Messages: new(text.Messages).Add(text.NewErrorValidationInvalidCredentials()),
	})
}

type ValidationErrorContextDuplicateCredentialsError struct{}

func (r *ValidationErrorContextDuplicateCredentialsError) AddContext(_, _ string) {}

func (r *ValidationErrorContextDuplicateCredentialsError) FinishInstanceContext() {}

func NewDuplicateCredentialsError() error {
	return errors.WithStack(&ValidationError{
		ValidationError: &jsonschema.ValidationError{
			Message:     `an account with the same identifier (email, phone, username, ...) exists already`,
			InstancePtr: "#/",
			Context:     &ValidationErrorContextDuplicateCredentialsError{},
		},
		Messages: new(text.Messages).Add(text.NewErrorValidationDuplicateCredentials()),
	})
}
