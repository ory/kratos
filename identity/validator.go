package identity

import (
	"encoding/json"

	"github.com/ory/jsonschema/v3"

	"github.com/ory/kratos/schema"

	"github.com/ory/x/errorsx"
)

type (
	validatorDependencies interface {
		IdentityTraitsSchemas() schema.Schemas
	}
	Validator struct {
		v *schema.Validator
		d validatorDependencies
	}
	ValidationProvider interface {
		IdentityValidator() *Validator
	}
)

func NewValidator(d validatorDependencies) *Validator {
	return &Validator{
		v: schema.NewValidator(),
		d: d,
	}
}

func (v *Validator) Validate(i *Identity) error {
	runner, err := schema.NewExtensionRunner(
		schema.ExtensionRunnerIdentityMetaSchema,
		NewValidationExtensionRunner(i).Runner,
	)
	if err != nil {
		return err
	}

	s, err := v.d.IdentityTraitsSchemas().GetByID(i.TraitsSchemaID)
	if err != nil {
		return err
	}

	err = v.v.Validate(
		s.URL.String(),
		json.RawMessage(i.Traits),
		schema.WithExtensionRunner(runner),
	)

	switch e := errorsx.Cause(err).(type) {
	case *jsonschema.ValidationError:
		return schema.ContextSetRoot(e, "traits")
	}

	return err
}
