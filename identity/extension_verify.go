// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package identity

import (
	"fmt"
	"strings"
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
	case AddressTypeEmail:
		if !jsonschema.Formats["email"](value) {
			return ctx.Error("format", "%q is not valid %q", value, "email")
		}

		address := NewVerifiableEmailAddress(
			strings.ToLower(strings.TrimSpace(
				fmt.Sprintf("%s", value))), r.i.ID)

		r.appendAddress(address)

		return nil

	case AddressTypePhone:
		if !jsonschema.Formats["tel"](value) {
			return ctx.Error("format", "%q is not valid %q", value, "phone")
		}

		address := NewVerifiablePhoneAddress(fmt.Sprintf("%s", value), r.i.ID)

		r.appendAddress(address)

		return nil

	case "":
		return nil
	}

	return ctx.Error("", "verification.via has unknown value %q", s.Verification.Via)
}

func (r *SchemaExtensionVerification) Finish() error {
	r.i.VerifiableAddresses = merge(r.v, r.i.VerifiableAddresses)
	return nil
}

// merge merges the base with the overrides through comparison with `has`. It changes the base slice in place.
func merge(base []VerifiableAddress, overrides []VerifiableAddress) []VerifiableAddress {
	for i := range base {
		if override := has(overrides, &base[i]); override != nil {
			base[i] = *override
		}
	}

	return base
}

func (r *SchemaExtensionVerification) appendAddress(address *VerifiableAddress) {
	if h := has(r.i.VerifiableAddresses, address); h != nil {
		if has(r.v, address) == nil {
			r.v = append(r.v, *h)
		}
		return
	}

	if has(r.v, address) == nil {
		r.v = append(r.v, *address)
	}
}

func has(haystack []VerifiableAddress, needle *VerifiableAddress) *VerifiableAddress {
	for _, has := range haystack {
		if has.Value == needle.Value && has.Via == needle.Via {
			return &has
		}
	}
	return nil
}
