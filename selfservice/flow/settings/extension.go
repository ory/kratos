package settings

import (
	"bytes"
	"encoding/json"

	"github.com/pkg/errors"
	"github.com/tidwall/gjson"

	"github.com/ory/kratos/selfservice/form"

	"github.com/ory/jsonschema/v3"
	"github.com/ory/x/jsonschemax"
)

const (
	extensionName = "disableIdentifiersExtension"
	extensionKey  = "ory.sh/kratos"
)

type disableIdentifiersExtConfig struct {
	disabled bool
}

func registerNewDisableIdentifiersExtension(compiler *jsonschema.Compiler) {
	compiler.Extensions[extensionName] = jsonschema.Extension{
		// as we just use the compiler to check whether there the flag *.ory\.sh/kratos.credentials.password.identifier
		// set we don't really need a meta schema
		Meta:    nil,
		Compile: compile,
	}
}

// this function adds the custom property IsDisabledField to the path so that we can use it to disable form fields etc.
func (ec *disableIdentifiersExtConfig) EnhancePath(_ jsonschemax.Path) map[string]interface{} {
	// if this path is an identifier the form field should be disabled
	if ec.disabled {
		return map[string]interface{}{
			form.DisableFormField: true,
		}
	}
	return nil
}

func compile(_ jsonschema.CompilerContext, m map[string]interface{}) (interface{}, error) {
	if raw, ok := m[extensionKey]; ok {
		var b bytes.Buffer
		if err := json.NewEncoder(&b).Encode(raw); err != nil {
			return nil, errors.WithStack(err)
		}

		var e disableIdentifiersExtConfig
		e.disabled = gjson.GetBytes(b.Bytes(), "credentials.password.identifier").Bool()

		return &e, nil
	}
	return nil, nil
}
