// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package schema

import (
	"bytes"
	"context"
	"encoding/json"
	"sync"

	"github.com/pkg/errors"

	"github.com/ory/herodot"

	"github.com/ory/jsonschema/v3"
)

type Validator struct {
	sync.RWMutex
}

type ValidationProvider interface {
	SchemaValidator() *Validator
}

func NewValidator() *Validator {
	return &Validator{}
}

type validatorOptions struct {
	e            *ExtensionRunner
	disallowRefs bool
}

func WithExtensionRunner(e *ExtensionRunner) func(*validatorOptions) {
	return func(o *validatorOptions) {
		o.e = e
	}
}

// WithDisallowRefs toggles the restriction on `$ref` URL schemes. When true,
// the compiler rejects `$ref` values whose scheme is not `http`, `https`, or
// `base64`, preventing local-file reads via `file://`. Callers should forward
// the `security.disallow_ref_in_identity_schemas` config value here.
func WithDisallowRefs(disallow bool) func(*validatorOptions) {
	return func(o *validatorOptions) {
		o.disallowRefs = disallow
	}
}

func (v *Validator) Validate(
	ctx context.Context,
	href string,
	document json.RawMessage,
	opts ...func(*validatorOptions),
) error {
	var o validatorOptions
	for _, opt := range opts {
		opt(&o)
	}

	compiler, err := NewCompilerWithURL(ctx, href, o.disallowRefs)
	if err != nil {
		return errors.WithStack(herodot.ErrMisconfiguration().WithReasonf("Unable to load or parse the JSON schema.").WithWrap(err).WithDebugf("%s", err))
	}

	if o.e != nil {
		o.e.Register(compiler)
	}

	schema, err := compiler.Compile(ctx, href)
	if err != nil {
		return errors.WithStack(herodot.ErrMisconfiguration().WithReasonf("Unable to parse validate JSON object against JSON schema.").WithWrap(err).WithDebugf("%s", err))
	}

	// we decode explicitly here, so we can handle the error, and it is not lost in the schema validation
	dec, err := jsonschema.DecodeJSON(bytes.NewBuffer(document))
	if err != nil {
		return errors.WithStack(herodot.ErrBadRequest().WithReasonf("Unable to parse validate JSON object against JSON schema.").WithWrap(err).WithDebugf("%s", err))
	}
	if err := schema.ValidateInterface(dec); err != nil {
		return errors.WithStack(err)
	}

	if o.e != nil {
		return o.e.Finish()
	}

	return nil
}
