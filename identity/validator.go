package identity

import (
	"github.com/ory/gojsonschema"
	"github.com/ory/x/errorsx"
	"github.com/ory/x/stringsx"

	"github.com/ory/kratos/driver/configuration"
	"github.com/ory/kratos/schema"
)

type Validator struct {
	c configuration.Provider
	d validationDependencies
	v *schema.Validator
}

type validationDependencies interface {
	schema.PersistenceProvider
}

type ValidationProvider interface {
	IdentityValidator() *Validator
}

func NewValidator(c configuration.Provider, d validationDependencies) *Validator {
	return &Validator{
		c: c,
		d: d,
		v: schema.NewValidator(),
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

	v.d.SchemaPersister()

	s, err := v.d.SchemaPersister().GetSchema(i.TraitsSchemaID)
	if err != nil {
		return err
	}

	err = v.v.Validate(
		stringsx.Coalesce(
			s.URL,
			v.c.DefaultIdentityTraitsSchemaURL().String(),
		),
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
