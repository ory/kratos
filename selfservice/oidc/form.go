package oidc

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/imdario/mergo"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"

	"github.com/ory/hive/selfservice"
)

func toMap(j json.RawMessage) map[string]interface{} {
	dest := make(map[string]interface{})

	p := gjson.ParseBytes(j)
	if p.IsObject() {
		for k, v := range p.Map() {
			dest[k] = flatten([]byte(v.Raw))
		}
	} else if p.IsArray() {
		for k, v := range p.Array() {
			dest[fmt.Sprintf("%d", k)] = flatten([]byte(v.Raw))
		}
	} else {
		dest[""] = p.Value()
	}

	return dest
}

func flattenKeys(j json.RawMessage, keys map[string]interface{}, dest map[string]interface{}, prefix string) {
	for k, v := range keys {
		path := prefix
		if k != "" {
			if path != "" {
				path = path + "." + k
			} else {
				path = k
			}
		}

		if inner, ok := v.(map[string]interface{}); ok {
			flattenKeys(j, inner, dest, path)
		} else {
			dest[path] = gjson.GetBytes(j, path).Value()
		}
	}
}

func flatten(j json.RawMessage) map[string]interface{} {
	result := make(map[string]interface{})
	flattenKeys(j, toMap(j), result, "")

	return result
}

// toFormValues converts a json object with nested fields to a url.Values representation, for example:
//
// - { "foo": [{ "bar": "baz" }] } -> url.Values{"foo[0].bar": {"baz"}}
func toFormValues(traits json.RawMessage) url.Values {
	result := url.Values{}

	for k := range flatten(traits) {
		result.Set(k, gjson.GetBytes(traits, k).String())
	}

	return result
}

// merge merges form values prefixed with `traits` (encoded) with a JSON raw message.
func merge(dc *selfservice.BodyDecoder, form string, traits json.RawMessage) (json.RawMessage, error) {
	if form == "" {
		return traits, nil
	}

	q, err := url.ParseQuery(form)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	var decodedForm struct {
		Traits map[string]interface{} `json:"traits"`
	}
	if err := dc.DecodeForm(q, &decodedForm); err != nil {
		return nil, err
	}

	var decodedTraits map[string]interface{}
	if err := json.NewDecoder(bytes.NewBuffer(traits)).Decode(&decodedTraits); err != nil {
		return nil, err
	}

	if err := mergo.Merge(&decodedTraits, decodedForm.Traits, mergo.WithOverride); err != nil {
		return nil, err
	}

	var result bytes.Buffer
	if err := json.NewEncoder(&result).Encode(decodedTraits); err != nil {
		return nil, err
	}

	return result.Bytes(), nil
}
