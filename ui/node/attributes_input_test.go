// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package node

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	"github.com/ory/jsonschema/v3"
	"github.com/ory/x/jsonschemax"
)

func TestFieldFromPath(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	t.Run("all properties are properly transferred", func(t *testing.T) {
		t.Parallel()

		schema, err := os.ReadFile("./fixtures/all_formats.schema.json")
		require.NoError(t, err)

		c := jsonschema.NewCompiler()
		require.NoError(t, c.AddResource("test.json", bytes.NewBuffer(schema)))

		paths, err := jsonschemax.ListPaths(ctx, "test.json", c)
		require.NoError(t, err)

		for _, path := range paths {
			node := NewInputFieldFromSchema(path.Name, DefaultGroup, path)
			assert.EqualValues(t, "input", node.Type)
			require.IsType(t, new(InputAttributes), node.Attributes)
			attr := node.Attributes.(*InputAttributes)

			assert.EqualValues(t, gjson.GetBytes(schema, fmt.Sprintf("properties.%s.test_expected_type", path.Name)).String(), attr.Type)
			assert.True(t, !gjson.GetBytes(schema, fmt.Sprintf("properties.%s.test_expected_pattern", path.Name)).Exists() ||
				(gjson.GetBytes(schema, fmt.Sprintf("properties.%s.test_expected_pattern", path.Name)).Bool() && attr.Pattern != ""))

			expectedAutocomplete := gjson.GetBytes(schema, fmt.Sprintf("properties.%s.test_expected_autocomplete", path.Name))

			if expectedAutocomplete.Exists() {
				assert.EqualValues(t, expectedAutocomplete.String(), attr.Autocomplete)
			}

			expectedOptions := gjson.GetBytes(schema, fmt.Sprintf("properties.%s.test_expected_options", path.Name))
			if expectedOptions.Exists() {
				// Compare through JSON so string, number, and boolean
				// option values all survive the round-trip without relying
				// on a lossy `%v` coercion.
				var got []any
				for _, o := range attr.Options {
					got = append(got, o.Value)
				}
				gotJSON, err := json.Marshal(got)
				require.NoError(t, err)
				assert.JSONEq(t, expectedOptions.Raw, string(gotJSON), "field %s", path.Name)
			} else {
				assert.Empty(t, attr.Options, "field %s should have no options", path.Name)
			}
		}
	})

	t.Run("caps enum options at maxEnumOptions", func(t *testing.T) {
		t.Parallel()

		enum := make([]any, maxEnumOptions+50)
		for i := range enum {
			enum[i] = fmt.Sprintf("v%d", i)
		}
		node := NewInputFieldFromSchema("field", DefaultGroup, jsonschemax.Path{
			Name: "field",
			Type: "",
			Enum: enum,
		})
		attr := node.Attributes.(*InputAttributes)
		assert.Len(t, attr.Options, maxEnumOptions)
		assert.Equal(t, "v0", attr.Options[0].Value)
		assert.Equal(t, fmt.Sprintf("v%d", maxEnumOptions-1), attr.Options[maxEnumOptions-1].Value)
	})

	t.Run("empty enum produces no options", func(t *testing.T) {
		t.Parallel()

		node := NewInputFieldFromSchema("field", DefaultGroup, jsonschemax.Path{
			Name: "field",
			Type: "",
			Enum: []any{},
		})
		attr := node.Attributes.(*InputAttributes)
		assert.Empty(t, attr.Options)
	})
}
