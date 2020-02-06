package schema

import (
	"bytes"
	"encoding/json"

	"github.com/gobuffalo/packr/v2"
	"github.com/pkg/errors"

	"github.com/ory/jsonschema/v3"
)

var box = packr.New("contrib", "contrib")

const (
	ExtensionRunnerIdentityMetaSchema ExtensionRunnerMetaSchema = "extension/identity.schema.json"
	ExtensionRunnerOIDCMetaSchema     ExtensionRunnerMetaSchema = "extension/oidc.schema.json"
	extensionName                                               = "ory.sh/kratos"
)

type (
	ExtensionRunnerMetaSchema string
	ExtensionConfig           struct {
		Credentials struct {
			Password struct {
				Identifier bool `json:"identifier"`
			} `json:"password"`
		} `json:"credentials"`
		Mappings struct {
			Identity struct {
				Traits []struct {
					Path string `json:"path"`
				} `json:"traits"`
			} `json:"identity"`
		} `json:"mappings"`
	}
	runner          func(ctx jsonschema.ValidationContext, config ExtensionConfig, value interface{}) error
	ExtensionRunner struct {
		meta     *jsonschema.Schema
		compile  func(ctx jsonschema.CompilerContext, m map[string]interface{}) (interface{}, error)
		validate func(ctx jsonschema.ValidationContext, s interface{}, v interface{}) error

		runners []runner
	}
)

func NewExtensionRunner(meta ExtensionRunnerMetaSchema, runners ...runner) (*ExtensionRunner, error) {
	var err error
	schema, err := box.FindString(string(meta))
	if err != nil {
		return nil, errors.WithStack(err)
	}

	r := new(ExtensionRunner)
	r.meta, err = jsonschema.CompileString(string(meta), schema)
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
			if err := runner(ctx, *c, v); err != nil {
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

func (r *ExtensionRunner) AddRunner(run runner) *ExtensionRunner {
	r.runners = append(r.runners, run)
	return r
}
