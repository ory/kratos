package errors

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

func NewTOTPVerifierWrongError(instancePtr string) error {
	t := text.NewErrorValidationTOTPVerifierWrong()
	return errors.WithStack(&ValidationError{
		ValidationError: &jsonschema.ValidationError{
			Message:     t.Text,
			InstancePtr: instancePtr,
		},
		Messages: new(text.Messages).Add(t),
	})
}

func NewWebAuthnVerifierWrongError(instancePtr string) error {
	t := text.NewErrorValidationTOTPVerifierWrong()
	return errors.WithStack(&ValidationError{
		ValidationError: &jsonschema.ValidationError{
			Message:     t.Text,
			InstancePtr: instancePtr,
		},
		Messages: new(text.Messages).Add(t),
	})
}

func NewLookupAlreadyUsed() error {
	t := text.NewErrorValidationLookupAlreadyUsed()
	return errors.WithStack(&ValidationError{
		ValidationError: &jsonschema.ValidationError{
			Message:     t.Text,
			InstancePtr: "#/",
		},
		Messages: new(text.Messages).Add(t),
	})
}

func NewErrorValidationLookupInvalid() error {
	t := text.NewErrorValidationLookupInvalid()
	return errors.WithStack(&ValidationError{
		ValidationError: &jsonschema.ValidationError{
			Message:     t.Text,
			InstancePtr: "#/",
		},
		Messages: new(text.Messages).Add(t),
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

func NewMissingIdentifierError() error {
	return errors.WithStack(&ValidationError{
		ValidationError: &jsonschema.ValidationError{
			Message:     "could not find any identifiers",
			InstancePtr: "#/",
		},
		Messages: new(text.Messages).Add(text.NewErrorValidationIdentifierMissing()),
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

func NewNoLoginStrategyResponsible() error {
	return errors.WithStack(&ValidationError{
		ValidationError: &jsonschema.ValidationError{
			Message:     `could not find a strategy to login with`,
			InstancePtr: "#/",
		},
		Messages: new(text.Messages).Add(text.NewErrorValidationLoginNoStrategyFound()),
	})
}

func NewNoRegistrationStrategyResponsible() error {
	return errors.WithStack(&ValidationError{
		ValidationError: &jsonschema.ValidationError{
			Message:     `could not find a strategy to sign up with`,
			InstancePtr: "#/",
		},
		Messages: new(text.Messages).Add(text.NewErrorValidationRegistrationNoStrategyFound()),
	})
}

func NewNoSettingsStrategyResponsible() error {
	return errors.WithStack(&ValidationError{
		ValidationError: &jsonschema.ValidationError{
			Message:     `could not find a strategy to update settings with`,
			InstancePtr: "#/",
		},
		Messages: new(text.Messages).Add(text.NewErrorValidationSettingsNoStrategyFound()),
	})
}

func NewNoRecoveryStrategyResponsible() error {
	return errors.WithStack(&ValidationError{
		ValidationError: &jsonschema.ValidationError{
			Message:     `could not find a strategy to recover your account with`,
			InstancePtr: "#/",
		},
		Messages: new(text.Messages).Add(text.NewErrorValidationRecoveryNoStrategyFound()),
	})
}

func NewNoVerificationStrategyResponsible() error {
	return errors.WithStack(&ValidationError{
		ValidationError: &jsonschema.ValidationError{
			Message:     `could not find a strategy to verify your account with`,
			InstancePtr: "#/",
		},
		Messages: new(text.Messages).Add(text.NewErrorValidationVerificationNoStrategyFound()),
	})
}

func NewAddressNotVerifiedError() error {
	return errors.WithStack(&ValidationError{
		ValidationError: &jsonschema.ValidationError{
			Message:     `account address not yet verified`,
			InstancePtr: "#/",
		},
		Messages: new(text.Messages).Add(text.NewErrorValidationAddressNotVerified()),
	})
}

func NewNoTOTPDeviceRegistered() error {
	return errors.WithStack(&ValidationError{
		ValidationError: &jsonschema.ValidationError{
			Message:     `you have no TOTP device set up`,
			InstancePtr: "#/",
		},
		Messages: new(text.Messages).Add(text.NewErrorValidationNoTOTPDevice()),
	})
}

func NewNoLookupDefined() error {
	return errors.WithStack(&ValidationError{
		ValidationError: &jsonschema.ValidationError{
			Message:     `you have no backup recovery codes set up`,
			InstancePtr: "#/",
		},
		Messages: new(text.Messages).Add(text.NewErrorValidationNoLookup()),
	})
}

func NewNoWebAuthnRegistered() error {
	return errors.WithStack(&ValidationError{
		ValidationError: &jsonschema.ValidationError{
			Message:     `you have no WebAuthn device set up`,
			InstancePtr: "#/",
		},
		Messages: new(text.Messages).Add(text.NewErrorValidationNoWebAuthnDevice()),
	})
}
