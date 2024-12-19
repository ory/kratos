// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package identity

import (
	"fmt"
	"maps"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/ory/jsonschema/v3"
	"github.com/ory/kratos/schema"
)

func init() {
	jsonschema.Formats["no-validate"] = func(v interface{}) bool {
		return true
	}
}

type SchemaExtensionVerification struct {
	lifespan time.Duration
	l        sync.Mutex
	v        []VerifiableAddress
	i        *Identity
}

func NewSchemaExtensionVerification(i *Identity, lifespan time.Duration) *SchemaExtensionVerification {
	return &SchemaExtensionVerification{i: i, lifespan: lifespan}
}

const (
	ChannelTypeEmail = "email"
	ChannelTypeSMS   = "sms"
)

func (r *SchemaExtensionVerification) Run(ctx jsonschema.ValidationContext, s schema.ExtensionConfig, value interface{}) error {
	r.l.Lock()
	defer r.l.Unlock()

	if s.Verification.Via == "" {
		return nil
	}

	format, ok := s.RawSchema["format"]
	if !ok {
		format = ""
	}
	formatString, ok := format.(string)
	if !ok {
		return nil
	}

	if formatString == "" {
		switch s.Verification.Via {
		case ChannelTypeEmail:
			formatString = "email"
			formatter, ok := jsonschema.Formats[formatString]
			if !ok {
				supportedKeys := slices.Collect(maps.Keys(jsonschema.Formats))
				return ctx.Error("format", "format %q is not supported. Supported formats are [%s]", formatString, strings.Join(supportedKeys, ", "))
			}

			if !formatter(value) {
				return ctx.Error("format", "%q is not valid %q", value, formatString)
			}
		default:
			return ctx.Error("format", "no format specified. A format is required if verification is enabled. If this was intentional, please set \"format\" to \"no-validate\"")
		}
	}

	var normalized string
	switch formatString {
	case "email":
		normalized = strings.ToLower(strings.TrimSpace(fmt.Sprintf("%s", value)))
	default:
		normalized = strings.TrimSpace(fmt.Sprintf("%s", value))
	}

	address := NewVerifiableAddress(normalized, r.i.ID, s.Verification.Via)
	r.appendAddress(address)
	return nil
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
