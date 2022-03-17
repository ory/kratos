package node

import (
	"bytes"
	"context"
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

	var ctx = context.Background()
	t.Run("all properties are properly transferred", func(t *testing.T) {
		schema, err := ioutil.ReadFile("./fixtures/all_formats.schema.json")
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
		}
	})
}
