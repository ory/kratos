package identity_test

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/ory/jsonschema/v3"
	_ "github.com/ory/jsonschema/v3/fileloader"

	"github.com/ory/kratos/schema"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	. "github.com/ory/kratos/identity"
)

func TestExtensionRunnerIdentityTraits(t *testing.T) {
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
			doc:    `{"emails":["foo@ory.sh","bar@ory.sh"], "username": "foobar"}`,
			schema: "file://./stub/extension/multi.schema.json",
			expect: []string{"foo@ory.sh", "bar@ory.sh", "foobar"},
		},
	} {
		t.Run(fmt.Sprintf("case=%d", k), func(t *testing.T) {
			c := jsonschema.NewCompiler()
			runner, err := schema.NewExtensionRunner(schema.ExtensionRunnerIdentityMetaSchema)
			require.NoError(t, err)

			i := new(Identity)
			e := NewValidationExtensionRunner(i)
			runner.AddRunner(e.Runner).Register(c)

			err = c.MustCompile(tc.schema).Validate(bytes.NewBufferString(tc.doc))
			if tc.expectErr != nil {
				require.EqualError(t, err, tc.expectErr.Error())
			}

			credentials, ok := i.GetCredentials(CredentialsTypePassword)
			require.True(t, ok)
			assert.ElementsMatch(t, tc.expect, credentials.Identifiers)
		})
	}
}
