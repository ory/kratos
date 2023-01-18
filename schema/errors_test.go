// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package schema

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ory/jsonschema/v3"

	"github.com/ory/kratos/text"
)

func TestListValidationErrors(t *testing.T) {
	testErr := ValidationListError{}

	assert.False(t, testErr.HasErrors())

	testErr.WithError("#/traits/password", "error message", new(text.Messages).Add(text.NewErrorValidationDuplicateCredentials()))
	assert.True(t, testErr.HasErrors())
	assert.Len(t, testErr.Validations, 1)

	validationError := &ValidationError{
		ValidationError: &jsonschema.ValidationError{
			Message:     `the provided credentials are invalid, check for spelling mistakes in your password or username, email address, or phone number`,
			InstancePtr: "#/",
			Context:     &ValidationErrorContextPasswordPolicyViolation{},
		},
		Messages: new(text.Messages).Add(text.NewErrorValidationInvalidCredentials()),
	}
	testErr.Add(validationError)
	assert.Len(t, testErr.Validations, 2)
	assert.Equal(t, "2 validation errors occurred:"+
		"\n(0) I[#/traits/password] S[] error message"+
		"\n(1) I[#/] S[] the provided credentials are invalid, check for spelling mistakes in your password or username, email address, or phone number",
		testErr.Error())
}
