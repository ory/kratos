// Copyright © 2022 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package code

import (
	"net/http"

	"github.com/pkg/errors"

	"github.com/ory/jsonschema/v3"
	"github.com/ory/kratos/schema"
	"github.com/ory/kratos/text"
)

type ValidationErrorContextCodePolicyViolation struct {
	Reason string
}

type CodeSentError struct {
	*schema.ValidationError
}

func (e CodeSentError) Error() string {
	return e.ValidationError.Error()
}

func (e CodeSentError) Unwrap() error {
	return e.ValidationError
}

func (e CodeSentError) StatusCode() int {
	return http.StatusOK
}

func NewCodeSentError() error {
	return CodeSentError{
		ValidationError: &schema.ValidationError{
			ValidationError: &jsonschema.ValidationError{
				Message:     `access code has been sent`,
				InstancePtr: "#/",
				Context:     &ValidationErrorContextCodePolicyViolation{},
			},
			Messages: new(text.Messages).Add(text.NewErrorCodeSent()),
		}}
}

func (r *ValidationErrorContextCodePolicyViolation) AddContext(_, _ string) {}

func (r *ValidationErrorContextCodePolicyViolation) FinishInstanceContext() {}

func NewInvalidCodeError() error {
	return errors.WithStack(&schema.ValidationError{
		ValidationError: &jsonschema.ValidationError{
			Message:     `the provided code is invalid, check for spelling mistakes in the code or phone number`,
			InstancePtr: "#/",
			Context:     &ValidationErrorContextCodePolicyViolation{},
		},
		Messages: new(text.Messages).Add(text.NewErrorValidationInvalidCode()),
	})
}

func NewAttemptsExceededError() error {
	return errors.WithStack(&schema.ValidationError{
		ValidationError: &jsonschema.ValidationError{
			Message:     `maximum code verification attempts exceeded`,
			InstancePtr: "#/",
			Context:     &ValidationErrorContextCodePolicyViolation{},
		},
		Messages: new(text.Messages).Add(text.NewErrorValidationInvalidCode()),
	})
}

func NewSMSSpamError() error {
	return errors.WithStack(&schema.ValidationError{
		ValidationError: &jsonschema.ValidationError{
			Message:     `sms spam detected`,
			InstancePtr: "#/",
			Context:     &ValidationErrorContextCodePolicyViolation{},
		},
		Messages: new(text.Messages).Add(text.NewErrorValidationSMSSpam()),
	})
}
