// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package identity

import (
	"context"
	"maps"
	"slices"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"

	"github.com/ory/herodot"
	"github.com/ory/jsonschema/v3"
	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/schema"
	"github.com/ory/kratos/x"
	"github.com/ory/x/otelx"
)

type (
	validatorDependencies interface {
		schema.IdentitySchemaProvider
		config.Provider
	}
	Validator struct {
		v *schema.Validator
		d validatorDependencies
	}
	ValidationProvider interface {
		IdentityValidator() *Validator
	}
)

func NewValidator(d validatorDependencies) *Validator {
	return &Validator{v: schema.NewValidator(), d: d}
}

func (v *Validator) ValidateWithRunner(ctx context.Context, i *Identity, runners ...schema.ValidateExtension) error {
	runner, err := schema.NewExtensionRunner(ctx, schema.WithValidateRunners(runners...))
	if err != nil {
		return err
	}

	ss, err := v.d.IdentityTraitsSchemas(ctx)
	if err != nil {
		return err
	}

	s, err := ss.GetByID(i.SchemaID)
	if err != nil {
		return err
	}

	if len(i.Traits) == 0 {
		i.Traits = []byte(`{}`)
	}

	if err := v.normalizePhoneTraits(ctx, i, s.URL.String()); err != nil {
		return err
	}

	traits, err := sjson.SetRawBytes([]byte(`{}`), "traits", i.Traits)
	if err != nil {
		return errors.WithStack(herodot.ErrBadRequest().WithError(err.Error()))
	}

	return v.v.Validate(ctx, s.URL.String(), traits,
		schema.WithExtensionRunner(runner),
		schema.WithDisallowRefs(v.d.Config().SecurityDisallowRefInIdentitySchemas(ctx)),
	)
}

// normalizePhoneTraits rewrites trait values used as phone-channel
// identifiers (code+sms credential identifiers, recovery via sms, or
// verification via sms) into the E.164 form Kratos stores in its side
// tables. Without this, webhook payloads templated against
// `identity.traits` see the raw user input while Kratos keys on the
// normalized identifier, producing inconsistent values across systems.
//
// Email-channel identifiers preserve case per the regression test for
// https://github.com/ory/kratos/issues/3187.
//
// Schema-load and compile errors are swallowed; the subsequent
// `v.v.Validate` call uses the same loader and compiler wiring and
// re-surfaces the same error with full herodot context. Context
// cancellation, however, must propagate so callers see the cancellation
// instead of a generic schema error from the second compile.
//
// WORKAROUND: this compiles the schema a second time and walks it from
// scratch because github.com/ory/jsonschema/v3 does not expose the JSON
// pointer of the value being validated to extension hooks. Once we move
// to santhosh-tekuri/jsonschema/v6, which provides
// ValidatorContext.ValueLocation(), this and walkPhoneTraits collapse
// into a single ValidateExtension that writes back to i.Traits inline.
// Tracking: ~/.claude/docs/cloud/plans/jsonschema-v6-migration.md.
func (v *Validator) normalizePhoneTraits(ctx context.Context, i *Identity, schemaURL string) error {
	swallow := func(err error) error {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return err
		}
		return nil
	}

	compiler, err := schema.NewCompilerWithURL(ctx, schemaURL,
		v.d.Config().SecurityDisallowRefInIdentitySchemas(ctx))
	if err != nil {
		return swallow(err)
	}
	// Register populates each compiled *Schema's Extensions map with the
	// parsed *ExtensionConfig that walkPhoneTraits reads. Without this
	// call, the walk sees nil extensions and never normalizes.
	runner, err := schema.NewExtensionRunner(ctx)
	if err != nil {
		return swallow(err)
	}
	runner.Register(compiler)

	compiled, err := compiler.Compile(ctx, schemaURL)
	if err != nil {
		return swallow(err)
	}

	i.Traits = walkPhoneTraits(i.Traits, compiled.Properties["traits"], "", map[*jsonschema.Schema]struct{}{})
	return nil
}

