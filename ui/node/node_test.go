// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package node_test

import (
	"bytes"
	"context"
	"embed"
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/ory/x/snapshotx"

	"github.com/ory/kratos/text"

	"github.com/ory/x/assertx"

	"github.com/ory/kratos/corpx"

	"github.com/ory/kratos/ui/container"

	"github.com/go-faker/faker/v4"
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
var ctx = context.Background()

func TestNodesSort(t *testing.T) {
	// use a schema compiler that disables identifiers
	schemaCompiler := jsonschema.NewCompiler()
	schemaPath := "fixtures/identity.schema.json"

	f, err := container.NewFromJSONSchema(ctx, "/foo", node.DefaultGroup, schemaPath, "", schemaCompiler)
	require.NoError(t, err)

	f.UpdateNodeValuesFromJSON(json.RawMessage(`{}`), "traits", node.DefaultGroup)
	f.SetCSRF("csrf_token")

	inputs, err := sortFixtures.ReadDir("fixtures/sort/input")
	require.NoError(t, err)

	options := map[string][]node.SortOption{
		"1.json": {
			node.SortUseOrder([]string{"password_identifier"}),
			node.SortUpdateOrder(node.PasswordLoginOrder),
			node.SortByGroups([]node.UiNodeGroup{
				node.DefaultGroup,
				node.ProfileGroup,
				node.OpenIDConnectGroup,
				node.PasswordGroup,
				node.LinkGroup,
				node.LinkGroup,
			}),
		},
		"2.json": {
			node.SortBySchema(filepath.Join("fixtures/sort/schema", "2.json")),
			node.SortUpdateOrder(node.PasswordLoginOrder),
			node.SortByGroups([]node.UiNodeGroup{
				node.DefaultGroup,
				node.OpenIDConnectGroup,
				node.PasswordGroup,
			}),
		},
		"3.json": {
			node.SortBySchema(filepath.Join("fixtures/sort/schema", "3.json")),
			node.SortByGroups([]node.UiNodeGroup{
				node.DefaultGroup,
				node.OpenIDConnectGroup,
				node.PasswordGroup,
			}),
		},
		"4.json": {
			node.SortBySchema(filepath.Join("fixtures/sort/schema", "4.json")),
			node.SortByGroups([]node.UiNodeGroup{
				node.DefaultGroup,
				node.ProfileGroup,
				node.PasswordGroup,
				node.OpenIDConnectGroup,
				node.LookupGroup,
				node.WebAuthnGroup,
				node.TOTPGroup,
			}),
			node.SortUseOrderAppend([]string{
				// Lookup
				node.LookupReveal,
				node.LookupRegenerate,
				node.LookupCodes,
				node.LookupConfirm,

				// Lookup
				node.WebAuthnRemove,
				node.WebAuthnRegisterDisplayName,
				node.WebAuthnRegister,

				// TOTP
				node.TOTPSecretKey,
				node.TOTPQR,
				node.TOTPUnlink,
				node.TOTPCode,
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

			require.NoError(t, nodes.SortBySchema(ctx, options[in.Name()]...))

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

func TestMatchesNode(t *testing.T) {
	// Test when ID is different
	node1 := &node.Node{Type: node.Input, Group: node.PasswordGroup, Attributes: &node.InputAttributes{Name: "foo"}}
	node2 := &node.Node{Type: node.Input, Group: node.PasswordGroup, Attributes: &node.InputAttributes{Name: "bar"}}
	assert.False(t, node1.Matches(node2))

	// Test when Type is different
	node1 = &node.Node{Type: node.Input, Group: node.PasswordGroup, Attributes: &node.InputAttributes{Name: "foo"}}
	node2 = &node.Node{Type: node.Text, Group: node.PasswordGroup, Attributes: &node.InputAttributes{Name: "foo"}}
	assert.False(t, node1.Matches(node2))

	// Test when Group is different
	node1 = &node.Node{Type: node.Input, Group: node.PasswordGroup, Attributes: &node.InputAttributes{Name: "foo"}}
	node2 = &node.Node{Type: node.Input, Group: node.OpenIDConnectGroup, Attributes: &node.InputAttributes{Name: "foo"}}
	assert.False(t, node1.Matches(node2))

	// Test when all fields are the same
	node1 = &node.Node{Type: node.Input, Group: node.PasswordGroup, Attributes: &node.InputAttributes{Name: "foo"}}
	node2 = &node.Node{Type: node.Input, Group: node.PasswordGroup, Attributes: &node.InputAttributes{Name: "foo"}}
	assert.True(t, node1.Matches(node2))
}

func TestRemoveMatchingNodes(t *testing.T) {
	nodes := node.Nodes{
		&node.Node{Type: node.Input, Group: node.PasswordGroup, Attributes: &node.InputAttributes{Name: "foo"}},
		&node.Node{Type: node.Input, Group: node.PasswordGroup, Attributes: &node.InputAttributes{Name: "bar"}},
		&node.Node{Type: node.Input, Group: node.PasswordGroup, Attributes: &node.InputAttributes{Name: "baz"}},
	}

	// Test when node to remove is present
	nodeToRemove := &node.Node{Type: node.Input, Group: node.PasswordGroup, Attributes: &node.InputAttributes{Name: "bar"}}
	nodes.RemoveMatching(nodeToRemove)
	assert.Len(t, nodes, 2)
	for _, n := range nodes {
		assert.NotEqual(t, nodeToRemove.ID(), n.ID())
	}

	// Test when node to remove is not present
	nodeToRemove = &node.Node{Type: node.Input, Group: node.PasswordGroup, Attributes: &node.InputAttributes{Name: "qux"}}
	nodes.RemoveMatching(nodeToRemove)
	assert.Len(t, nodes, 2) // length should remain the same

	// Test when node to remove is present
	nodeToRemove = &node.Node{Type: node.Input, Group: node.PasswordGroup, Attributes: &node.InputAttributes{Name: "baz"}}
	ui := &container.Container{
		Nodes: nodes,
	}

	ui.GetNodes().RemoveMatching(nodeToRemove)
	assert.Len(t, *ui.GetNodes(), 1)
	for _, n := range *ui.GetNodes() {
		assert.NotEqual(t, "bar", n.ID())
		assert.NotEqual(t, "baz", n.ID())
	}

	ui.Nodes.Append(node.NewInputField("method", "foo", "bar", node.InputAttributeTypeSubmit).WithMetaLabel(text.NewInfoNodeLabelContinue()))
	assert.NotNil(t, ui.Nodes.Find("method"))
	ui.GetNodes().RemoveMatching(node.NewInputField("method", "foo", "bar", node.InputAttributeTypeSubmit))
	assert.Nil(t, ui.Nodes.Find("method"))
}

func TestNodeMarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		node    *node.Node
		wantErr bool
		errMsg  string
	}{
		{
			name: "text node",
			node: &node.Node{
				Type:  node.Text,
				Group: node.DefaultGroup,
				Attributes: &node.TextAttributes{
					NodeType: node.Text,
				},
				Messages: text.Messages{},
				Meta:     &node.Meta{},
			},
		},
		{
			name: "input node",
			node: &node.Node{
				Type:  node.Input,
				Group: node.DefaultGroup,
				Attributes: &node.InputAttributes{
					NodeType:   node.Input,
					Name:       "password",
					Type:       "password",
					FieldValue: "secret",
				},
				Messages: text.Messages{},
				Meta:     &node.Meta{},
			},
		},
		{
			name: "anchor node",
			node: &node.Node{
				Type:  node.Anchor,
				Group: node.DefaultGroup,
				Attributes: &node.AnchorAttributes{
					NodeType: node.Anchor,
					HREF:     "https://example.com",
				},
				Messages: text.Messages{},
				Meta:     &node.Meta{},
			},
		},
		{
			name: "image node",
			node: &node.Node{
				Type:  node.Image,
				Group: node.DefaultGroup,
				Attributes: &node.ImageAttributes{
					NodeType: node.Image,
					Source:   "image.jpg",
				},
				Messages: text.Messages{},
				Meta:     &node.Meta{},
			},
		},
		{
			name: "script node",
			node: &node.Node{
				Type:  node.Script,
				Group: node.DefaultGroup,
				Attributes: &node.ScriptAttributes{
					NodeType: node.Script,
					Source:   "script.js",
				},
				Messages: text.Messages{},
				Meta:     &node.Meta{},
			},
		},
		{
			name: "division node",
			node: &node.Node{
				Type:  node.Division,
				Group: node.DefaultGroup,
				Attributes: &node.DivisionAttributes{
					NodeType: node.Division,
				},
				Messages: text.Messages{},
				Meta:     &node.Meta{},
			},
		},
		{
			name: "type mismatch",
			node: &node.Node{
				Type:       node.Image,
				Group:      node.DefaultGroup,
				Attributes: &node.InputAttributes{NodeType: node.Input},
			},
			wantErr: true,
			errMsg:  "node type and node attributes mismatch",
		},
		{
			name: "empty type inferred from attributes",
			node: &node.Node{
				Group: node.DefaultGroup,
				Attributes: &node.InputAttributes{
					NodeType: node.Input,
					Name:     "email",
				},
				Messages: text.Messages{},
				Meta:     &node.Meta{},
			},
		},
		{
			name: "nil attributes",
			node: &node.Node{
				Type:       node.Image,
				Group:      node.DefaultGroup,
				Attributes: nil,
				Messages:   text.Messages{},
			},
			wantErr: true,
			errMsg:  "node type and node attributes mismatch",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.node)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				return
			}

			require.NoError(t, err)

			// Use snapshotx for testing serialization
			snapshotx.SnapshotT(t, json.RawMessage(data))

			// Verify roundtrip
			var unmarshalled node.Node
			err = json.Unmarshal(data, &unmarshalled)
			require.NoError(t, err)

			// Re-marshal for comparison
			remarshalled, err := json.Marshal(&unmarshalled)
			require.NoError(t, err)
			assert.JSONEq(t, string(data), string(remarshalled))
		})
	}
}
