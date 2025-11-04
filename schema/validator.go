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
	e *ExtensionRunner
}

func WithExtensionRunner(e *ExtensionRunner) func(*validatorOptions) {
	return func(o *validatorOptions) {
		o.e = e
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

	compiler := jsonschema.NewCompiler()
	resource, err := jsonschema.LoadURL(ctx, href)
	if err != nil {
		return errors.WithStack(herodot.ErrMisconfiguration.WithReasonf("Unable to load or parse the JSON schema.").WithWrap(err).WithDebugf("%s", err))
	}

	if o.e != nil {
		o.e.Register(compiler)
	}

	if err := compiler.AddResource(href, resource); err != nil {
		return errors.WithStack(herodot.ErrMisconfiguration.WithReasonf("Unable to parse validate JSON object against JSON schema.").WithWrap(err).WithDebugf("%s", err))
	}

	schema, err := compiler.Compile(ctx, href)
	if err != nil {
		return errors.WithStack(herodot.ErrMisconfiguration.WithReasonf("Unable to parse validate JSON object against JSON schema.").WithWrap(err).WithDebugf("%s", err))
	}

	// we decode explicitly here, so we can handle the error, and it is not lost in the schema validation
	dec, err := jsonschema.DecodeJSON(bytes.NewBuffer(document))
	if err != nil {
		return errors.WithStack(herodot.ErrBadRequest.WithReasonf("Unable to parse validate JSON object against JSON schema.").WithWrap(err).WithDebugf("%s", err))
	}
	if err := schema.ValidateInterface(dec); err != nil {
		return errors.WithStack(err)
	}

	if o.e != nil {
		return o.e.Finish()
	}

	return nil
}
