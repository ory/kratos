package identity

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/jsonschema/v3"
	_ "github.com/ory/jsonschema/v3/fileloader"
	"github.com/ory/kratos/schema"
	"github.com/ory/kratos/x"
)

const (
	emailVerificationSchemaPath = "file://./stub/extension/verify/email.schema.json"
	phoneVerificationSchemaPath = "file://./stub/extension/verify/phone.schema.json"
)

var ctx = context.Background()

func TestSchemaExtensionVerification(t *testing.T) {
	t.Run("address verification", func(t *testing.T) {
		iid := x.NewUUID()

		for _, tc := range []struct {
			name      string
			schema    string
			expectErr error
			doc       string
			existing  []VerifiableAddress
			expect    []VerifiableAddress
		}{
			{
				name:   "email:must create new address",
				schema: emailVerificationSchemaPath,
				doc:    `{"username":"foo@ory.sh"}`,
				expect: []VerifiableAddress{
					{
						Value:      "foo@ory.sh",
						Verified:   false,
						Status:     VerifiableAddressStatusPending,
						Via:        VerifiableAddressTypeEmail,
						IdentityID: iid,
					},
				},
			},
			{
				name:   "email:must create new address because new and existing doesn't match",
				schema: emailVerificationSchemaPath,
				doc:    `{"username":"foo@ory.sh"}`,
				existing: []VerifiableAddress{
					{
						Value:      "bar@ory.sh",
						Verified:   false,
						Status:     VerifiableAddressStatusPending,
						Via:        VerifiableAddressTypeEmail,
						IdentityID: iid,
					},
				},
				expect: []VerifiableAddress{
					{
						Value:      "foo@ory.sh",
						Verified:   false,
						Status:     VerifiableAddressStatusPending,
						Via:        VerifiableAddressTypeEmail,
						IdentityID: iid,
					},
				},
			},
			{
				name:   "email:must find existing address in case of match",
				schema: emailVerificationSchemaPath,
				doc:    `{"emails":["baz@ory.sh","foo@ory.sh"]}`,
				existing: []VerifiableAddress{
					{
						Value:      "foo@ory.sh",
						Verified:   true,
						Status:     VerifiableAddressStatusCompleted,
						Via:        VerifiableAddressTypeEmail,
						IdentityID: iid,
					},
					{
						Value:      "bar@ory.sh",
						Verified:   true,
						Status:     VerifiableAddressStatusCompleted,
						Via:        VerifiableAddressTypeEmail,
						IdentityID: iid,
					},
				},
				expect: []VerifiableAddress{
					{
						Value:      "foo@ory.sh",
						Verified:   true,
						Status:     VerifiableAddressStatusCompleted,
						Via:        VerifiableAddressTypeEmail,
						IdentityID: iid,
					},
					{
						Value:      "baz@ory.sh",
						Verified:   false,
						Status:     VerifiableAddressStatusPending,
						Via:        VerifiableAddressTypeEmail,
						IdentityID: iid,
					},
				},
			},
			{
				name:   "email:must return only one address in case of duplication",
				schema: emailVerificationSchemaPath,
				doc:    `{"emails":["foo@ory.sh","foo@ory.sh","baz@ory.sh"]}`,
				existing: []VerifiableAddress{
					{
						Value:      "foo@ory.sh",
						Verified:   true,
						Status:     VerifiableAddressStatusCompleted,
						Via:        VerifiableAddressTypeEmail,
						IdentityID: iid,
					},
					{
						Value:      "bar@ory.sh",
						Verified:   true,
						Status:     VerifiableAddressStatusCompleted,
						Via:        VerifiableAddressTypeEmail,
						IdentityID: iid,
					},
				},
				expect: []VerifiableAddress{
					{
						Value:      "foo@ory.sh",
						Verified:   true,
						Status:     VerifiableAddressStatusCompleted,
						Via:        VerifiableAddressTypeEmail,
						IdentityID: iid,
					},
					{
						Value:      "baz@ory.sh",
						Verified:   false,
						Status:     VerifiableAddressStatusPending,
						Via:        VerifiableAddressTypeEmail,
						IdentityID: iid,
					},
				},
			},
			{
				name:      "email:must return error for malformed input",
				schema:    emailVerificationSchemaPath,
				doc:       `{"emails":["foo@ory.sh","bar@ory.sh"], "username": "foobar"}`,
				expectErr: errors.New("I[#/username] S[#/properties/username/format] \"foobar\" is not valid \"email\""),
			},
			{
				name:   "email:must merge email addresses from multiple attributes",
				schema: emailVerificationSchemaPath,
				doc:    `{"emails":["foo@ory.sh","bar@ory.sh","bar@ory.sh"], "username": "foobar@ory.sh"}`,
				expect: []VerifiableAddress{
					{
						Value:      "foo@ory.sh",
						Verified:   false,
						Status:     VerifiableAddressStatusPending,
						Via:        VerifiableAddressTypeEmail,
						IdentityID: iid,
					},
					{
						Value:      "bar@ory.sh",
						Verified:   false,
						Status:     VerifiableAddressStatusPending,
						Via:        VerifiableAddressTypeEmail,
						IdentityID: iid,
					},
					{
						Value:      "foobar@ory.sh",
						Verified:   false,
						Status:     VerifiableAddressStatusPending,
						Via:        VerifiableAddressTypeEmail,
						IdentityID: iid,
					},
				},
			},
			{
				name:   "phone:must create new address",
				schema: phoneVerificationSchemaPath,
				doc:    `{"username":"+18004444444"}`,
				expect: []VerifiableAddress{
					{
						Value:      "+18004444444",
						Verified:   false,
						Status:     VerifiableAddressStatusPending,
						Via:        VerifiableAddressTypePhone,
						IdentityID: iid,
					},
				},
			},
			{
				name:   "phone:must create new address because new and existing doesn't match",
				schema: phoneVerificationSchemaPath,
				doc:    `{"username":"+18004444444"}`,
				existing: []VerifiableAddress{
					{
						Value:      "+442087599036",
						Verified:   false,
						Status:     VerifiableAddressStatusPending,
						Via:        VerifiableAddressTypePhone,
						IdentityID: iid,
					},
				},
				expect: []VerifiableAddress{
					{
						Value:      "+18004444444",
						Verified:   false,
						Status:     VerifiableAddressStatusPending,
						Via:        VerifiableAddressTypePhone,
						IdentityID: iid,
					},
				},
			},
			{
				name:   "phone:must find existing addresses in case of match",
				schema: phoneVerificationSchemaPath,
				doc:    `{"phones":["+18004444444","+442087599036"]}`,
				existing: []VerifiableAddress{
					{
						Value:      "+442087599036",
						Verified:   true,
						Status:     VerifiableAddressStatusCompleted,
						Via:        VerifiableAddressTypePhone,
						IdentityID: iid,
					},
					{
						Value:      "+380634872774",
						Verified:   true,
						Status:     VerifiableAddressStatusCompleted,
						Via:        VerifiableAddressTypePhone,
						IdentityID: iid,
					},
				},
				expect: []VerifiableAddress{
					{
						Value:      "+442087599036",
						Verified:   true,
						Status:     VerifiableAddressStatusCompleted,
						Via:        VerifiableAddressTypePhone,
						IdentityID: iid,
					},
					{
						Value:      "+18004444444",
						Verified:   false,
						Status:     VerifiableAddressStatusPending,
						Via:        VerifiableAddressTypePhone,
						IdentityID: iid,
					},
				},
			},
			{
				name:   "phone:must return only one address in case of duplication",
				schema: phoneVerificationSchemaPath,
				doc:    `{"phones": ["+18004444444","+18004444444","+442087599036"]}`,
				existing: []VerifiableAddress{
					{
						Value:      "+18004444444",
						Verified:   false,
						Status:     VerifiableAddressStatusPending,
						Via:        VerifiableAddressTypePhone,
						IdentityID: iid,
					},
					{
						Value:      "+380634872774",
						Verified:   true,
						Status:     VerifiableAddressStatusCompleted,
						Via:        VerifiableAddressTypePhone,
						IdentityID: iid,
					},
				},
				expect: []VerifiableAddress{
					{
						Value:      "+18004444444",
						Verified:   false,
						Status:     VerifiableAddressStatusPending,
						Via:        VerifiableAddressTypePhone,
						IdentityID: iid,
					},
					{
						Value:      "+442087599036",
						Verified:   false,
						Status:     VerifiableAddressStatusPending,
						Via:        VerifiableAddressTypePhone,
						IdentityID: iid,
					},
				},
			},
			{
				name:   "phone:must merge phones from multiple attributes",
				schema: phoneVerificationSchemaPath,
				doc:    `{"phones": ["+18004444444","+18004444444","+442087599036"], "username": "+380634872774"}`,
				expect: []VerifiableAddress{
					{
						Value:      "+18004444444",
						Verified:   false,
						Status:     VerifiableAddressStatusPending,
						Via:        VerifiableAddressTypePhone,
						IdentityID: iid,
					},
					{
						Value:      "+442087599036",
						Verified:   false,
						Status:     VerifiableAddressStatusPending,
						Via:        VerifiableAddressTypePhone,
						IdentityID: iid,
					},
					{
						Value:      "+380634872774",
						Verified:   false,
						Status:     VerifiableAddressStatusPending,
						Via:        VerifiableAddressTypePhone,
						IdentityID: iid,
					},
				},
			},
			{
				name:      "phone:must return error for malformed input",
				schema:    phoneVerificationSchemaPath,
				doc:       `{"phones":["+18004444444","+18004444444","12112112"], "username": "+380634872774"}`,
				expectErr: errors.New("I[#/phones/2] S[#/properties/phones/items/format] \"12112112\" is not valid \"phone\""),
			},
		} {
			t.Run(fmt.Sprintf("case=%v", tc.name), func(t *testing.T) {
				id := &Identity{ID: iid, VerifiableAddresses: tc.existing}

				c := jsonschema.NewCompiler()

				runner, err := schema.NewExtensionRunner(ctx)
				require.NoError(t, err)

				e := NewSchemaExtensionVerification(id, time.Minute)
				runner.AddRunner(e).Register(c)

				err = c.MustCompile(ctx, tc.schema).Validate(bytes.NewBufferString(tc.doc))
				if tc.expectErr != nil {
					require.EqualError(t, err, tc.expectErr.Error())
					return
				}

				require.NoError(t, e.Finish())

				addresses := id.VerifiableAddresses
				require.Len(t, addresses, len(tc.expect))

				mustContainVerifiableAddress(t, tc.expect, addresses)
			})
		}
	})

}

func mustContainVerifiableAddress(t *testing.T, expected, actual []VerifiableAddress) {
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
