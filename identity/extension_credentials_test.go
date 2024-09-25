// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package identity_test

import (
	"bytes"
	"context"
	"fmt"
	"testing"

	"github.com/ory/x/sqlxx"

	"github.com/ory/x/snapshotx"

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
		expectErr           error
		schema              string
		doc                 string
		expectedIdentifiers []string
		existing            *identity.Credentials
		ct                  identity.CredentialsType
	}{
		{
			doc:                 `{"email":"foo@ory.sh"}`,
			schema:              "file://./stub/extension/credentials/schema.json",
			expectedIdentifiers: []string{"foo@ory.sh"},
			ct:                  identity.CredentialsTypePassword,
		},
		{
			doc:                 `{"emails":["foo@ory.sh","foo@ory.sh","bar@ory.sh"], "username": "foobar"}`,
			schema:              "file://./stub/extension/credentials/multi.schema.json",
			expectedIdentifiers: []string{"foo@ory.sh", "bar@ory.sh", "foobar"},
			ct:                  identity.CredentialsTypePassword,
		},
		{
			doc:                 `{"emails":["foo@ory.sh","foo@ory.sh","bar@ory.sh"], "username": "foobar"}`,
			schema:              "file://./stub/extension/credentials/multi.schema.json",
			expectedIdentifiers: []string{"foo@ory.sh", "bar@ory.sh"},
			ct:                  identity.CredentialsTypeWebAuthn,
		},
		{
			doc:                 `{"emails":["FOO@ory.sh","bar@ory.sh"], "username": "foobar"}`,
			schema:              "file://./stub/extension/credentials/multi.schema.json",
			expectedIdentifiers: []string{"foo@ory.sh", "bar@ory.sh", "foobar"},
			existing: &identity.Credentials{
				Identifiers: []string{"not-foo@ory.sh"},
			},
			ct: identity.CredentialsTypePassword,
		},
		{
			doc:                 `{"email":"foo@ory.sh"}`,
			schema:              "file://./stub/extension/credentials/webauthn.schema.json",
			expectedIdentifiers: []string{"foo@ory.sh"},
			ct:                  identity.CredentialsTypeWebAuthn,
		},
		{
			doc:                 `{"email":"FOO@ory.sh"}`,
			schema:              "file://./stub/extension/credentials/webauthn.schema.json",
			expectedIdentifiers: []string{"foo@ory.sh"},
			existing: &identity.Credentials{
				Identifiers: []string{"not-foo@ory.sh"},
			},
			ct: identity.CredentialsTypeWebAuthn,
		},
		{
			doc:                 `{"email":"foo@ory.sh"}`,
			schema:              "file://./stub/extension/credentials/code.schema.json",
			expectedIdentifiers: []string{"foo@ory.sh"},
			ct:                  identity.CredentialsTypeCodeAuth,
		},
		{
			doc:                 `{"email":"FOO@ory.sh"}`,
			schema:              "file://./stub/extension/credentials/code.schema.json",
			expectedIdentifiers: []string{"foo@ory.sh"},
			existing: &identity.Credentials{
				Identifiers: []string{"not-foo@ory.sh"},
			},
			ct: identity.CredentialsTypeCodeAuth,
		},
		{
			doc:                 `{"email":"FOO@ory.sh"}`,
			schema:              "file://./stub/extension/credentials/code.schema.json",
			expectedIdentifiers: []string{"foo@ory.sh"},
			existing: &identity.Credentials{
				Identifiers: []string{"not-foo@ory.sh", "foo@ory.sh"},
				Config:      sqlxx.JSONRawMessage(`{"addresses":[{"channel":"email","address":"not-foo@ory.sh"}]}`),
			},
			ct: identity.CredentialsTypeCodeAuth,
		},
		{
			doc:                 `{"email":"FOO@ory.sh","phone":"+49 176 671 11 638"}`,
			schema:              "file://./stub/extension/credentials/code-phone-email.schema.json",
			expectedIdentifiers: []string{"+4917667111638", "foo@ory.sh"},
			existing: &identity.Credentials{
				Identifiers: []string{"not-foo@ory.sh", "foo@ory.sh"},
				Config:      sqlxx.JSONRawMessage(`{"addresses":[{"channel":"email","address":"not-foo@ory.sh"}]}`),
			},
			ct: identity.CredentialsTypeCodeAuth,
		},
		{
			doc:                 `{"email":"FOO@ory.sh","phone":"+49 176 671 11 638"}`,
			schema:              "file://./stub/extension/credentials/code-phone-email.schema.json",
			expectedIdentifiers: []string{"+4917667111638", "foo@ory.sh"},
			existing: &identity.Credentials{
				Identifiers: []string{"not-foo@ory.sh", "foo@ory.sh"},
				Config:      sqlxx.JSONRawMessage(`{"addresses":[{"channel":"email","address":"not-foo@ory.sh"}]}`),
			},
			ct: identity.CredentialsTypeCodeAuth,
		},
		{
			doc:                 `{"email":"FOO@ory.sh","email2":"FOO@ory.sh","phone":"+49 176 671 11 638"}`,
			schema:              "file://./stub/extension/credentials/code-phone-email.schema.json",
			expectedIdentifiers: []string{"+4917667111638", "foo@ory.sh"},
			existing: &identity.Credentials{
				Identifiers: []string{"not-foo@ory.sh", "fOo@ory.sh"},
				Config:      sqlxx.JSONRawMessage(`{"addresses":[{"channel":"email","address":"not-foo@ory.sh"}]}`),
			},
			ct: identity.CredentialsTypeCodeAuth,
		},
		{
			doc:                 `{"email":"FOO@ory.sh","email2":"FOO@ory.sh","email3":"bar@ory.sh","phone":"+49 176 671 11 638"}`,
			schema:              "file://./stub/extension/credentials/code-phone-email.schema.json",
			expectedIdentifiers: []string{"+4917667111638", "foo@ory.sh", "bar@ory.sh"},
			existing: &identity.Credentials{
				Identifiers: []string{"not-foo@ory.sh", "fOo@ory.sh"},
				Config:      sqlxx.JSONRawMessage(`{"addresses":[{"channel":"email","address":"not-foo@ory.sh"}]}`),
			},
			ct: identity.CredentialsTypeCodeAuth,
		},
		{
			doc:                 `{"email":"bar@ory.sh"}`,
			schema:              "file://./stub/extension/credentials/email.schema.json",
			expectedIdentifiers: []string{"bar@ory.sh"},
			existing: &identity.Credentials{
				Identifiers: []string{"foo@ory.sh"},
				Config:      sqlxx.JSONRawMessage(`{"addresses":[{"channel":"email","address":"foo@ory.sh"}]}`),
			},
			ct: identity.CredentialsTypeCodeAuth,
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
			} else {
				require.NoError(t, err)
			}
			require.NoError(t, e.Finish())

			credentials, ok := i.GetCredentials(tc.ct)
			require.True(t, ok)
			assert.ElementsMatch(t, tc.expectedIdentifiers, credentials.Identifiers)
			snapshotx.SnapshotT(t, credentials, snapshotx.ExceptPaths("identifiers"))
		})
	}
}
