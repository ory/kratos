package oidc

import (
	"sync"

	"github.com/pkg/errors"
	"github.com/tidwall/sjson"

	"github.com/ory/gojsonschema"

	"github.com/ory/hive-cloud/hive/identity"
	"github.com/ory/hive-cloud/hive/schema"
)

var _ identity.ValidationExtender = new(ValidationExtension)

type ValidationExtension struct {
	l sync.Mutex
	i *identity.Identity
}

func NewValidationExtension() *ValidationExtension {
	return &ValidationExtension{}
}

func (e *ValidationExtension) WithIdentity(i *identity.Identity) identity.ValidationExtender {
	ve := *e
	ve.i = i
	return &ve
}

func (e *ValidationExtension) Call(value interface{}, config *schema.Extension, context *gojsonschema.JsonContext) error {
	if len(config.Mappings.Identity.Traits) > 0 {
		e.l.Lock()
		defer e.l.Unlock()

		for _, t := range config.Mappings.Identity.Traits {
			res, err := sjson.SetBytes(e.i.Traits, t.Path, value)
			if err != nil {
				return errors.Errorf(`schema: unable to apply mapping from path "%s.hive.path.traits.identity.mappings": %s`, context.String("."), err)
			}
			e.i.Traits = res
		}
	}
	return nil
}
