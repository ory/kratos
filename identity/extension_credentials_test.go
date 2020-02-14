package identity_test

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/ory/jsonschema/v3"
	_ "github.com/ory/jsonschema/v3/fileloader"

	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/schema"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSchemaExtensionCredentials(t *testing.T) {
	for k, tc := range []struct {
		expectErr error
		schema    string
		doc       string
		expect    []string
		existing  *identity.Credentials
	}{
		{
			doc:    `{"email":"foo@ory.sh"}`,
			schema: "file://./stub/extension/credentials/schema.json",
			expect: []string{"foo@ory.sh"},
		},
		{
			doc:    `{"emails":["foo@ory.sh","foo@ory.sh","bar@ory.sh"], "username": "foobar"}`,
			schema: "file://./stub/extension/credentials/multi.schema.json",
			expect: []string{"foo@ory.sh", "bar@ory.sh", "foobar"},
		},
		{
			doc:    `{"emails":["foo@ory.sh","bar@ory.sh"], "username": "foobar"}`,
			schema: "file://./stub/extension/credentials/multi.schema.json",
			expect: []string{"foo@ory.sh", "bar@ory.sh", "foobar"},
			existing: &identity.Credentials{
				Identifiers: []string{"not-foo@ory.sh"},
			},
		},
	} {
		t.Run(fmt.Sprintf("case=%d", k), func(t *testing.T) {
			c := jsonschema.NewCompiler()
			runner, err := schema.NewExtensionRunner(schema.ExtensionRunnerIdentityMetaSchema)
			require.NoError(t, err)

			i := new(identity.Identity)
			e := identity.NewSchemaExtensionCredentials(i)
			if tc.existing != nil {
				i.SetCredentials(identity.CredentialsTypePassword, *tc.existing)
			}

			runner.AddRunner(e).Register(c)
			err = c.MustCompile(tc.schema).Validate(bytes.NewBufferString(tc.doc))
			if tc.expectErr != nil {
				require.EqualError(t, err, tc.expectErr.Error())
			}
			require.NoError(t, e.Finish())

			credentials, ok := i.GetCredentials(identity.CredentialsTypePassword)
			require.True(t, ok)
			assert.ElementsMatch(t, tc.expect, credentials.Identifiers)
		})
	}
}
