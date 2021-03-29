package node_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/ory/kratos/x"

	"github.com/ory/kratos/corpx"

	"github.com/ory/kratos/ui/container"

	"github.com/bxcodec/faker/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	"github.com/ory/jsonschema/v3"
	"github.com/ory/kratos/ui/node"
)

func init() {
	corpx.RegisterFakes()
}

func TestNodesSort(t *testing.T) {
	// use a schema compiler that disables identifiers
	schemaCompiler := jsonschema.NewCompiler()
	schemaPath := "stub/identity.schema.json"

	f, err := container.NewFromJSONSchema("/foo", node.DefaultGroup, schemaPath, "", schemaCompiler)
	require.NoError(t, err)

	f.UpdateNodesFromJSON(json.RawMessage(`{}`), "traits", node.DefaultGroup)
	f.SetCSRF("csrf_token")

	for k, tc := range []struct {
		p string
		k []string
		e []string
	}{
		{
			k: []string{
				"traits.stringy",
				x.CSRFTokenName,
				"traits.numby",
			},
			e: []string{"traits.stringy", "csrf_token", "traits.numby", "traits.email", "traits.booly", "traits.should_big_number", "traits.should_long_string"},
		},
		{
			p: "traits",
			k: []string{
				x.CSRFTokenName,
				"traits.stringy",
				"traits.numby",
			},
			e: []string{"csrf_token", "traits.stringy", "traits.numby", "traits.email", "traits.booly", "traits.should_big_number", "traits.should_long_string"},
		},
	} {
		t.Run(fmt.Sprintf("case=%d", k), func(t *testing.T) {
			require.NoError(t, f.SortNodes(schemaPath, tc.p, tc.k))

			var names []string
			for _, f := range f.Nodes {
				names = append(names, f.Attributes.ID())
			}

			assert.EqualValues(t, tc.e, names, "%+v", f.Nodes)
		})
	}
}

func TestNodesUpsert(t *testing.T) {
	var nodes node.Nodes
	nodes.Upsert(node.NewCSRFNode("foo"))
	require.Len(t, nodes, 1)
	nodes.Upsert(node.NewCSRFNode("bar"))
	require.Len(t, nodes, 1)
	assert.EqualValues(t, "bar", nodes[0].Attributes.GetValue())
}

func TestNodesRemove(t *testing.T) {
	var nodes node.Nodes
	nodes.Append(node.NewInputField("other", "foo", node.OpenIDConnectGroup, node.InputAttributeTypeSubmit))
	nodes.Append(node.NewInputField("link", "foo", node.OpenIDConnectGroup, node.InputAttributeTypeSubmit))
	nodes.Append(node.NewInputField("link", "bar", node.OpenIDConnectGroup, node.InputAttributeTypeSubmit))
	nodes.Append(node.NewInputField("unlink", "baz", node.OpenIDConnectGroup, node.InputAttributeTypeSubmit))
	require.Len(t, nodes, 4)

	nodes.Remove("link", "unlink")
	require.Len(t, nodes, 1)
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