// walkPhoneTraits traverses the compiled schema tree and rewrites every
// trait leaf marked as a phone-channel identifier to its E.164 form.
//
// Properties extend the path with the property name; array items extend
// it with the index. Combinators (allOf, anyOf, oneOf, if/then/else,
// not, $ref) sit at the same path. PatternProperties and contains are
// not walked: identity schemas in practice keep credential identifiers
// on named or indexed positions.
//
// `stack` holds the schemas currently on the recursion path so true
// cycles short-circuit without preventing legitimate DAG reuse —
// e.g. two trait properties that share a `$ref` definition each get
// walked because the shared node is removed from the stack when the
// first walk returns. Pure-$ref cycles are already rejected by
// schema/prevalidate.go, but cycles with intermediate validation
// (a $ref back to an ancestor that has `properties`) reach the walk
// and need this guard.
func walkPhoneTraits(traits Traits, node *jsonschema.Schema, path string, stack map[*jsonschema.Schema]struct{}) Traits {
	if node == nil {
		return traits
	}
	if _, onStack := stack[node]; onStack {
		return traits
	}
	stack[node] = struct{}{}
	defer delete(stack, node)

	if isPhoneIdentifier(node) {
		return normalizePhoneAt(traits, path)
	}

	// Combinators evaluated at the same path.
	for _, sub := range slices.Concat(node.AllOf, node.AnyOf, node.OneOf,
		[]*jsonschema.Schema{node.Not, node.If, node.Then, node.Else, node.Ref}) {
		traits = walkPhoneTraits(traits, sub, path, stack)
	}

	// Properties extend the path. Iterate in sorted order so the
	// rewrite sequence is deterministic and tests are reproducible.
	for _, name := range slices.Sorted(maps.Keys(node.Properties)) {
		next := escapePathSegment(name)
		if path != "" {
			next = path + "." + next
		}
		traits = walkPhoneTraits(traits, node.Properties[name], next, stack)
	}

	// Array items. `Items` is nil, *Schema (homogeneous — every element
	// validates against the same schema), or []*Schema (tuple — index N
	// validates against the Nth schema). For homogeneous items we walk
	// against the actual data because schema length is unbounded; for
	// tuple items the schema length defines the slots to walk.
	switch items := node.Items.(type) {
	case *jsonschema.Schema:
		index := 0
		gjson.GetBytes(traits, path).ForEach(func(_, _ gjson.Result) bool {
			next := strconv.Itoa(index)
			if path != "" {
				next = path + "." + next
			}
			traits = walkPhoneTraits(traits, items, next, stack)
			index++
			return true
		})
	case []*jsonschema.Schema:
		for i, item := range items {
			next := strconv.Itoa(i)
			if path != "" {
				next = path + "." + next
			}
			traits = walkPhoneTraits(traits, item, next, stack)
		}
	}
	return traits
}

// normalizePhoneAt rewrites the string trait at `path` to its E.164
// form when present and parseable. It returns the input unchanged on
// any miss so the caller can chain it through a walk.
func normalizePhoneAt(traits Traits, path string) Traits {
	value := gjson.GetBytes(traits, path)
	if value.Type != gjson.String {
		return traits
	}
	normalized, err := x.NormalizeIdentifier(value.String(), "sms")
	if err != nil || normalized == value.String() {
		return traits
	}
	if updated, err := sjson.SetBytes(traits, path, normalized); err == nil {
		return updated
	}
	return traits
}

// escapePathSegment escapes characters that gjson/sjson treat as path
// metacharacters so a property like "phone.number" is addressed as the
// single key "phone.number" rather than as nested `phone` → `number`.
func escapePathSegment(segment string) string {
	if !strings.ContainsAny(segment, `.*?\`) {
		return segment
	}
	var b strings.Builder
	b.Grow(len(segment) + 4)
	for _, r := range segment {
		switch r {
		case '.', '*', '?', '\\':
			b.WriteByte('\\')
		}
		b.WriteRune(r)
	}
	return b.String()
}

// isPhoneIdentifier reports whether any identity-extension hook would
// normalize the value at this schema node as an E.164 phone number.
func isPhoneIdentifier(node *jsonschema.Schema) bool {
	cfg, _ := node.Extensions[schema.ExtensionName].(*schema.ExtensionConfig)
	if cfg == nil {
		return false
	}
	return (cfg.Credentials.Code.Identifier && cfg.Credentials.Code.Via == "sms") ||
		cfg.Recovery.Via == "sms" ||
		cfg.Verification.Via == "sms"
}

func (v *Validator) Validate(ctx context.Context, i *Identity) error {
	return otelx.WithSpan(ctx, "identity.Validator.Validate", func(ctx context.Context) error {
		return v.ValidateWithRunner(ctx, i,
			NewSchemaExtensionCredentials(i),
			NewSchemaExtensionVerification(i, v.d.Config().SelfServiceFlowVerificationRequestLifespan(ctx)),
			NewSchemaExtensionRecovery(i),
		)
	})
}
