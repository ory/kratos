package schema

import (
	"bytes"
	"encoding/json"
	"sync"
	"unicode"

	"github.com/pkg/errors"

	"github.com/ory/gojsonschema"
	"github.com/ory/x/jsonx"

	"github.com/ory/herodot"
)

type Validator struct {
	sync.RWMutex
	// schemas map[string]*gojsonschema.Schema
}

type ResultErrors []gojsonschema.ResultError

func (e ResultErrors) Error() string {
	if len(e) > 0 {
		return lowerCaseFirst(e[0].Description())
	}
	panic("no errors available")
}

func lowerCaseFirst(str string) string {
	for i, v := range str {
		return string(unicode.ToLower(v)) + str[i+1:]
	}
	return ""
}

type ValidationProvider interface {
	SchemaValidator() *Validator
}

type Extension struct {
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

type ValidationExtender interface {
	Call(node interface{}, extension *Extension, context *gojsonschema.JsonContext) error
}

func NewValidator() *Validator {
	return &Validator{
		// schemas: make(map[string]*gojsonschema.Schema),
	}
}

func (v *Validator) schema(source string) (schema *gojsonschema.Schema, err error) {
	// v.RLock()
	// schema, ok := v.schemas[source]
	// v.RUnlock()
	//
	// if ok {
	// 	return schema, nil
	// }

	schema, err = gojsonschema.NewSchema(gojsonschema.NewReferenceLoader(source))
	if err != nil {
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("Unable to parse JSON schema: %s", err))
	}

	// v.Lock()
	// v.schemas[source] = schema
	// v.Unlock()

	return schema, nil
}

func (v *Validator) hook(h ValidationExtender) func(schema *gojsonschema.SubSchema, node interface{}, _ *gojsonschema.Result, context *gojsonschema.JsonContext) error {
	return func(schema *gojsonschema.SubSchema, node interface{}, _ *gojsonschema.Result, context *gojsonschema.JsonContext) error {
		var b bytes.Buffer

		m, ok := schema.Node().(map[string]interface{})
		if !ok {
			return nil
		}

		raw, ok := m["hive"]
		if !ok {
			return nil
		}

		if err := json.NewEncoder(&b).Encode(raw); err != nil {
			return errors.WithStack(err)
		}

		var meta Extension
		if err := jsonx.NewStrictDecoder(&b).Decode(&meta); err != nil {
			return errors.WithStack(err)
		}

		return h.Call(node, &meta, context)
	}
}

func (v *Validator) hookify(hs []ValidationExtender) (hookified []gojsonschema.Hook) {
	hookified = make([]gojsonschema.Hook, len(hs))
	for k, h := range hs {
		hookified[k] = v.hook(h)
	}

	return hookified
}

func (v *Validator) Validate(
	href string,
	object gojsonschema.JSONLoader,
	extensions ...ValidationExtender,
) error {
	schema, err := v.schema(href)
	if err != nil {
		return err
	}

	schema.SetHooks(v.hookify(extensions))

	r, err := schema.
		Validate(object)
	if err != nil {
		return errors.WithStack(herodot.ErrInternalServerError.WithReasonf("Unable to parse validate JSON object against JSON schema: %s", err))
	}

	if !r.Valid() {
		return errors.WithStack(ResultErrors(r.Errors()))
	}
	return nil
}
