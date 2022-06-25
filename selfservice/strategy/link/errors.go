package link

import (
	"github.com/ory/jsonschema/v3"
	"github.com/ory/kratos/schema"
	"github.com/ory/kratos/text"
	"github.com/pkg/errors"
)

func NewValidationVerificationTokenInvalidOrAlreadyUsedError() error {
	return errors.WithStack(&schema.ValidationError{
		ValidationError: &jsonschema.ValidationError{
			Message:     `the verification token is invalid or has already been used. Please retry the flow`,
			InstancePtr: "#/",
		},
		Messages: new(text.Messages).Add(text.NewErrorValidationVerificationTokenInvalidOrAlreadyUsed()),
	})
}
