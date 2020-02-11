package verify

import (
	"fmt"
	"sync"
	"time"

	"github.com/ory/jsonschema/v3"

	"github.com/ory/kratos/driver/configuration"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/schema"
)

type ValidationExtensionRunner struct {
	c        configuration.Provider
	as       []Address
	lifespan time.Duration
	l        sync.Mutex
	i        *identity.Identity
}

func NewValidationExtensionRunner(i *identity.Identity, lifespan time.Duration) *ValidationExtensionRunner {
	return &ValidationExtensionRunner{i: i, lifespan: lifespan}
}

func (r *ValidationExtensionRunner) Runner(ctx jsonschema.ValidationContext, s schema.ExtensionConfig, value interface{}) error {
	r.l.Lock()
	defer r.l.Unlock()

	switch s.Verification.Via {
	case "email":
		if !jsonschema.Formats["email"](value) {
			return ctx.Error("format", "%q is not valid %q", value, "email")
		}

		address, err := NewEmailAddress(fmt.Sprintf("%s", value), r.i.ID, r.lifespan)
		if err != nil {
			return err
		}

		r.as = append(r.as, *address)
	default:
		return nil
	}

	return nil
}

func (r *ValidationExtensionRunner) Addresses() []Address {
	return r.as
}
