package oidc

import (
	"encoding/json"
	"sync"

	"github.com/pkg/errors"
	"github.com/tidwall/sjson"

	"github.com/ory/gojsonschema"

	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/schema"
)

var _ identity.ValidationExtender = new(ValidationExtension)

type ValidationExtension struct {
	l      sync.Mutex
	i      *identity.Identity
	values json.RawMessage
}

func NewValidationExtension() *ValidationExtension {
	return &ValidationExtension{values: json.RawMessage("{}")}
}

func (e *ValidationExtension) WithIdentity(i *identity.Identity) identity.ValidationExtender {
	e.i = i
	return e
}

func (e *ValidationExtension) Call(value interface{}, config *schema.Extension, context *gojsonschema.JsonContext) error {
	if len(config.Mappings.Identity.Traits) > 0 {
		e.l.Lock()
		defer e.l.Unlock()

		for _, t := range config.Mappings.Identity.Traits {
			res, err := sjson.SetBytes(e.i.Traits, t.Path, value)
			if err != nil {
				return errors.Errorf(`schema: unable to apply mapping from path "%s.kratos.path.traits.identity.mappings": %s`, context.String("."), err)
			}

			if e.values, err = sjson.SetBytes(e.values, t.Path, value); err != nil {
				return errors.Errorf(`schema: unable to apply mapping from path "%s.kratos.path.traits.identity.mappings": %s`, context.String("."), err)
			}

			e.i.Traits = res
		}
	}
	return nil
}

func (e *ValidationExtension) Values() json.RawMessage {
	e.l.Lock()
	defer e.l.Unlock()
	return e.values
}
