package schema

import (
	"bytes"
	"encoding/json"
	"github.com/ory/x/pkgerx"
	"path"

	"github.com/markbates/pkger"
	"github.com/pkg/errors"

	"github.com/ory/jsonschema/v3"
)

var schemas = pkger.Dir("/schema/.schema")

const (
	ExtensionRunnerIdentityMetaSchema ExtensionRunnerMetaSchema = "extension/identity.schema.json"
	extensionName                     string                    = "ory.sh/kratos"
)

type (
	ExtensionRunnerMetaSchema string
	ExtensionConfig           struct {
		Credentials struct {
			Password struct {
				Identifier bool `json:"identifier"`
			} `json:"password"`
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

func NewExtensionRunner(meta ExtensionRunnerMetaSchema, runners ...Extension) (*ExtensionRunner, error) {
	var err error
	schema, err := pkgerx.Read(pkger.Open(path.Join(string(schemas), string(meta))))
	if err != nil {
		return nil, errors.WithStack(err)
	}

	r := new(ExtensionRunner)
	r.meta, err = jsonschema.CompileString(string(meta), string(schema))
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
