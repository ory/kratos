package identity

import (
	"github.com/tidwall/sjson"

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
	runner, err := schema.NewExtensionRunner(schema.ExtensionRunnerIdentityMetaSchema, runners...)
	if err != nil {
		return err
	}

	s, err := v.d.IdentityTraitsSchemas().GetByID(i.SchemaID)
	if err != nil {
		return err
	}

	traits, err := sjson.SetRawBytes([]byte(`{}`), "traits", i.Traits)
	if err != nil {
		return err
	}

	return v.v.Validate(s.URL.String(), traits, schema.WithExtensionRunner(runner))
}

func (v *Validator) Validate(i *Identity) error {
	return v.ValidateWithRunner(i,
		NewSchemaExtensionCredentials(i),
		NewSchemaExtensionVerification(i, v.c.SelfServiceFlowVerificationRequestLifespan()),
		NewSchemaExtensionRecovery(i),
	)
}
