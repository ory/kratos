package schema

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/ory/jsonschema/v3"
	_ "github.com/ory/jsonschema/v3/fileloader"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExtensionRunner(t *testing.T) {
	for k, tc := range []struct {
		expectErr error
		schema    string
		doc       string
		expect    []string
	}{
		{
			doc:    `{"email":"foo@ory.sh"}`,
			schema: "file://./stub/extension/schema.json",
			expect: []string{"foo@ory.sh"},
		},
		{
			doc:    `{"emails":["foo@ory.sh","bar@ory.sh"]}`,
			schema: "file://./stub/extension/schema.nested.json",
			expect: []string{"foo@ory.sh","bar@ory.sh"},
		},
	} {
		t.Run(fmt.Sprintf("case=%d", k), func(t *testing.T) {
			c := jsonschema.NewCompiler()
			runner, err := NewExtensionRunner(ExtensionRunnerIdentityMetaSchema)
			require.NoError(t, err)

			var identifiers []string
			runner.AddRunner(func(ctx jsonschema.ValidationContext, config ExtensionConfig, value interface{}) error {
				if config.Credentials.Password.Identifier {
					identifiers = append(identifiers, fmt.Sprintf("%s", value))
				}
				return nil
			}).Register(c)

			err = c.MustCompile(tc.schema).Validate(bytes.NewBufferString(tc.doc))
			if tc.expectErr != nil {
				require.EqualError(t, err, tc.expectErr.Error())
			}

			assert.EqualValues(t, tc.expect, identifiers)
		})
	}
}
