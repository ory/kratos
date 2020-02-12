package identity

import (
	"fmt"
	"sync"
	"time"

	"github.com/ory/jsonschema/v3"

	"github.com/ory/kratos/schema"
)

type SchemaExtensionVerify struct {
	as       []VerifiableAddress
	lifespan time.Duration
	l        sync.Mutex
	i        *Identity
}

func NewSchemaExtensionVerify(i *Identity, lifespan time.Duration) *SchemaExtensionVerify {
	return &SchemaExtensionVerify{i: i, lifespan: lifespan}
}

func (r *SchemaExtensionVerify) Runner(ctx jsonschema.ValidationContext, s schema.ExtensionConfig, value interface{}) error {
	r.l.Lock()
	defer r.l.Unlock()

	switch s.Verification.Via {
	case "email":
		if !jsonschema.Formats["email"](value) {
			return ctx.Error("format", "%q is not valid %q", value, "email")
		}

		address, err := NewVerifiableEmailAddress(fmt.Sprintf("%s", value), r.i.ID, r.lifespan)
		if err != nil {
			return err
		}

		r.as = append(r.as, *address)
	default:
		return nil
	}

	return nil
}

func (r *SchemaExtensionVerify) Addresses() []VerifiableAddress {
	return r.as
}
