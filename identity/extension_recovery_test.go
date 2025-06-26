// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package identity

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/ory/jsonschema/v3"
	_ "github.com/ory/jsonschema/v3/fileloader"

	"github.com/ory/kratos/schema"
	"github.com/ory/kratos/x"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSchemaExtensionRecovery(t *testing.T) {
	iid := x.NewUUID()
	for k, tc := range []struct {
		expectErr   error
		schema      string
		doc         string
		expect      []RecoveryAddress
		existing    []RecoveryAddress
		description string
	}{
		{
			description: "valid email, no existing",
			doc:         `{"username":"foo@ory.sh"}`,
			schema:      "file://./stub/extension/recovery/email.schema.json",
			expect: []RecoveryAddress{
				{
					Value:      "foo@ory.sh",
					Via:        RecoveryAddressTypeEmail,
					IdentityID: iid,
				},
			},
		},
		{
			description: "valid email, some existing, no overlap",
			doc:         `{"username":"foo@ory.sh"}`,
			schema:      "file://./stub/extension/recovery/email.schema.json",
			expect: []RecoveryAddress{
				{
					Value:      "foo@ory.sh",
					Via:        RecoveryAddressTypeEmail,
					IdentityID: iid,
				},
			},
			existing: []RecoveryAddress{
				{
					Value:      "bar@ory.sh",
					Via:        RecoveryAddressTypeEmail,
					IdentityID: iid,
				},
			},
		},
		{
			description: "valid emails, some existing, overlap",
			doc:         `{"emails":["baz@ory.sh","foo@ory.sh"]}`,
			schema:      "file://./stub/extension/recovery/email.schema.json",
			expect: []RecoveryAddress{
				{
					Value:      "foo@ory.sh",
					Via:        RecoveryAddressTypeEmail,
					IdentityID: iid,
				},
				{
					Value:      "baz@ory.sh",
					Via:        RecoveryAddressTypeEmail,
					IdentityID: iid,
				},
			},
			existing: []RecoveryAddress{
				{
					Value:      "foo@ory.sh",
					Via:        RecoveryAddressTypeEmail,
					IdentityID: iid,
				},
				{
					Value:      "bar@ory.sh",
					Via:        RecoveryAddressTypeEmail,
					IdentityID: iid,
				},
			},
		},
		{
			description: "valid emails, no existing, overlap",
			doc:         `{"emails":["foo@ory.sh","foo@ory.sh","baz@ory.sh"]}`,
			schema:      "file://./stub/extension/recovery/email.schema.json",
			expect: []RecoveryAddress{
				{
					Value:      "foo@ory.sh",
					Via:        RecoveryAddressTypeEmail,
					IdentityID: iid,
				},
				{
					Value:      "baz@ory.sh",
					Via:        RecoveryAddressTypeEmail,
					IdentityID: iid,
				},
			},
			existing: []RecoveryAddress{
				{
					Value:      "foo@ory.sh",
					Via:        RecoveryAddressTypeEmail,
					IdentityID: iid,
				},
				{
					Value:      "bar@ory.sh",
					Via:        RecoveryAddressTypeEmail,
					IdentityID: iid,
				},
			},
		},
		{
			description: "invalid email",
			doc:         `{"emails":["foo@ory.sh","bar@ory.sh"], "username": "foobar"}`,
			schema:      "file://./stub/extension/recovery/email.schema.json",
			expectErr:   errors.New("I[#/username] S[#/properties/username/format] \"foobar\" is not valid \"email\""),
		},
		{
			description: "valid emails, no existing",
			doc:         `{"emails":["foo@ory.sh","bar@ory.sh","bar@ory.sh"], "username": "foobar@ory.sh"}`,
			schema:      "file://./stub/extension/recovery/email.schema.json",
			expect: []RecoveryAddress{
				{
					Value:      "foo@ory.sh",
					Via:        RecoveryAddressTypeEmail,
					IdentityID: iid,
				},
				{
					Value:      "bar@ory.sh",
					Via:        RecoveryAddressTypeEmail,
					IdentityID: iid,
				},
				{
					Value:      "foobar@ory.sh",
					Via:        RecoveryAddressTypeEmail,
					IdentityID: iid,
				},
			},
		},

		{
			description: "valid phone number, no existing",
			doc:         `{"telephoneNumber":"+68672098006"}`,
			schema:      "file://./stub/extension/recovery/sms.schema.json",
			expect: []RecoveryAddress{
				{
					Value:      "+68672098006",
					Via:        RecoveryAddressTypeSMS,
					IdentityID: iid,
				},
			},
		},
		{
			description: "valid phone number, some existing, no overlap",
			doc:         `{"telephoneNumber":"+68672098006"}`,
			schema:      "file://./stub/extension/recovery/sms.schema.json",
			expect: []RecoveryAddress{
				{
					Value:      "+68672098006",
					Via:        RecoveryAddressTypeSMS,
					IdentityID: iid,
				},
			},
			existing: []RecoveryAddress{
				{
					Value:      "+12 345 67890123",
					Via:        RecoveryAddressTypeSMS,
					IdentityID: iid,
				},
			},
		},
		{
			description: "valid phone number, some existing, overlap",
			doc:         `{"telephoneNumber":"+68672098006"}`,
			schema:      "file://./stub/extension/recovery/sms.schema.json",
			expect: []RecoveryAddress{
				{
					Value:      "+68672098006",
					Via:        RecoveryAddressTypeSMS,
					IdentityID: iid,
				},
			},
			existing: []RecoveryAddress{
				{
					Value:      "+68672098006",
					Via:        RecoveryAddressTypeSMS,
					IdentityID: iid,
				},
				{
					Value:      "+33 07856952",
					Via:        RecoveryAddressTypeSMS,
					IdentityID: iid,
				},
			},
		},
		{
			description: "invalid phone number",
			doc:         `{"telephoneNumber": "foobar"}`,
			schema:      "file://./stub/extension/recovery/sms.schema.json",
			// We get 2 errors: one from the JSON schema `format` validation and one from the Go validation.
			expectErr: errors.New("I[#/telephoneNumber] S[#/properties/telephoneNumber] validation failed\n  I[#/telephoneNumber] S[#/properties/telephoneNumber/format] \"foobar\" is not valid \"tel\"\n  I[#/telephoneNumber] S[#/properties/telephoneNumber/format] \"foobar\" is not valid \"tel\""),
		},
	} {
		t.Run(fmt.Sprintf("case=%d description=%s", k, tc.description), func(t *testing.T) {
			id := &Identity{ID: iid, RecoveryAddresses: tc.existing}
			c := jsonschema.NewCompiler()
			runner, err := schema.NewExtensionRunner(ctx)
			require.NoError(t, err)

			e := NewSchemaExtensionRecovery(id)
			runner.AddRunner(e).Register(c)

			err = c.MustCompile(ctx, tc.schema).Validate(bytes.NewBufferString(tc.doc))
			if tc.expectErr != nil {
				require.EqualError(t, err, tc.expectErr.Error())
				return
			}

			require.NoError(t, e.Finish())

			addresses := id.RecoveryAddresses
			require.Len(t, addresses, len(tc.expect))

			for _, actual := range addresses {
				var found bool
				for _, expect := range tc.expect {
					if reflect.DeepEqual(actual, expect) {
						found = true
						break
					}
				}
				assert.True(t, found, "%+v not in %+v", actual, tc.expect)
			}
		})
	}
}
