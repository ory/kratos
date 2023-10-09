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
)

const (
	extensionName string = "ory.sh/kratos"
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
		Mappings struct {
			Identity struct {
				Traits []struct {
					Path string `json:"path"`
				} `json:"traits"`
			} `json:"identity"`
		} `json:"mappings"`
	}

	Extension interface {
		Run(ctx jsonschema.ValidationContext, config ExtensionConfig, value interface{}) error
		Finish() error
	}

	ExtensionRunner struct {
		meta     *jsonschema.Schema
		compile  func(ctx jsonschema.CompilerContext, m map[string]interface{}) (interface{}, error)
		validate func(ctx jsonschema.ValidationContext, s interface{}, v interface{}) error

		runners []Extension
	}
)

func NewExtensionRunner(ctx context.Context, runners ...Extension) (*ExtensionRunner, error) {
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
		if raw, ok := m[extensionName]; ok {
			var b bytes.Buffer
			if err := json.NewEncoder(&b).Encode(raw); err != nil {
				return nil, errors.WithStack(err)
			}

			var e ExtensionConfig
			if err := json.NewDecoder(&b).Decode(&e); err != nil {
				return nil, errors.WithStack(err)
			}

			return &e, nil
		}
		return nil, nil
	}

	r.validate = func(ctx jsonschema.ValidationContext, s interface{}, v interface{}) error {
		c, ok := s.(*ExtensionConfig)
		if !ok {
			return nil
		}

		for _, runner := range r.runners {
			if err := runner.Run(ctx, *c, v); err != nil {
				return err
			}
		}
		return nil
	}

	r.runners = runners
	return r, nil
}

func (r *ExtensionRunner) Register(compiler *jsonschema.Compiler) *ExtensionRunner {
	compiler.Extensions[extensionName] = r.Extension()
	return r
}

func (r *ExtensionRunner) Extension() jsonschema.Extension {
	return jsonschema.Extension{
		Meta:     r.meta,
		Compile:  r.compile,
		Validate: r.validate,
	}
}

func (r *ExtensionRunner) AddRunner(run Extension) *ExtensionRunner {
	r.runners = append(r.runners, run)
	return r
}

func (r *ExtensionRunner) Finish() error {
	for _, runner := range r.runners {
		if err := runner.Finish(); err != nil {
			return err
		}
	}
	return nil
}
