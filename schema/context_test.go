package schema

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ory/gojsonschema"
)

func TestContextRemoveRootStub(t *testing.T) {
	assert.Equal(t, "bar.baz",
		ContextRemoveRootStub(
			gojsonschema.NewJsonContext(
				"(root)",
				gojsonschema.NewJsonContext(
					"baz",
					gojsonschema.NewJsonContext(
						"bar",
						nil,
					),
				),
			),
		).String("."),
	)
}

func TestContextSetRoot(t *testing.T) {
	assert.Equal(t, "new_root.baz.bar.foo",
		ContextSetRoot(
			gojsonschema.NewJsonContext(
				"foo",
				gojsonschema.NewJsonContext(
					"bar",
					gojsonschema.NewJsonContext(
						"baz",
						nil,
					),
				),
			),
			"new_root",
		).String("."),
	)
}
