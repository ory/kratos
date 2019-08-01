package schema

import (
	"fmt"

	"github.com/ory/gojsonschema"
)

func NewRequiredError(value interface{}, context *gojsonschema.JsonContext) error {
	err := &gojsonschema.RequiredError{}
	err.SetContext(context)
	err.SetValue(value)
	err.SetType("required")

	err.SetDescription(gojsonschema.FormatErrorDescription(
		gojsonschema.DefaultLocale{}.Required(),
		gojsonschema.ErrorDetails{
			"context":  context.String("."),
			"value":    fmt.Sprintf("%v", value),
			"field":    context.String("."),
			"property": context.String("."),
		},
	))
	return ResultErrors{err}
}

func NewPasswordPolicyValidation(value interface{}, reason string, context *gojsonschema.JsonContext) error {
	err := &gojsonschema.ResultErrorFields{}
	err.SetContext(context)
	err.SetValue(value)
	err.SetType("password_policy_validation")

	err.SetDescription(gojsonschema.FormatErrorDescription(
		`{{.property}} does not meet the password policy because: {{.reason}}`,
		gojsonschema.ErrorDetails{
			"context":  context.String("."),
			"value":    fmt.Sprintf("%v", value),
			"field":    context.String("."),
			"property": context.String("."),
			"reason":   reason,
		},
	))
	return ResultErrors{err}
}

func NewInvalidCredentialsError() error {
	err := &gojsonschema.ResultErrorFields{}
	context := gojsonschema.NewJsonContext("", nil)
	err.SetContext(context)
	err.SetValue("")
	err.SetType("invalid_credentials")

	err.SetDescription(gojsonschema.FormatErrorDescription(
		`The provided credentials are invalid. Check for spelling mistakes in your password or username, email address, or phone number.`,
		gojsonschema.ErrorDetails{
			"context":  context.String("."),
			"value":    fmt.Sprintf("%v", ""),
			"field":    context.String("."),
			"property": context.String("."),
		},
	))
	return ResultErrors{err}
}
