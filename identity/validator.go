package identity

import (
	"encoding/json"

	"github.com/ory/jsonschema/v3"

	"github.com/ory/x/errorsx"

	"github.com/ory/kratos/driver/configuration"
	"github.com/ory/kratos/schema"
)

type (
	validatorDependencies interface {
		IdentityTraitsSchemas() schema.Schemas
	}
	Validator struct {
		v *schema.Validator
		d validatorDependencies
		c configuration.Provider
	}
	ValidationProvider interface {
		IdentityValidator() *Validator
	}
)

func NewValidator(d validatorDependencies, c configuration.Provider) *Validator {
	return &Validator{
		v: schema.NewValidator(),
		d: d,
		c: c,
	}
}

func (v *Validator) ValidateWithRunner(i *Identity, runners ...schema.Extension) error {
	runner, err := schema.NewExtensionRunner(
		schema.ExtensionRunnerIdentityMetaSchema,
		runners...,
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

func (v *Validator) Validate(i *Identity) error {
	return v.ValidateWithRunner(i,
		NewSchemaExtensionCredentials(i),
		NewSchemaExtensionVerify(i, v.c.SelfServiceVerificationRequestLifespan()),
	)
}
