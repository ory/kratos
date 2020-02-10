package profile

import (
	"bytes"
	"encoding/json"
	"io/ioutil"

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

type DisableIdentifiersExtension struct {
	disabled bool
}

func RegisterNewDisableIdentifiersExtension(compiler *jsonschema.Compiler) {
	if _, ok := compiler.Extensions[extensionName]; ok {
		// extension is already registered
		return
	}

	ext := jsonschema.Extension{
		// as we just use the content to check whether there the flag *.ory\.sh/kratos.credentials.password.identifier
		// set we don't really need a meta schema
		Meta:     nil,
		Compile:  compile,
		Validate: validate,
	}
	compiler.Extensions[extensionName] = ext
}

// this function adds the custom property IsDisabledField to the path so that we can use it to disable form fields etc.
func (ec *DisableIdentifiersExtension) EnhancePath(_ jsonschemax.Path) map[string]interface{} {
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

		schema, err := ioutil.ReadAll(&b)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		var e DisableIdentifiersExtension
		e.disabled = gjson.GetBytes(schema, "credentials.password.identifier").Bool()

		return &e, nil
	}
	return nil, nil
}

func validate(_ jsonschema.ValidationContext, _, _ interface{}) error {
	return nil
}
