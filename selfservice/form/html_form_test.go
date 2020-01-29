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
	"github.com/santhosh-tekuri/jsonschema/v2"

	"github.com/ory/herodot"
	"github.com/ory/x/decoderx"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/schema"
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
						Field{Name: "stringy", Type: "text", Value: "foobar"},
						Field{Name: "objy.numby", Type: "number", Value: 1.5},
						Field{Name: "objy.stringy", Type: "text", Value: "foobar"},
					},
				},
			},
			{
				r:      `{"numby":1.5,"stringy":"foobar","objy":{"objy":{},"numby":1.5,"stringy":"foobar"}}`,
				prefix: "traits",
				expect: &HTMLForm{
					Fields: Fields{
						Field{Name: "traits.numby", Type: "number", Value: 1.5},
						Field{Name: "traits.stringy", Type: "text", Value: "foobar"},
						Field{Name: "traits.objy.numby", Type: "number", Value: 1.5},
						Field{Name: "traits.objy.stringy", Type: "text", Value: "foobar"},
					},
				},
			},
		} {
			t.Run(fmt.Sprintf("case=%d", k), func(t *testing.T) {
				actual := NewHTMLFormFromJSON("action", json.RawMessage(tc.r), tc.prefix)
				sort.Sort(tc.expect.Fields)
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
						Field{Name: "stringy", Type: "text", Value: "foobar"},
						Field{Name: "objy.numby", Type: "number", Value: 1.5},
						Field{Name: "objy.stringy", Type: "text", Value: "foobar"},
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
						Field{Name: "stringy", Type: "text", Value: "foobar"},
						Field{Name: "objy.numby", Type: "number", Value: 1.5},
						Field{Name: "objy.stringy", Type: "text", Value: "foobar"},
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
						Field{Name: "meal.name", Errors: []Error{{Message: "missing properties: \"name\""}}},
						Field{Name: "meal.chef", Type: "text", Value: "aeneas"},
					},
				},
			},
		} {
			t.Run(fmt.Sprintf("case=%d", k), func(t *testing.T) {
				actual, err := NewHTMLFormFromRequestBody(tc.r, "action", decoderx.HTTPJSONSchemaCompiler(tc.ref, nil))
				require.NoError(t, err)
				sort.Sort(tc.expect.Fields)
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
				actual, err := NewHTMLFormFromJSONSchema("action", tc.ref, tc.prefix)
				require.NoError(t, err)
				sort.Sort(tc.expect.Fields)
				assert.Equal(t, "action", actual.Action)
				assert.EqualValues(t, tc.expect.Errors, actual.Errors)
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
			{err: herodot.ErrBadRequest.WithReason("tests"), expect: HTMLForm{Fields: Fields{}, Errors: []Error{{Message: "tests"}}}},
			{err: schema.NewInvalidCredentialsError(), expect: HTMLForm{Fields: Fields{}, Errors: []Error{{Message: "The provided credentials are invalid. Check for spelling mistakes in your password or username, email address, or phone number."}}}},
			{err: &jsonschema.ValidationError{Message: "test", InstancePtr: "#/foo/bar/baz"}, expect: HTMLForm{Fields: Fields{Field{Name: "foo.bar.baz", Type: "", Errors: []Error{{Message: "test"}}}}}},
			{err: &jsonschema.ValidationError{Message: "test", InstancePtr: ""}, expect: HTMLForm{Fields: Fields{}, Errors: []Error{{Message: "test"}}}},
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
					sort.Sort(tc.expect.Fields)
					assert.EqualValues(t, tc.expect.Errors, c.Errors)
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

	t.Run("method=AddError", func(t *testing.T) {
		c := HTMLForm{
			Fields: Fields{
				{Name: "1", Value: "foo", Errors: []Error{{Message: "foo"}}},
				{Name: "2", Value: "", Errors: []Error{}},
			},
		}
		assert.Len(t, c.Fields, 2)
		c.AddError(&Error{Message: "baz1"}, "1")
		c.AddError(&Error{Message: "baz2"}, "2")
		c.AddError(&Error{Message: "baz3"}, "3")
		c.AddError(&Error{Message: "baz"}, "4", "5", "6")
		c.AddError(&Error{Message: "rootbar"})

		assert.Len(t, c.Fields, 6)
		for _, k := range []string{"1", "2", "3"} {
			assert.EqualValues(t, fmt.Sprintf("baz%s", k), c.getField(k).Errors[len(c.getField(k).Errors)-1].Message, "%+v", c)
		}
		for _, k := range []string{"4", "5", "6"} {
			assert.EqualValues(t, "baz", c.getField(k).Errors[0].Message, "%+v", c)
		}

		assert.Len(t, c.Errors, 1)
		assert.Equal(t, "rootbar", c.Errors[0].Message)
	})

	t.Run("method=Reset", func(t *testing.T) {
		c := HTMLForm{
			Fields: Fields{
				{Name: "1", Value: "foo", Errors: []Error{{Message: "foo"}}},
				{Name: "2", Value: "bar", Errors: []Error{{Message: "bar"}}},
			},
			Errors: []Error{{Message: ""}},
		}
		c.Reset()

		assert.Empty(t, c.Errors)
		assert.Empty(t, c.getField("1").Errors)
		assert.Empty(t, c.getField("1").Value)
		assert.Empty(t, c.getField("2").Errors)
		assert.Empty(t, c.getField("2").Value)
	})
}
