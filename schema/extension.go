// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package schema

import (
	"bytes"
	"context"
	"encoding/json"

	"github.com/pkg/errors"

	"github.com/ory/jsonschema/v3"
	"github.com/ory/kratos/embedx"
	"github.com/ory/x/jsonschemax"
)

const (
	ExtensionName string = "ory.sh/kratos"
)

type (
	ExtensionConfig struct {
		Credentials struct {
			Password struct {
				Identifier bool `json:"identifier"`
			} `json:"password"`
			WebAuthn struct {
				Identifier bool `json:"identifier"`
			} `json:"webauthn"`
			Passkey struct {
				DisplayName bool `json:"display_name"`
			} `json:"passkey"`
			TOTP struct {
				AccountName bool `json:"account_name"`
			} `json:"totp"`
			Code struct {
				Identifier bool   `json:"identifier"`
				Via        string `json:"via"`
			} `json:"code"`
		} `json:"credentials"`
		Verification struct {
			Via string `json:"via"`
		} `json:"verification"`
		Recovery struct {
			Via string `json:"via"`
		} `json:"recovery"`
		Organization struct {
			Matcher string `json:"matcher"`
		} `json:"organizations"`
		RawSchema map[string]interface{} `json:"-"`
	}

	ValidateExtension interface {
		Run(ctx jsonschema.ValidationContext, config ExtensionConfig, value interface{}) error
		Finish() error
	}
	CompileExtension interface {
		Run(ctx jsonschema.CompilerContext, config ExtensionConfig, rawSchema map[string]interface{}) error
	}

	ExtensionRunner struct {
		meta     *jsonschema.Schema
		compile  func(ctx jsonschema.CompilerContext, rawSchema map[string]interface{}) (interface{}, error)
		validate func(ctx jsonschema.ValidationContext, s interface{}, v interface{}) error

		validateRunners []ValidateExtension
		compileRunners  []CompileExtension
	}

	ExtensionRunnerOption func(*ExtensionRunner)
)

func (e *ExtensionConfig) EnhancePath(path jsonschemax.Path) map[string]any {
	props := path.CustomProperties
	if props == nil {
		props = make(map[string]any)
	}
	props[ExtensionName] = e

	return props
}

func WithValidateRunners(runners ...ValidateExtension) ExtensionRunnerOption {
	return func(r *ExtensionRunner) {
		r.validateRunners = append(r.validateRunners, runners...)
	}
}

func WithCompileRunners(runners ...CompileExtension) ExtensionRunnerOption {
	return func(r *ExtensionRunner) {
		r.compileRunners = append(r.compileRunners, runners...)
	}
}

func NewExtensionRunner(ctx context.Context, opts ...ExtensionRunnerOption) (*ExtensionRunner, error) {
	var err error
	r := new(ExtensionRunner)
	c := jsonschema.NewCompiler()

	if err = embedx.AddSchemaResources(c, embedx.IdentityExtension); err != nil {
		return nil, errors.WithStack(err)
	}

	r.meta, err = c.Compile(ctx, embedx.IdentityExtension.GetSchemaID())
	if err != nil {
		return nil, errors.WithStack(err)
	}

	r.compile = func(ctx jsonschema.CompilerContext, m map[string]interface{}) (interface{}, error) {
		if raw, ok := m[ExtensionName]; ok {
			var b bytes.Buffer
			if err := json.NewEncoder(&b).Encode(raw); err != nil {
				return nil, errors.WithStack(err)
			}

			var e ExtensionConfig
			if err := json.NewDecoder(&b).Decode(&e); err != nil {
				return nil, errors.WithStack(err)
			}

			for _, runner := range r.compileRunners {
				if err := runner.Run(ctx, e, m); err != nil {
					return nil, err
				}
			}
			e.RawSchema = m

			return &e, nil
		}
		return nil, nil
	}

	r.validate = func(ctx jsonschema.ValidationContext, s interface{}, v interface{}) error {
		c, ok := s.(*ExtensionConfig)
		if !ok {
			return nil
		}

		for _, runner := range r.validateRunners {
			if err := runner.Run(ctx, *c, v); err != nil {
				return err
			}
		}
		return nil
	}

	for _, opt := range opts {
		opt(r)
	}

	return r, nil
}

func (r *ExtensionRunner) Register(compiler *jsonschema.Compiler) *ExtensionRunner {
	compiler.Extensions[ExtensionName] = r.Extension()
	return r
}

func (r *ExtensionRunner) Extension() jsonschema.Extension {
	return jsonschema.Extension{
		Meta:     r.meta,
		Compile:  r.compile,
		Validate: r.validate,
	}
}

func (r *ExtensionRunner) AddRunner(run ValidateExtension) *ExtensionRunner {
	r.validateRunners = append(r.validateRunners, run)
	return r
}

func (r *ExtensionRunner) Finish() error {
	for _, runner := range r.validateRunners {
		if err := runner.Finish(); err != nil {
			return err
		}
	}
	return nil
}
