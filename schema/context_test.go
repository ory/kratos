package schema

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ory/jsonschema/v3"
)

func TestContextSetRoot(t *testing.T) {
	for k, tc := range []struct {
		in     jsonschema.ValidationError
		head   string
		expect jsonschema.ValidationError
	}{
		{
			head: "traits",
			in: jsonschema.ValidationError{
				InstancePtr: "#/foo/bar",
				Context:     &jsonschema.ValidationErrorContextRequired{Missing: []string{"#/foo/bar/baz"}},
				Causes: []*jsonschema.ValidationError{{
					InstancePtr: "#/foo/bar",
					Context:     &jsonschema.ValidationErrorContextRequired{Missing: []string{"#/foo/bar/baz"}},
				}},
			},
			expect: jsonschema.ValidationError{
				InstancePtr: "#/traits/foo/bar",
				Context:     &jsonschema.ValidationErrorContextRequired{Missing: []string{"#/traits/foo/bar/baz"}},
				Causes: []*jsonschema.ValidationError{{
					InstancePtr: "#/traits/foo/bar",
					Context:     &jsonschema.ValidationErrorContextRequired{Missing: []string{"#/traits/foo/bar/baz"}},
				}},
			},
		},
	} {
		t.Run(fmt.Sprintf("case=%d", k), func(t *testing.T) {
			assert.EqualValues(t, tc.expect, *ContextSetRoot(&tc.in, tc.head))
		})
	}
}
