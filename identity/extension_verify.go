package identity

import (
	"fmt"
	"sync"
	"time"

	"github.com/ory/jsonschema/v3"

	"github.com/ory/kratos/schema"
)

type SchemaExtensionVerification struct {
	lifespan time.Duration
	l        sync.Mutex
	v        []VerifiableAddress
	i        *Identity
}

func NewSchemaExtensionVerification(i *Identity, lifespan time.Duration) *SchemaExtensionVerification {
	return &SchemaExtensionVerification{i: i, lifespan: lifespan}
}

func (r *SchemaExtensionVerification) Run(ctx jsonschema.ValidationContext, s schema.ExtensionConfig, value interface{}) error {
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

		if has := r.has(r.i.VerifiableAddresses, address); has != nil {
			if r.has(r.v, address) == nil {
				r.v = append(r.v, *has)
			}
			return nil
		}

		if has := r.has(r.v, address); has == nil {
			r.v = append(r.v, *address)
		}

		return nil
	case "":
		return nil
	}

	return ctx.Error("", "verification.via has unknown value %q", s.Verification.Via)
}

func (r *SchemaExtensionVerification) has(haystack []VerifiableAddress, needle *VerifiableAddress) *VerifiableAddress {
	for _, has := range haystack {
		if has.Value == needle.Value && has.Via == needle.Via {
			return &has
		}
	}
	return nil
}

func (r *SchemaExtensionVerification) Finish() error {
	r.i.VerifiableAddresses = r.v
	return nil
}
