package identity

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/jsonschema/v3"
	_ "github.com/ory/jsonschema/v3/fileloader"
	"github.com/ory/kratos/schema"
	"github.com/ory/kratos/x"
)

const (
	emailRecoverySchemaPath = "file://./stub/extension/recovery/email.schema.json"
	phoneRecoverySchemaPath = "file://./stub/extension/recovery/phone.schema.json"
)

func TestSchemaExtensionRecovery(t *testing.T) {
	iid := x.NewUUID()

	for _, tc := range []struct {
		name      string
		schema    string
		expectErr error
		doc       string
		existing  []RecoveryAddress
		expect    []RecoveryAddress
	}{
		{
			name:   "email:must create new address",
			schema: emailRecoverySchemaPath,
			doc:    `{"username":"foo@ory.sh"}`,
			expect: []RecoveryAddress{
				{
					Value:      "foo@ory.sh",
					Via:        RecoveryAddressTypeEmail,
					IdentityID: iid,
				},
			},
		},
		{
			name:   "email:must create new address because new and existing doesn't match",
			schema: emailRecoverySchemaPath,
			doc:    `{"username":"foo@ory.sh"}`,
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
			name:   "email:must find existing address in case of match",
			schema: emailRecoverySchemaPath,
			doc:    `{"emails":["baz@ory.sh","foo@ory.sh"]}`,
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
			name:   "email:must return only one address in case of duplication",
			schema: emailRecoverySchemaPath,
			doc:    `{"emails":["foo@ory.sh","foo@ory.sh","baz@ory.sh"]}`,
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
			name:      "email:must return error for malformed input",
			schema:    emailRecoverySchemaPath,
			doc:       `{"emails":["foo@ory.sh","bar@ory.sh"], "username": "foobar"}`,
			expectErr: errors.New("I[#/username] S[#/properties/username/format] \"foobar\" is not valid \"email\""),
		},
		{
			name:   "email:must merge email addresses from multiple attributes",
			schema: emailRecoverySchemaPath,
			doc:    `{"emails":["foo@ory.sh","bar@ory.sh","bar@ory.sh"], "username": "foobar@ory.sh"}`,
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
			name:   "phone:must create new address",
			schema: phoneRecoverySchemaPath,
			doc:    `{"username":"+18004444444"}`,
			expect: []RecoveryAddress{
				{
					Value:      "+18004444444",
					Via:        RecoveryAddressTypePhone,
					IdentityID: iid,
				},
			},
		},
		{
			name:   "phone:must create new address because new and existing doesn't match",
			schema: phoneRecoverySchemaPath,
			doc:    `{"username":"+18004444444"}`,
			existing: []RecoveryAddress{
				{
					Value:      "+442087599036",
					Via:        RecoveryAddressTypePhone,
					IdentityID: iid,
				},
			},
			expect: []RecoveryAddress{
				{
					Value:      "+18004444444",
					Via:        RecoveryAddressTypePhone,
					IdentityID: iid,
				},
			},
		},
		{
			name:   "phone:must find existing addresses in case of match",
			schema: phoneRecoverySchemaPath,
			doc:    `{"phones":["+18004444444","+442087599036"]}`,
			existing: []RecoveryAddress{
				{
					Value:      "+442087599036",
					Via:        RecoveryAddressTypePhone,
					IdentityID: iid,
				},
			},
			expect: []RecoveryAddress{
				{
					Value:      "+442087599036",
					Via:        RecoveryAddressTypePhone,
					IdentityID: iid,
				},
				{
					Value:      "+18004444444",
					Via:        RecoveryAddressTypePhone,
					IdentityID: iid,
				},
			},
		},
		{
			name:   "phone:must return only one address in case of duplication",
			schema: phoneRecoverySchemaPath,
			doc:    `{"phones": ["+18004444444","+18004444444","+442087599036"]}`,
			existing: []RecoveryAddress{
				{
					Value:      "+18004444444",
					Via:        RecoveryAddressTypePhone,
					IdentityID: iid,
				},
			},
			expect: []RecoveryAddress{
				{
					Value:      "+18004444444",
					Via:        RecoveryAddressTypePhone,
					IdentityID: iid,
				},
				{
					Value:      "+442087599036",
					Via:        RecoveryAddressTypePhone,
					IdentityID: iid,
				},
			},
		},
		{
			name:   "phone:must merge phones from multiple attributes",
			schema: phoneRecoverySchemaPath,
			doc:    `{"phones": ["+18004444444","+18004444444","+442087599036"], "username": "+380634872774"}`,
			expect: []RecoveryAddress{
				{
					Value:      "+18004444444",
					Via:        RecoveryAddressTypePhone,
					IdentityID: iid,
				},
				{
					Value:      "+442087599036",
					Via:        RecoveryAddressTypePhone,
					IdentityID: iid,
				},
				{
					Value:      "+380634872774",
					Via:        RecoveryAddressTypePhone,
					IdentityID: iid,
				},
			},
		},
		{
			name:      "phone:must return error for malformed input",
			schema:    phoneRecoverySchemaPath,
			doc:       `{"phones":["+18004444444","+18004444444","12112112"], "username": "+380634872774"}`,
			expectErr: errors.New("I[#/phones/2] S[#/properties/phones/items/format] \"12112112\" is not valid \"phone\""),
		},
	} {
		t.Run(fmt.Sprintf("case=%s", tc.name), func(t *testing.T) {
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

			mustContainRecoveryAddress(t, tc.expect, addresses)
		})
	}
}

func mustContainRecoveryAddress(t *testing.T, expected, actual []RecoveryAddress) {
	for _, act := range actual {
		var found bool
		for _, expect := range expected {
			if reflect.DeepEqual(act, expect) {
				found = true
				break
			}
		}
		assert.True(t, found, "%+v not in %+v", act, expected)
	}
}
