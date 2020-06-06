package form

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"testing"

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
					Fields: Fields{
						Field{Name: "numby", Type: "number", Value: 1.5},
						Field{Name: "objy.numby", Type: "number", Value: 1.5},
						Field{Name: "objy.stringy", Type: "text", Value: "foobar"},
						Field{Name: "stringy", Type: "text", Value: "foobar"},
					},
				},
			},
			{
				r:      `{"numby":1.5,"stringy":"foobar","objy":{"objy":{},"numby":1.5,"stringy":"foobar"}}`,
				prefix: "traits",
				expect: &HTMLForm{
					Fields: Fields{
						Field{Name: "traits.numby", Type: "number", Value: 1.5},
						Field{Name: "traits.objy.numby", Type: "number", Value: 1.5},
						Field{Name: "traits.objy.stringy", Type: "text", Value: "foobar"},
						Field{Name: "traits.stringy", Type: "text", Value: "foobar"},
					},
				},
			},
		} {
			t.Run(fmt.Sprintf("case=%d", k), func(t *testing.T) {
				actual := NewHTMLFormFromJSON("action", json.RawMessage(tc.r), tc.prefix)
				// sort actual.fields lexicographically to have a deterministic order
				sort.SliceStable(actual.Fields, func(i, j int) bool {
					return actual.Fields[i].Name < actual.Fields[j].Name
				})
				assert.Equal(t, "action", actual.Action)
				assert.EqualValues(t, tc.expect.Fields, actual.Fields)
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
					Fields: Fields{
						Field{Name: "numby", Type: "number", Value: 1.5},
						Field{Name: "objy.numby", Type: "number", Value: 1.5},
						Field{Name: "objy.stringy", Type: "text", Value: "foobar"},
						Field{Name: "stringy", Type: "text", Value: "foobar"},
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
					Fields: Fields{
						Field{Name: "numby", Type: "number", Value: 1.5},
						Field{Name: "objy.numby", Type: "number", Value: 1.5},
						Field{Name: "objy.stringy", Type: "text", Value: "foobar"},
						Field{Name: "stringy", Type: "text", Value: "foobar"},
					},
				},
			},
			{
				ref: "./stub/complex.schema.json",
				r: newFormRequest(t, url.Values{
					"meal.chef": {"aeneas"},
				}),
				expect: &HTMLForm{
					Fields: Fields{
						Field{Name: "meal.chef", Type: "text", Value: "aeneas"},
						Field{Name: "meal.name", Messages: text.Messages{*text.NewValidationErrorRequired("name")}},
					},
				},
			},
		} {
			t.Run(fmt.Sprintf("case=%d", k), func(t *testing.T) {
				actual, err := NewHTMLFormFromRequestBody(tc.r, "action", decoderx.HTTPJSONSchemaCompiler(tc.ref, nil))
				require.NoError(t, err)
				// sort actual.fields lexicographically to have a deterministic order
				sort.SliceStable(actual.Fields, func(i, j int) bool {
					return actual.Fields[i].Name < actual.Fields[j].Name
				})
				assert.Equal(t, "action", actual.Action)
				assert.EqualValues(t, tc.expect.Fields, actual.Fields)
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
					Fields: Fields{
						Field{Name: "numby", Type: "number"},
						Field{Name: "objy.numby", Type: "number"},
						Field{Name: "objy.objy", Type: "text"},
						Field{Name: "objy.stringy", Type: "text"},
						Field{Name: "stringy", Type: "text"},
					},
				},
			},
			{
				ref:    "./stub/simple.schema.json",
				prefix: "traits",
				expect: &HTMLForm{
					Fields: Fields{
						Field{Name: "traits.numby", Type: "number"},
						Field{Name: "traits.objy.numby", Type: "number"},
						Field{Name: "traits.objy.objy", Type: "text"},
						Field{Name: "traits.objy.stringy", Type: "text"},
						Field{Name: "traits.stringy", Type: "text"},
					},
				},
			},
			{
				ref: "./stub/complex.schema.json",
				expect: &HTMLForm{
					Fields: Fields{
						Field{Name: "fruits", Type: "text"},
						Field{Name: "meal.chef", Type: "text"},
						Field{Name: "meal.name", Type: "text"},
						Field{Name: "vegetables", Type: "text"},
					},
				},
			},
		} {
			t.Run(fmt.Sprintf("case=%d", k), func(t *testing.T) {
				actual, err := NewHTMLFormFromJSONSchema("action", tc.ref, tc.prefix, nil)
				require.NoError(t, err)
				assert.Equal(t, "action", actual.Action)
				assert.EqualValues(t, tc.expect.Messages, actual.Messages)
				assert.EqualValues(t, tc.expect.Fields, actual.Fields)
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
			{err: herodot.ErrBadRequest.WithReason("tests"), expect: HTMLForm{Fields: Fields{}, Messages: text.Messages{*text.NewValidationErrorGeneric("tests")}}},
			{err: schema.NewInvalidCredentialsError(), expect: HTMLForm{Fields: Fields{}, Messages: text.Messages{*text.NewErrorValidationInvalidCredentials()}}},
			{err: &jsonschema.ValidationError{Message: "test", InstancePtr: "#/foo/bar/baz"}, expect: HTMLForm{Fields: Fields{Field{Name: "foo.bar.baz", Type: "", Messages: text.Messages{*text.NewValidationErrorGeneric("test")}}}}},
			{err: &jsonschema.ValidationError{Message: "test", InstancePtr: ""}, expect: HTMLForm{Fields: Fields{}, Messages: text.Messages{*text.NewValidationErrorGeneric("test")}}},
		} {
			t.Run(fmt.Sprintf("case=%d", k), func(t *testing.T) {
				for _, in := range []error{tc.err, errors.WithStack(tc.err)} {
					c := NewHTMLForm("")
					err := c.ParseError(in)
					if tc.expectErr {
						require.Error(t, err)
						return
					}
					require.NoError(t, err)
					assert.EqualValues(t, tc.expect.Messages, c.Messages)
					assert.EqualValues(t, tc.expect.Fields, c.Fields)
				}
			})
		}
	})

	t.Run("method=SetValue", func(t *testing.T) {
		c := HTMLForm{
			Fields: Fields{
				{Name: "1", Value: "foo"},
				{Name: "2", Value: ""},
			},
		}

		assert.Len(t, c.Fields, 2)

		c.SetValue("1", "baz1")
		c.SetValue("2", "baz2")
		c.SetValue("3", "baz3")

		assert.Len(t, c.Fields, 3)
		for _, k := range []string{"1", "2", "3"} {
			assert.EqualValues(t, fmt.Sprintf("baz%s", k), c.getField(k).Value, "%+v", c)
		}
	})

	t.Run("method=SetCSRF", func(t *testing.T) {
		f := &HTMLForm{Fields: Fields{{Name: "1", Value: "bar"}}}
		f.SetCSRF("csrf-token")
		assert.Contains(
			t,
			f.Fields,
			Field{Name: CSRFTokenName, Value: "csrf-token", Type: "hidden", Required: true},
		)

		f = &HTMLForm{Fields: Fields{{Name: "1", Value: "bar"}}}
		f.SetCSRF("csrf-token")
		assert.Contains(
			t,
			f.Fields,
			Field{Name: CSRFTokenName, Value: "csrf-token", Type: "hidden", Required: true},
		)
	})

	t.Run("method=AddMessage", func(t *testing.T) {
		c := HTMLForm{
			Fields: Fields{
				{Name: "1", Value: "foo", Messages: text.Messages{{Text: "foo"}}},
				{Name: "2", Value: "", Messages: text.Messages{}},
			},
		}
		assert.Len(t, c.Fields, 2)
		c.AddMessage(&text.Message{Text: "baz1"}, "1")
		c.AddMessage(&text.Message{Text: "baz2"}, "2")
		c.AddMessage(&text.Message{Text: "baz3"}, "3")
		c.AddMessage(&text.Message{Text: "baz"}, "4", "5", "6")
		c.AddMessage(&text.Message{Text: "rootbar"})

		assert.Len(t, c.Fields, 6)
		for _, k := range []string{"1", "2", "3"} {
			assert.EqualValues(t, fmt.Sprintf("baz%s", k), c.getField(k).Messages[len(c.getField(k).Messages)-1].Text, "%+v", c)
		}
		for _, k := range []string{"4", "5", "6"} {
			assert.EqualValues(t, "baz", c.getField(k).Messages[0].Text, "%+v", c)
		}

		assert.Len(t, c.Messages, 1)
		assert.Equal(t, "rootbar", c.Messages[0].Text)
	})

	t.Run("method=Reset", func(t *testing.T) {
		c := HTMLForm{
			Fields: Fields{
				{Name: "1", Value: "foo", Messages: text.Messages{{Text: "foo"}}},
				{Name: "2", Value: "bar", Messages: text.Messages{{Text: "bar"}}},
			},
			Messages: text.Messages{{Text: ""}},
		}
		c.Reset()

		assert.Empty(t, c.Messages)
		assert.Empty(t, c.getField("1").Messages)
		assert.Empty(t, c.getField("1").Value)
		assert.Empty(t, c.getField("2").Messages)
		assert.Empty(t, c.getField("2").Value)
	})
}
