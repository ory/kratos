// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package node

import (
	"testing"

	"github.com/tidwall/gjson"

	"github.com/ory/x/jsonx"

	"github.com/stretchr/testify/assert"
)

func TestIDs(t *testing.T) {
	assert.EqualValues(t, "foo", (&AnchorAttributes{Identifier: "foo"}).ID())
	assert.EqualValues(t, "foo", (&ImageAttributes{Identifier: "foo"}).ID())
	assert.EqualValues(t, "foo", (&TextAttributes{Identifier: "foo"}).ID())
	assert.EqualValues(t, "foo", (&InputAttributes{Name: "foo"}).ID())
	assert.EqualValues(t, "foo", (&ScriptAttributes{Identifier: "foo"}).ID())
}

func TestMatchesAnchorAttributes(t *testing.T) {
	assert.True(t, (&AnchorAttributes{Identifier: "foo"}).Matches(&AnchorAttributes{Identifier: "foo"}))
	assert.True(t, (&AnchorAttributes{HREF: "bar"}).Matches(&AnchorAttributes{HREF: "bar"}))
	assert.False(t, (&AnchorAttributes{HREF: "foo"}).Matches(&AnchorAttributes{HREF: "bar"}))
	assert.False(t, (&AnchorAttributes{Identifier: "foo"}).Matches(&AnchorAttributes{HREF: "bar"}))

	assert.True(t, (&AnchorAttributes{Identifier: "foo", HREF: "bar"}).Matches(&AnchorAttributes{Identifier: "foo", HREF: "bar"}))
	assert.False(t, (&AnchorAttributes{Identifier: "foo", HREF: "bar"}).Matches(&AnchorAttributes{Identifier: "foo", HREF: "baz"}))
	assert.False(t, (&AnchorAttributes{Identifier: "foo", HREF: "bar"}).Matches(&AnchorAttributes{Identifier: "bar", HREF: "bar"}))

	assert.False(t, (&AnchorAttributes{Identifier: "foo"}).Matches(&TextAttributes{Identifier: "foo"}))
}

func TestMatchesImageAttributes(t *testing.T) {
	assert.True(t, (&ImageAttributes{Identifier: "foo"}).Matches(&ImageAttributes{Identifier: "foo"}))
	assert.True(t, (&ImageAttributes{Source: "bar"}).Matches(&ImageAttributes{Source: "bar"}))
	assert.False(t, (&ImageAttributes{Source: "foo"}).Matches(&ImageAttributes{Source: "bar"}))
	assert.False(t, (&ImageAttributes{Identifier: "foo"}).Matches(&ImageAttributes{Source: "bar"}))

	assert.True(t, (&ImageAttributes{Identifier: "foo", Source: "bar"}).Matches(&ImageAttributes{Identifier: "foo", Source: "bar"}))
	assert.False(t, (&ImageAttributes{Identifier: "foo", Source: "bar"}).Matches(&ImageAttributes{Identifier: "foo", Source: "baz"}))
	assert.False(t, (&ImageAttributes{Identifier: "foo", Source: "bar"}).Matches(&ImageAttributes{Identifier: "bar", Source: "bar"}))

	assert.False(t, (&ImageAttributes{Identifier: "foo"}).Matches(&TextAttributes{Identifier: "foo"}))
}

func TestMatchesInputAttributes(t *testing.T) {
	// Test when other is not of type *InputAttributes
	var attr Attributes = &ImageAttributes{}
	inputAttr := &InputAttributes{Name: "foo"}
	assert.False(t, inputAttr.Matches(attr))

	// Test when ID is different
	attr = &InputAttributes{Name: "foo", Type: InputAttributeTypeText}
	inputAttr = &InputAttributes{Name: "bar", Type: InputAttributeTypeText}
	assert.False(t, inputAttr.Matches(attr))

	// Test when Type is different
	attr = &InputAttributes{Name: "foo", Type: InputAttributeTypeText}
	inputAttr = &InputAttributes{Name: "foo", Type: InputAttributeTypeNumber}
	assert.False(t, inputAttr.Matches(attr))

	// Test when FieldValue is different
	attr = &InputAttributes{Name: "foo", Type: InputAttributeTypeText, FieldValue: "bar"}
	inputAttr = &InputAttributes{Name: "foo", Type: InputAttributeTypeText, FieldValue: "baz"}
	assert.False(t, inputAttr.Matches(attr))

	// Test when Name is different
	attr = &InputAttributes{Name: "foo", Type: InputAttributeTypeText}
	inputAttr = &InputAttributes{Name: "bar", Type: InputAttributeTypeText}
	assert.False(t, inputAttr.Matches(attr))

	// Test when all fields are the same
	attr = &InputAttributes{Name: "foo", Type: InputAttributeTypeText, FieldValue: "bar"}
	inputAttr = &InputAttributes{Name: "foo", Type: InputAttributeTypeText, FieldValue: "bar"}
	assert.True(t, inputAttr.Matches(attr))
}

func TestMatchesTextAttributes(t *testing.T) {
	assert.True(t, (&TextAttributes{Identifier: "foo"}).Matches(&TextAttributes{Identifier: "foo"}))
	assert.True(t, (&TextAttributes{Identifier: "foo"}).Matches(&TextAttributes{Identifier: "foo"}))
	assert.False(t, (&TextAttributes{Identifier: "foo"}).Matches(&ImageAttributes{Identifier: "foo"}))
}

func TestNodeEncode(t *testing.T) {
	script := jsonx.TestMarshalJSONString(t, &Node{Attributes: &ScriptAttributes{}})
	assert.EqualValues(t, Script, gjson.Get(script, "attributes.node_type").String())
	assert.EqualValues(t, Script, gjson.Get(script, "type").String())

	text := jsonx.TestMarshalJSONString(t, &Node{Attributes: &TextAttributes{}})
	assert.EqualValues(t, Text, gjson.Get(text, "attributes.node_type").String())
	assert.EqualValues(t, Text, gjson.Get(text, "type").String())

	image := jsonx.TestMarshalJSONString(t, &Node{Attributes: &ImageAttributes{}})
	assert.EqualValues(t, Image, gjson.Get(image, "attributes.node_type").String())
	assert.EqualValues(t, Image, gjson.Get(image, "type").String())

	input := jsonx.TestMarshalJSONString(t, &Node{Attributes: &InputAttributes{}})
	assert.EqualValues(t, Input, gjson.Get(input, "attributes.node_type").String())
	assert.EqualValues(t, Input, gjson.Get(input, "type").String())

	anchor := jsonx.TestMarshalJSONString(t, &Node{Attributes: &AnchorAttributes{}})
	assert.EqualValues(t, Anchor, gjson.Get(anchor, "attributes.node_type").String())
	assert.EqualValues(t, Anchor, gjson.Get(anchor, "type").String())
}

func TestNodeDecode(t *testing.T) {
	for _, kind := range []UiNodeType{
		Text,
		Input,
		Image,
		Anchor,
		Script,
	} {
		var n Node
		jsonx.TestUnmarshalJSON(t, []byte(`{"type":"`+kind+`"}`), &n)
		assert.EqualValues(t, kind, n.Attributes.GetNodeType())
	}
}
