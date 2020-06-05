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
		expectErr error
		schema    string
		doc       string
		expect    []RecoveryAddress
		existing  []RecoveryAddress
	}{
		{
			doc:    `{"username":"foo@ory.sh"}`,
			schema: "file://./stub/extension/recovery/schema.json",
			expect: []RecoveryAddress{
				{
					Value:      "foo@ory.sh",
					Via:        RecoveryAddressTypeEmail,
					IdentityID: iid,
				},
			},
		},
		{
			doc:    `{"username":"foo@ory.sh"}`,
			schema: "file://./stub/extension/recovery/schema.json",
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
			doc:    `{"emails":["baz@ory.sh","foo@ory.sh"]}`,
			schema: "file://./stub/extension/recovery/schema.json",
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
			doc:    `{"emails":["foo@ory.sh","foo@ory.sh","baz@ory.sh"]}`,
			schema: "file://./stub/extension/recovery/schema.json",
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
			doc:       `{"emails":["foo@ory.sh","bar@ory.sh"], "username": "foobar"}`,
			schema:    "file://./stub/extension/recovery/schema.json",
			expectErr: errors.New("I[#/username] S[#/properties/username/format] \"foobar\" is not valid \"email\""),
		},
		{
			doc:    `{"emails":["foo@ory.sh","bar@ory.sh","bar@ory.sh"], "username": "foobar@ory.sh"}`,
			schema: "file://./stub/extension/recovery/schema.json",
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
	} {
		t.Run(fmt.Sprintf("case=%d", k), func(t *testing.T) {
			id := &Identity{ID: iid, RecoveryAddresses: tc.existing}
			c := jsonschema.NewCompiler()
			runner, err := schema.NewExtensionRunner(schema.ExtensionRunnerIdentityMetaSchema)
			require.NoError(t, err)

			e := NewSchemaExtensionRecovery(id)
			runner.AddRunner(e).Register(c)

			err = c.MustCompile(tc.schema).Validate(bytes.NewBufferString(tc.doc))
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
