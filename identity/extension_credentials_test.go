// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package identity_test

import (
	"bytes"
	"context"
	"errors"
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
			doc:    `{}`,
			schema: "file://./stub/extension/credentials/schema.json",
			expect: []string{},
			existing: &identity.Credentials{
				Identifiers: []string{"foo@ory.sh"},
			},
			ct: identity.CredentialsTypePassword,
		},
		{
			doc:    `{"emails":["foo@ory.sh","foo@ory.sh","bar@ory.sh"], "username": "foobar"}`,
			schema: "file://./stub/extension/credentials/multi.schema.json",
			expect: []string{"foo@ory.sh", "bar@ory.sh", "foobar"},
			ct:     identity.CredentialsTypePassword,
		},
		{
			doc:    `{"emails":["foo@ory.sh","foo@ory.sh","bar@ory.sh"], "username": "foobar"}`,
			schema: "file://./stub/extension/credentials/multi.schema.json",
			expect: []string{"foo@ory.sh", "bar@ory.sh"},
			ct:     identity.CredentialsTypeWebAuthn,
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
		{
			doc:    `{"email":"foo@ory.sh"}`,
			schema: "file://./stub/extension/credentials/code.schema.json",
			expect: []string{"foo@ory.sh"},
			ct:     identity.CredentialsTypeCodeAuth,
		},
		{
			doc:    `{"email":"FOO@ory.sh"}`,
			schema: "file://./stub/extension/credentials/code.schema.json",
			expect: []string{"foo@ory.sh"},
			existing: &identity.Credentials{
				Identifiers: []string{"not-foo@ory.sh"},
			},
			ct: identity.CredentialsTypeCodeAuth,
		},
		{
			doc:       `{"phone":"not-valid-number"}`,
			schema:    "file://./stub/extension/credentials/code.schema.json",
			ct:        identity.CredentialsTypeCodeAuth,
			expectErr: errors.New("I[#/phone] S[#/properties/phone] validation failed\n  I[#/phone] S[#/properties/phone/format] \"not-valid-number\" is not valid \"tel\"\n  I[#/phone] S[#/properties/phone/format] the phone number supplied is not a number"),
		},
		{
			doc:    `{"phone":"+4407376494399"}`,
			schema: "file://./stub/extension/credentials/code.schema.json",
			expect: []string{"+447376494399"},
			ct:     identity.CredentialsTypeCodeAuth,
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

			if tc.expectErr == nil {
				credentials, ok := i.GetCredentials(tc.ct)
				require.True(t, ok)
				assert.ElementsMatch(t, tc.expect, credentials.Identifiers)
			}
		})
	}
}
