// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package identity

import (
	"fmt"
	"maps"
	"slices"
	"strings"
	"sync"

	"github.com/ory/jsonschema/v3"

	"github.com/ory/kratos/schema"
)

type SchemaExtensionRecovery struct {
	l sync.Mutex
	v []RecoveryAddress
	i *Identity
}

func NewSchemaExtensionRecovery(i *Identity) *SchemaExtensionRecovery {
	return &SchemaExtensionRecovery{i: i}
}

func (r *SchemaExtensionRecovery) Run(ctx jsonschema.ValidationContext, s schema.ExtensionConfig, value interface{}) error {
	r.l.Lock()
	defer r.l.Unlock()

	var address *RecoveryAddress
	switch s.Recovery.Via {
	case "email":
		formatString := "email"
		formatter, ok := jsonschema.Formats[formatString]
		if !ok {
			supportedKeys := slices.Collect(maps.Keys(jsonschema.Formats))
			return ctx.Error("format", "format %q is not supported. Supported formats are [%s]", formatString, strings.Join(supportedKeys, ", "))
		}

		if !formatter(value) {
			return ctx.Error("format", "%q is not valid %q", value, formatString)
		}

		address = NewRecoveryEmailAddress(
			strings.ToLower(strings.TrimSpace(
				fmt.Sprintf("%s", value))), r.i.ID)

	case "sms":
		formatString := "tel"
		formatter, ok := jsonschema.Formats[formatString]
		if !ok {
			supportedKeys := slices.Collect(maps.Keys(jsonschema.Formats))
			return ctx.Error("format", "format %q is not supported. Supported formats are [%s]", formatString, strings.Join(supportedKeys, ", "))
		}

		if !formatter(value) {
			return ctx.Error("format", "%q is not valid %q", value, formatString)
		}

		address = NewRecoverySMSAddress(
			strings.TrimSpace(
				fmt.Sprintf("%s", value)), r.i.ID)

	case "":
		return nil
	default:
		return ctx.Error("", "recovery.via has unknown value %q", s.Recovery.Via)
	}

	if has := r.has(r.i.RecoveryAddresses, address); has != nil {
		if r.has(r.v, address) == nil {
			r.v = append(r.v, *has)
		}
		return nil
	}

	if has := r.has(r.v, address); has == nil {
		r.v = append(r.v, *address)
	}

	return nil
}

func (r *SchemaExtensionRecovery) has(haystack []RecoveryAddress, needle *RecoveryAddress) *RecoveryAddress {
	for _, has := range haystack {
		if has.Value == needle.Value && has.Via == needle.Via {
			return &has
		}
	}
	return nil
}

func (r *SchemaExtensionRecovery) Finish() error {
	r.i.RecoveryAddresses = r.v
	return nil
}
