package form

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"testing"

	"github.com/ory/kratos/ui/node"

	"github.com/pkg/errors"

	"github.com/ory/jsonschema/v3"

	"github.com/ory/herodot"
	"github.com/ory/x/decoderx"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/schema"
	"github.com/ory/kratos/text"
)

func newJSONRequest(t *testing.T, j string) *http.Request {
	req, err := http.NewRequest("POST", "/", bytes.NewBufferString(j))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	return req
}

func newFormRequest(t *testing.T, values url.Values) *http.Request {
	req, err := http.NewRequest("POST", "/", bytes.NewBufferString(values.Encode()))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return req
}

func TestContainer(t *testing.T) {
	t.Run("method=NewHTMLFormFromJSON", func(t *testing.T) {
		for k, tc := range []struct {
			r      string
			prefix string
			expect *HTMLForm
		}{
			{
				r: `{"numby":1.5,"stringy":"foobar","objy":{"objy":{},"numby":1.5,"stringy":"foobar"}}`,
				expect: &HTMLForm{
					Nodes: node.Nodes{
						node.NewInputFieldFromJSON("numby", 1.5, node.DefaultGroup),
						node.NewInputFieldFromJSON("objy.numby", 1.5, node.DefaultGroup),
						node.NewInputFieldFromJSON("objy.stringy", "foobar", node.DefaultGroup),
						node.NewInputFieldFromJSON("stringy", "foobar", node.DefaultGroup),
					},
				},
			},
			{
				r:      `{"numby":1.5,"stringy":"foobar","objy":{"objy":{},"numby":1.5,"stringy":"foobar"}}`,
				prefix: "traits",
				expect: &HTMLForm{
					Nodes: node.Nodes{
						node.NewInputFieldFromJSON("traits.numby", 1.5, node.DefaultGroup),
						node.NewInputFieldFromJSON("traits.objy.numby", 1.5, node.DefaultGroup),
						node.NewInputFieldFromJSON("traits.objy.stringy", "foobar", node.DefaultGroup),
						node.NewInputFieldFromJSON("traits.stringy", "foobar", node.DefaultGroup),
					},
				},
			},
		} {
			t.Run(fmt.Sprintf("case=%d", k), func(t *testing.T) {
				actual := NewHTMLFormFromJSON("action", node.DefaultGroup, json.RawMessage(tc.r), tc.prefix)

				// sort actual.fields lexicographically to have a deterministic order
				sort.SliceStable(actual.Nodes, func(i, j int) bool {
					return actual.Nodes[i].ID() < actual.Nodes[j].ID()
				})

				assert.Equal(t, "action", actual.Action)
				assert.EqualValues(t, tc.expect.Nodes, actual.Nodes)
			})
		}
	})

	t.Run("method=NewHTMLFormFromRequestBody", func(t *testing.T) {
		for k, tc := range []struct {
			ref    string
			r      *http.Request
			expect *HTMLForm
		}{
			{
				ref: "./stub/simple.schema.json",
				r:   newJSONRequest(t, `{"numby":1.5,"stringy":"foobar","objy":{"objy":{},"numby":1.5,"stringy":"foobar"}}`),
				expect: &HTMLForm{
					Nodes: node.Nodes{
						node.NewInputFieldFromJSON("numby", 1.5, node.DefaultGroup),
						node.NewInputFieldFromJSON("objy.numby", 1.5, node.DefaultGroup),
						node.NewInputFieldFromJSON("objy.stringy", "foobar", node.DefaultGroup),
						node.NewInputFieldFromJSON("stringy", "foobar", node.DefaultGroup),
					},
				},
			},
			{
				ref: "./stub/simple.schema.json",
				r: newFormRequest(t, url.Values{
					"numby":        {"1.5"},
					"stringy":      {"foobar"},
					"objy.numby":   {"1.5"},
					"objy.stringy": {"foobar"},
				}),
				expect: &HTMLForm{
					Nodes: node.Nodes{
						node.NewInputFieldFromJSON("numby", 1.5, node.DefaultGroup),
						node.NewInputFieldFromJSON("objy.numby", 1.5, node.DefaultGroup),
						node.NewInputFieldFromJSON("objy.stringy", "foobar", node.DefaultGroup),
						node.NewInputFieldFromJSON("stringy", "foobar", node.DefaultGroup),
					},
				},
			},
			{
				ref: "./stub/complex.schema.json",
				r: newFormRequest(t, url.Values{
					"meal.chef": {"aeneas"},
				}),
				expect: &HTMLForm{
					Nodes: node.Nodes{
						node.NewInputFieldFromJSON("meal.chef", "aeneas", node.DefaultGroup),
						&node.Node{Group: node.DefaultGroup, Type: node.Input, Attributes: &node.InputAttributes{Name: "meal.name", Type: node.InputAttributeTypeText}, Messages: text.Messages{*text.NewValidationErrorRequired("name")}},
					},
				},
			},
		} {
			t.Run(fmt.Sprintf("case=%d", k), func(t *testing.T) {
				actual, err := NewHTMLFormFromRequestBody(tc.r, node.DefaultGroup, "action", decoderx.HTTPJSONSchemaCompiler(tc.ref, nil))
				require.NoError(t, err)
				// sort actual.fields lexicographically to have a deterministic order
				sort.SliceStable(actual.Nodes, func(i, j int) bool {
					return actual.Nodes[i].ID() < actual.Nodes[j].ID()
				})
				assert.Equal(t, "action", actual.Action)
				assert.EqualValues(t, tc.expect.Nodes, actual.Nodes)
			})
		}
	})

	t.Run("method=NewHTMLFormFromJSONSchema", func(t *testing.T) {
		for k, tc := range []struct {
			ref    string
			prefix string
			expect *HTMLForm
		}{
			{
				ref:    "./stub/simple.schema.json",
				prefix: "",
				expect: &HTMLForm{
					Nodes: node.Nodes{
						node.NewInputField("numby", nil, node.DefaultGroup, node.InputAttributeTypeNumber),
						node.NewInputField("objy.numby", nil, node.DefaultGroup, node.InputAttributeTypeNumber),
						node.NewInputField("objy.objy", nil, node.DefaultGroup, node.InputAttributeTypeText),
						node.NewInputField("objy.stringy", nil, node.DefaultGroup, node.InputAttributeTypeText),
						node.NewInputField("stringy", nil, node.DefaultGroup, node.InputAttributeTypeText),
					},
				},
			},
			{
				ref:    "./stub/simple.schema.json",
				prefix: "traits",
				expect: &HTMLForm{
					Nodes: node.Nodes{
						node.NewInputField("traits.numby", nil, node.DefaultGroup, node.InputAttributeTypeNumber),
						node.NewInputField("traits.objy.numby", nil, node.DefaultGroup, node.InputAttributeTypeNumber),
						node.NewInputField("traits.objy.objy", nil, node.DefaultGroup, node.InputAttributeTypeText),
						node.NewInputField("traits.objy.stringy", nil, node.DefaultGroup, node.InputAttributeTypeText),
						node.NewInputField("traits.stringy", nil, node.DefaultGroup, node.InputAttributeTypeText),
					},
				},
			},
			{
				ref: "./stub/complex.schema.json",
				expect: &HTMLForm{
					Nodes: node.Nodes{
						node.NewInputField("fruits", nil, node.DefaultGroup, node.InputAttributeTypeText),
						node.NewInputField("meal.chef", nil, node.DefaultGroup, node.InputAttributeTypeText),
						node.NewInputField("meal.name", nil, node.DefaultGroup, node.InputAttributeTypeText),
						node.NewInputField("vegetables", nil, node.DefaultGroup, node.InputAttributeTypeText),
					},
				},
			},
		} {
			t.Run(fmt.Sprintf("case=%d", k), func(t *testing.T) {
				actual, err := NewHTMLFormFromJSONSchema("action",
					node.DefaultGroup, tc.ref, tc.prefix, nil)
				require.NoError(t, err)
				assert.Equal(t, "action", actual.Action)
				assert.EqualValues(t, tc.expect.Messages, actual.Messages)
				assert.EqualValues(t, tc.expect.Nodes, actual.Nodes)
			})
		}
	})

	t.Run("method=ParseError", func(t *testing.T) {
		for k, tc := range []struct {
			err       error
			expectErr bool
			expect    HTMLForm
		}{
			{err: errors.New("foo"), expectErr: true},
			{err: &herodot.ErrNotFound, expectErr: true},
			{err: herodot.ErrBadRequest.WithReason("tests"), expect: HTMLForm{Nodes: node.Nodes{}, Messages: text.Messages{*text.NewValidationErrorGeneric("tests")}}},
			{err: schema.NewInvalidCredentialsError(), expect: HTMLForm{Nodes: node.Nodes{}, Messages: text.Messages{*text.NewErrorValidationInvalidCredentials()}}},
			{err: &jsonschema.ValidationError{Message: "test", InstancePtr: "#/foo/bar/baz"}, expect: HTMLForm{Nodes: node.Nodes{
				&node.Node{Group: node.DefaultGroup, Type: node.Input, Attributes: &node.InputAttributes{Name: "foo.bar.baz", Type: node.InputAttributeTypeText}, Messages: text.Messages{*text.NewValidationErrorGeneric("test")}},
			}}},
			{err: &jsonschema.ValidationError{Message: "test", InstancePtr: ""}, expect: HTMLForm{Nodes: node.Nodes{}, Messages: text.Messages{*text.NewValidationErrorGeneric("test")}}},
		} {
			t.Run(fmt.Sprintf("case=%d", k), func(t *testing.T) {
				for _, in := range []error{tc.err, errors.WithStack(tc.err)} {
					c := NewHTMLForm("")
					err := c.ParseError(node.DefaultGroup, in)
					if tc.expectErr {
						require.Error(t, err)
						return
					}
					require.NoError(t, err)
					assert.EqualValues(t, tc.expect.Messages, c.Messages)
					assert.EqualValues(t, tc.expect.Nodes, c.Nodes)
				}
			})
		}
	})

	t.Run("method=SetValue", func(t *testing.T) {
		c := HTMLForm{
			Nodes: node.Nodes{
				node.NewInputField("1", "foo", node.DefaultGroup, node.InputAttributeTypeText),
				node.NewInputField("2", "", node.DefaultGroup, node.InputAttributeTypeText),
			},
		}

		assert.Len(t, c.Nodes, 2)

		c.SetValue("1", node.NewInputFieldFromJSON("1", "baz1", node.DefaultGroup))
		c.SetValue("2", node.NewInputFieldFromJSON("2", "baz2", node.DefaultGroup))
		c.SetValue("3", node.NewInputFieldFromJSON("3", "baz3", node.DefaultGroup))

		assert.Len(t, c.Nodes, 3)
		for _, k := range []string{"1", "2", "3"} {
			assert.EqualValues(t, fmt.Sprintf("baz%s", k), c.Nodes.Find(node.DefaultGroup, k).Attributes.GetValue(), "%+v", c)
		}
	})

	t.Run("method=SetCSRF", func(t *testing.T) {
		f := &HTMLForm{Nodes: node.Nodes{}}
		f.SetCSRF("csrf-token")
		assert.Contains(t, f.Nodes, node.NewCSRFNode("csrf-token"))
	})

	t.Run("method=AddMessage", func(t *testing.T) {
		c := HTMLForm{
			Nodes: node.Nodes{
				&node.Node{Messages: text.Messages{*text.NewValidationErrorGeneric("foo")}, Group: node.DefaultGroup, Type: node.Input, Attributes: &node.InputAttributes{Name: "1", Type: node.InputAttributeTypeText, FieldValue: ""}},
				&node.Node{Messages: text.Messages{}, Group: node.DefaultGroup, Type: node.Input, Attributes: &node.InputAttributes{Name: "2", Type: node.InputAttributeTypeText, FieldValue: ""}},
			},
		}
		assert.Len(t, c.Nodes, 2)
		c.AddMessage(node.DefaultGroup, &text.Message{Text: "baz1"}, "1")
		c.AddMessage(node.DefaultGroup, &text.Message{Text: "baz2"}, "2")
		c.AddMessage(node.DefaultGroup, &text.Message{Text: "baz3"}, "3")
		c.AddMessage(node.DefaultGroup, &text.Message{Text: "baz"}, "4", "5", "6")
		c.AddMessage(node.DefaultGroup, &text.Message{Text: "rootbar"})

		assert.Len(t, c.Nodes, 6)

		for _, k := range []string{"1", "2", "3"} {
			n := c.Nodes.Find(node.DefaultGroup, k)
			assert.EqualValues(t, fmt.Sprintf("baz%s", k), n.Messages[len(n.Messages)-1].Text, "%+v", c)
		}

		for _, k := range []string{"4", "5", "6"} {
			assert.EqualValues(t, "baz", c.Nodes.Find(node.DefaultGroup, k).Messages[0].Text, "%+v", c)
		}

		assert.Len(t, c.Messages, 1)
		assert.Equal(t, "rootbar", c.Messages[0].Text)
	})

	t.Run("method=Reset", func(t *testing.T) {
		c := HTMLForm{
			Nodes: node.Nodes{
				&node.Node{Messages: text.Messages{{Text: "foo"}}, Group: node.DefaultGroup, Type: node.Input, Attributes: &node.InputAttributes{Name: "1", Type: node.InputAttributeTypeText, FieldValue: "foo"}},
				&node.Node{Messages: text.Messages{{Text: "bar"}}, Group: node.DefaultGroup, Type: node.Input, Attributes: &node.InputAttributes{Name: "2", Type: node.InputAttributeTypeText, FieldValue: "bar"}},
			},
			Messages: text.Messages{{Text: ""}},
		}
		c.Reset()

		assert.Empty(t, c.Messages)
		assert.Empty(t, c.Nodes.Find(node.DefaultGroup, "1").Messages)
		assert.Empty(t, c.Nodes.Find(node.DefaultGroup, "1").Attributes.(*node.InputAttributes).FieldValue)
		assert.Empty(t, c.Nodes.Find(node.DefaultGroup, "2").Messages)
		assert.Empty(t, c.Nodes.Find(node.DefaultGroup, "2").Attributes.(*node.InputAttributes).FieldValue)
	})

	t.Run("method=SortFields", func(t *testing.T) {
		// use a schema compiler that disables identifiers
		schemaCompiler := jsonschema.NewCompiler()
		schemaPath := "stub/identity.schema.json"

		f, err := NewHTMLFormFromJSONSchema("/foo", node.DefaultGroup, schemaPath, "", schemaCompiler)
		require.NoError(t, err)

		f.UpdateNodesFromJSON(json.RawMessage(`{}`), "traits", node.DefaultGroup)
		f.SetCSRF("csrf_token")

		require.NoError(t, f.SortFields(schemaPath))

		var names []string
		for _, f := range f.Nodes {
			names = append(names, f.Attributes.(*node.InputAttributes).Name)
		}

		assert.EqualValues(t, []string{"csrf_token", "traits.email", "traits.stringy", "traits.numby", "traits.booly", "traits.should_big_number", "traits.should_long_string"}, names, "%+v", f.Nodes)
	})
}
