package identity

import (
	"github.com/ory/gojsonschema"
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

type ValidationExtender interface {
	WithIdentity(*Identity) ValidationExtender
	schema.ValidationExtender
}

func (v *Validator) Validate(i *Identity) error {
	es := []schema.ValidationExtender{
		NewValidationExtensionIdentifier().WithIdentity(i),
	}

	s, err := v.d.IdentityTraitsSchemas().GetByID(i.TraitsSchemaID)
	if err != nil {
		return err
	}

	err = v.v.Validate(
		s.URL.String(),
		gojsonschema.NewBytesLoader(i.Traits),
		es...,
	)

	switch errs := errorsx.Cause(err).(type) {
	case schema.ResultErrors:
		for k, err := range errs {
			errs[k].SetContext(schema.ContextSetRoot(schema.ContextRemoveRootStub(err.Context()), "traits"))
		}
		return errs
	}
	return err
}
