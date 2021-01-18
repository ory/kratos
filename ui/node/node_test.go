package node_test

import (
	"bytes"
	"encoding/json"
	"github.com/bxcodec/faker/v3"
	"github.com/ory/jsonschema/v3"
	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/selfservice/form"
	"github.com/ory/kratos/ui/node"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
	"testing"
)

func init() {
	internal.RegisterFakes()
}

func TestNodesSort(t *testing.T) {
	// use a schema compiler that disables identifiers
	schemaCompiler := jsonschema.NewCompiler()
	schemaPath := "stub/identity.schema.json"

	f, err := form.NewHTMLFormFromJSONSchema("/foo", node.DefaultGroup, schemaPath, "", schemaCompiler)
	require.NoError(t, err)

	f.UpdateNodesFromJSON(json.RawMessage(`{}`), "traits", node.DefaultGroup)
	f.SetCSRF("csrf_token")

	require.NoError(t, f.SortFields(schemaPath))

	var names []string
	for _, f := range f.Nodes {
		names = append(names, f.Attributes.ID())
	}

	assert.EqualValues(t, []string{"csrf_token", "traits.email", "traits.stringy", "traits.numby", "traits.booly", "traits.should_big_number", "traits.should_long_string"}, names, "%+v", f.Nodes)
}

func TestNodesUpsert(t *testing.T) {
	var nodes node.Nodes
	nodes.Upsert(node.NewCSRFNode("foo"))
	require.Len(t, nodes, 1)
	nodes.Upsert(node.NewCSRFNode("bar"))
	require.Len(t, nodes, 1)
	assert.EqualValues(t, "bar", nodes[0].Attributes.GetValue())
}

func TestNodeJSON(t *testing.T) {
	t.Run("idempotent decode", func(t *testing.T) {
		nodes := make(node.Nodes, 5)
		for k := range nodes {
			nodes[k] = new(node.Node)
			expected := nodes[k]

			require.NoError(t, faker.FakeData(expected))
			var b bytes.Buffer
			require.NoError(t, json.NewEncoder(&b).Encode(expected))

			var actual node.Node
			require.NoError(t, json.NewDecoder(&b).Decode(&actual))

			assert.EqualValues(t, *expected, actual)
			assert.NotEmpty(t, actual.Type)
			assert.NotNil(t, actual.Attributes)
		}
	})

	t.Run("type mismatch", func(t *testing.T) {
		n := &node.Node{Type: node.Image, Attributes: new(node.InputAttributes)}
		var b bytes.Buffer
		require.EqualError(t, json.NewEncoder(&b).Encode(n), "json: error calling MarshalJSON for type *node.Node: node type and node attributes mismatch: *node.InputAttributes != img")
	})

	t.Run("type empty", func(t *testing.T) {
		n := &node.Node{Attributes: new(node.InputAttributes)}
		var b bytes.Buffer
		require.NoError(t, json.NewEncoder(&b).Encode(n))
		assert.EqualValues(t, node.Input, gjson.GetBytes(b.Bytes(), "type").String())
	})

	t.Run("type decode unknown", func(t *testing.T) {
		var n node.Node
		require.EqualError(t, json.NewDecoder(bytes.NewReader(json.RawMessage(`{"type": "foo"}`))).Decode(&n), "unexpected node type: foo")
	})
}
