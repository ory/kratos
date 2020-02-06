package oidc

import (
	"sync"

	"github.com/tidwall/sjson"

	"github.com/ory/jsonschema/v3"

	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/schema"
)

type ValidationExtensionRunner struct {
	l sync.Mutex
	i *identity.Identity
}

func NewValidationExtensionRunner(i *identity.Identity) *ValidationExtensionRunner {
	return &ValidationExtensionRunner{i: i}
}

func (r *ValidationExtensionRunner) Runner(ctx jsonschema.ValidationContext, config schema.ExtensionConfig, value interface{}) error {
	r.l.Lock()
	defer r.l.Unlock()

	if len(config.Mappings.Identity.Traits) > 0 {
		for _, t := range config.Mappings.Identity.Traits {
			res, err := sjson.SetBytes(r.i.Traits, t.Path, value)
			if err != nil {
				return ctx.Error("ory.sh\\/kratos/mappings/identity/traits", "unable to apply mapping: %s", err)
			}

			r.i.Traits = res
		}
	}
	return nil
}
