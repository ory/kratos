package schema

import (
	"fmt"

	"github.com/pkg/errors"

	"github.com/ory/jsonschema/v3"
)

func NewRequiredError(instancePtr, missing string) error {
	return errors.WithStack(&jsonschema.ValidationError{
		Message:     fmt.Sprintf("missing properties: %s", missing),
		InstancePtr: instancePtr,
		Context: &jsonschema.ValidationErrorContextRequired{
			Missing: []string{missing},
		},
	})
}

type ValidationErrorContextPasswordPolicyViolation struct {
	Reason string
}

func (r *ValidationErrorContextPasswordPolicyViolation) AddContext(_, _ string) {}

func (r *ValidationErrorContextPasswordPolicyViolation) FinishInstanceContext() {}

func NewPasswordPolicyViolationError(instancePtr string, reason string) error {
	return errors.WithStack(&jsonschema.ValidationError{
		Message:     fmt.Sprintf("the password does not fulfill the password policy because: %s", reason),
		InstancePtr: instancePtr,
		Context: &ValidationErrorContextPasswordPolicyViolation{
			Reason: reason,
		},
	})
}

type ValidationErrorContextInvalidCredentialsError struct{}

func (r *ValidationErrorContextInvalidCredentialsError) AddContext(_, _ string) {}

func (r *ValidationErrorContextInvalidCredentialsError) FinishInstanceContext() {}

func NewInvalidCredentialsError() error {
	return errors.WithStack(&jsonschema.ValidationError{
		Message:     `the provided credentials are invalid, check for spelling mistakes in your password or username, email address, or phone number`,
		InstancePtr: "#/",
		Context:     &ValidationErrorContextPasswordPolicyViolation{},
	})
}

type ValidationErrorContextDuplicateCredentialsError struct{}

func (r *ValidationErrorContextDuplicateCredentialsError) AddContext(_, _ string) {}

func (r *ValidationErrorContextDuplicateCredentialsError) FinishInstanceContext() {}

func NewDuplicateCredentialsError() error {
	return errors.WithStack(&jsonschema.ValidationError{
		Message:     `an account with the same identifier (email, phone, username, ...) exists already`,
		InstancePtr: "#/",
		Context:     &ValidationErrorContextDuplicateCredentialsError{},
	})
}
