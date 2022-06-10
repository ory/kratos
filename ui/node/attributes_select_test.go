package node

import (
	"bytes"
	"context"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/jsonschema/v3"
	"github.com/ory/x/jsonschemax"
)

func TestSelectFromSchema(t *testing.T) {

	var ctx = context.Background()
	t.Run("all properties are properly transferred", func(t *testing.T) {
		schema, err := ioutil.ReadFile("./fixtures/enum.schema.json")
		require.NoError(t, err)

		c := jsonschema.NewCompiler()
		require.NoError(t, c.AddResource("test.json", bytes.NewBuffer(schema)))

		paths, err := jsonschemax.ListPaths(ctx, "test.json", c)
		require.NoError(t, err)

		for _, path := range paths {
			node := NewSelectFieldFromSchema(path.Name, DefaultGroup, path)
			assert.EqualValues(t, "select", node.Type)
			require.IsType(t, new(SelectAttributes), node.Attributes)
			attr := node.Attributes.(*SelectAttributes)

			assert.True(t, attr.Options != nil && len(attr.Options) > 0)

			for i, v := range path.Enum {
				assert.EqualValues(t, attr.Options[i].Label, v)
				assert.EqualValues(t, attr.Options[i].Value, v)
			}
		}
	})
}
