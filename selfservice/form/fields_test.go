package form

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	"github.com/ory/jsonschema/v3"

	"github.com/ory/x/jsonschemax"
)

func TestFieldFromPath(t *testing.T) {
	t.Run("all properties are properly transfered", func(t *testing.T) {
		schema, err := ioutil.ReadFile("./stub/all_formats.schema.json")
		require.NoError(t, err)

		c := jsonschema.NewCompiler()
		require.NoError(t, c.AddResource("test.json", bytes.NewBuffer(schema)))

		paths, err := jsonschemax.ListPaths("test.json", c)
		require.NoError(t, err)

		for _, path := range paths {
			htmlField := fieldFromPath(path.Name, path)
			assert.Equal(t, gjson.GetBytes(schema, fmt.Sprintf("properties.%s.test_expected_type", path.Name)).String(), htmlField.Type)
			assert.True(t, !gjson.GetBytes(schema, fmt.Sprintf("properties.%s.test_expected_pattern", path.Name)).Exists() || (gjson.GetBytes(schema, fmt.Sprintf("properties.%s.test_expected_pattern", path.Name)).Bool() && htmlField.Pattern != ""))
			fmt.Printf("name %s\ntype %s\n", htmlField.Name, htmlField.Type)
		}
	})
}

//
// func TestNewFormFieldsFromJSON(t *testing.T) {
// 	var js = json.RawMessage(`{"numby":1.5,"stringy":"foobar","objy":{"objy":{},"numby":1.5,"stringy":"foobar"}}`)
//
// 	assert.EqualValues(t, Fields{
// 		"traits.numby":        Field{Name: "traits.numby", Type: "number"},
// 		"traits.objy.numby":   Field{Name: "traits.objy.numby", Type: "number"},
// 		"traits.objy.objy":    Field{Name: "traits.objy.objy", Type: "text"},
// 		"traits.objy.stringy": Field{Name: "traits.objy.stringy", Type: "text"},
// 		"traits.stringy":      Field{Name: "traits.stringy", Type: "text"},
// 		CSRFTokenName:         Field{Name: CSRFTokenName, Type: "text", Value: "foo_token"},
// 	}, NewFormFieldsFromJSON(js, "traits", "foo_token"))
//
// 	assert.EqualValues(t, Fields{
// 		"numby":        Field{Name: "numby", Type: "number", Value:},
// 		"objy.numby":   Field{Name: "objy.numby", Type: "number"},
// 		"objy.objy":    Field{Name: "objy.objy", Type: "text"},
// 		"objy.stringy": Field{Name: "objy.stringy", Type: "text"},
// 		"stringy":      Field{Name: "stringy", Type: "text"},
// 		CSRFTokenName:  Field{Name: CSRFTokenName, Type: "text", Value: "foo_token"},
// 	}, NewFormFieldsFromJSON(js, "", "foo_token"))
// }
//
// func TestNewFormFieldsFromJSONSchema(t *testing.T) {
// 	for k, tc := range []struct {
// 		f        string
// 		prefix   string
// 		expected Fields
// 	}{
// 		{
// 			f:      "../stub/new-form.json",
// 			prefix: "",
// 			expected: Fields{
// 				"numby":        Field{Name: "numby", Type: "number"},
// 				"objy.numby":   Field{Name: "objy.numby", Type: "number"},
// 				"objy.objy":    Field{Name: "objy.objy", Type: "text"},
// 				"objy.stringy": Field{Name: "objy.stringy", Type: "text"},
// 				"stringy":      Field{Name: "stringy", Type: "text"},
// 				CSRFTokenName:  Field{Name: CSRFTokenName, Type: "hidden", Value: "foo_token"},
// 			},
// 		},
// 		{
// 			f:      "../stub/new-form.json",
// 			prefix: "traits",
// 			expected: Fields{
// 				"traits.numby":        Field{Name: "traits.numby", Type: "number"},
// 				"traits.objy.numby":   Field{Name: "traits.objy.numby", Type: "number"},
// 				"traits.objy.objy":    Field{Name: "traits.objy.objy", Type: "text"},
// 				"traits.objy.stringy": Field{Name: "traits.objy.stringy", Type: "text"},
// 				"traits.stringy":      Field{Name: "traits.stringy", Type: "text"},
// 				CSRFTokenName:         Field{Name: CSRFTokenName, Type: "hidden", Value: "foo_token"},
// 			},
// 		},
// 	} {
// 		t.Run(fmt.Sprintf("case=%d", k), func(t *testing.T) {
// 			fields, err := NewFormFieldsFromJSONSchema(tc.f, tc.prefix, "foo_token")
// 			require.NoError(t, err)
// 			assert.EqualValues(t, tc.expected, fields)
// 		})
// 	}
// }
