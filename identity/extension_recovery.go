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
	"github.com/ory/kratos/x"
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

		address = NewRecoveryEmailAddress(x.NormalizeEmailIdentifier(fmt.Sprintf("%s", value)), r.i.ID)

	case "sms":
		normalized, err := x.NormalizeIdentifier(fmt.Sprintf("%s", value), s.Recovery.Via)
		if err != nil {
			return ctx.Error("format", "%q is not valid \"tel\" for %q", value, s.Recovery.Via)
		}

		address = NewRecoverySMSAddress(normalized, r.i.ID)

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
	// Normalize both sides so pre-normalization persisted values still match.
	normalizedNeedle := x.GracefulNormalization(needle.Value)
	for _, has := range haystack {
		if has.Via == needle.Via && x.GracefulNormalization(has.Value) == normalizedNeedle {
			return &has
		}
	}
	return nil
}

func (r *SchemaExtensionRecovery) Finish() error {
	r.i.RecoveryAddresses = r.v
	return nil
}
