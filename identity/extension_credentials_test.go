package identity_test

import (
	"bytes"
	"context"
	"fmt"
	"testing"

	"github.com/ory/jsonschema/v3"
	_ "github.com/ory/jsonschema/v3/fileloader"

	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/schema"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var ctx = context.Background()

func TestSchemaExtensionCredentials(t *testing.T) {
	for k, tc := range []struct {
		expectErr error
		schema    string
		doc       string
		expect    []string
		existing  *identity.Credentials
		ct        identity.CredentialsType
	}{
		{
			doc:    `{"email":"foo@ory.sh"}`,
			schema: "file://./stub/extension/credentials/schema.json",
			expect: []string{"foo@ory.sh"},
			ct:     identity.CredentialsTypePassword,
		},
		{
			doc:    `{"emails":["foo@ory.sh","foo@ory.sh","bar@ory.sh"], "username": "foobar"}`,
			schema: "file://./stub/extension/credentials/multi.schema.json",
			expect: []string{"foo@ory.sh", "bar@ory.sh", "foobar"},
			ct:     identity.CredentialsTypePassword,
		},
		{
			doc:    `{"emails":["FOO@ory.sh","bar@ory.sh"], "username": "foobar"}`,
			schema: "file://./stub/extension/credentials/multi.schema.json",
			expect: []string{"foo@ory.sh", "bar@ory.sh", "foobar"},
			existing: &identity.Credentials{
				Identifiers: []string{"not-foo@ory.sh"},
			},
			ct: identity.CredentialsTypePassword,
		},
		{
			doc:    `{"email":"foo@ory.sh"}`,
			schema: "file://./stub/extension/credentials/webauthn.schema.json",
			expect: []string{"foo@ory.sh"},
			ct:     identity.CredentialsTypeWebAuthn,
		},
		{
			doc:    `{"email":"FOO@ory.sh"}`,
			schema: "file://./stub/extension/credentials/webauthn.schema.json",
			expect: []string{"foo@ory.sh"},
			existing: &identity.Credentials{
				Identifiers: []string{"not-foo@ory.sh"},
			},
			ct: identity.CredentialsTypeWebAuthn,
		},
	} {
		t.Run(fmt.Sprintf("case=%d", k), func(t *testing.T) {
			c := jsonschema.NewCompiler()
			runner, err := schema.NewExtensionRunner(ctx)
			require.NoError(t, err)

			i := new(identity.Identity)
			e := identity.NewSchemaExtensionCredentials(i)
			if tc.existing != nil {
				i.SetCredentials(tc.ct, *tc.existing)
			}

			runner.AddRunner(e).Register(c)
			err = c.MustCompile(ctx, tc.schema).Validate(bytes.NewBufferString(tc.doc))
			if tc.expectErr != nil {
				require.EqualError(t, err, tc.expectErr.Error())
			}
			require.NoError(t, e.Finish())

			credentials, ok := i.GetCredentials(tc.ct)
			require.True(t, ok)
			assert.ElementsMatch(t, tc.expect, credentials.Identifiers)
		})
	}
}
