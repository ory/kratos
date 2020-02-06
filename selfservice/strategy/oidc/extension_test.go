package oidc_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/jsonschema/v3"

	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/schema"
	"github.com/ory/kratos/selfservice/strategy/oidc"
)

func TestValidationExtensionRunner(t *testing.T) {
	for k, tc := range []struct {
		expectErr error
		schema    string
		doc       string
		expect    string
	}{
		{
			doc:    "./stub/extension/payload.json",
			schema: "file://stub/extension/schema.json",
			expect: `{"email": "someone@email.org","names": ["peter","pan"]}`,
		},
	} {
		t.Run(fmt.Sprintf("case=%d", k), func(t *testing.T) {
			c := jsonschema.NewCompiler()
			runner, err := schema.NewExtensionRunner(schema.ExtensionRunnerOIDCMetaSchema)
			require.NoError(t, err)

			var i identity.Identity
			runner.
				AddRunner(oidc.NewValidationExtensionRunner(&i).Runner).
				Register(c)

			doc, err := os.Open(tc.doc)
			require.NoError(t, err)

			err = c.MustCompile(tc.schema).Validate(doc)
			if tc.expectErr != nil {
				require.EqualError(t, err, tc.expectErr.Error())
			}

			assert.JSONEq(t, tc.expect, string(i.Traits))
		})
	}
}
