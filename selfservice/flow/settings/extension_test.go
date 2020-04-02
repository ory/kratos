package settings

import (
	"fmt"
	"testing"

	"github.com/ory/kratos/selfservice/form"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/jsonschema/v3"
	"github.com/ory/x/jsonschemax"
)

func TestDisableIdentifiersExtension(t *testing.T) {
	for k, tc := range []struct {
		schema                  string
		expectDisabledPathNames []string
	}{
		{
			schema:                  "file://./stub/identity.schema.json",
			expectDisabledPathNames: []string{"email"},
		},
	} {
		t.Run(fmt.Sprintf("case=%d", k), func(t *testing.T) {
			c := jsonschema.NewCompiler()
			registerNewDisableIdentifiersExtension(c)
			paths, err := jsonschemax.ListPaths(tc.schema, c)
			require.NoError(t, err)

			t.Logf("%+v", paths)
			var disabledPathNames []string
			for _, p := range paths {
				if disabled, ok := p.CustomProperties[form.DisableFormField]; ok {
					if d, ok := disabled.(bool); ok && d {
						disabledPathNames = append(disabledPathNames, p.Name)
					}
				}
			}

			assert.Equal(t, tc.expectDisabledPathNames, disabledPathNames)
		})
	}
}
