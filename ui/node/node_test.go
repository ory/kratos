package node_test

import (
	"bytes"
	"embed"
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/ory/x/assertx"

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

//go:embed fixtures/sort/*
var sortFixtures embed.FS

func TestNodesSort(t *testing.T) {
	// use a schema compiler that disables identifiers
	schemaCompiler := jsonschema.NewCompiler()
	schemaPath := "fixtures/identity.schema.json"

	f, err := container.NewFromJSONSchema("/foo", node.DefaultGroup, schemaPath, "", schemaCompiler)
	require.NoError(t, err)

	f.UpdateNodeValuesFromJSON(json.RawMessage(`{}`), "traits", node.DefaultGroup)
	f.SetCSRF("csrf_token")

	inputs, err := sortFixtures.ReadDir("fixtures/sort/input")
	require.NoError(t, err)

	options := map[string][]node.SortOption{
		"1.json": {
			node.SortUseOrder([]string{"password_identifier"}),
			node.SortUpdateOrder(node.PasswordLoginOrder),
			node.SortByGroups([]node.Group{
				node.DefaultGroup,
				node.ProfileGroup,
				node.OpenIDConnectGroup,
				node.PasswordGroup,
				node.RecoveryLinkGroup,
				node.VerificationLinkGroup,
			}),
		},
		"2.json": {
			node.SortBySchema(filepath.Join("fixtures/sort/schema", "2.json")),
			node.SortUpdateOrder(node.PasswordLoginOrder),
			node.SortByGroups([]node.Group{
				node.DefaultGroup,
				node.OpenIDConnectGroup,
				node.PasswordGroup,
			}),
		},
		"3.json": {
			node.SortBySchema(filepath.Join("fixtures/sort/schema", "3.json")),
			node.SortByGroups([]node.Group{
				node.DefaultGroup,
				node.OpenIDConnectGroup,
				node.PasswordGroup,
			}),
		},
	}

	for _, in := range inputs {
		t.Run("file="+in.Name(), func(t *testing.T) {
			if in.IsDir() {
				return
			}

			fi, err := sortFixtures.Open(filepath.Join("fixtures/sort/input", in.Name()))
			require.NoError(t, err)
			defer fi.Close()

			var nodes node.Nodes
			require.NoError(t, json.NewDecoder(fi).Decode(&nodes))
			require.NotEmpty(t, nodes)

			require.NoError(t, nodes.SortBySchema(options[in.Name()]...))

			fe, err := sortFixtures.ReadFile(filepath.Join("fixtures/sort/expected", in.Name()))
			require.NoError(t, err)

			assertx.EqualAsJSON(t, json.RawMessage(fe), nodes)
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
