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
