// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package identity

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"reflect"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/jsonschema/v3"
	_ "github.com/ory/jsonschema/v3/fileloader"
	"github.com/ory/kratos/schema"
	"github.com/ory/kratos/x"
)

const (
	emailSchemaPath                    = "file://./stub/extension/verify/email.schema.json"
	phoneSchemaPath                    = "file://./stub/extension/verify/phone.schema.json"
	missingFormatSchemaPath            = "file://./stub/extension/verify/missing-format.schema.json"
	legacyEmailMissingFormatSchemaPath = "file://./stub/extension/verify/legacy-email-missing-format.schema.json"
	noValidateSchemaPath               = "file://./stub/extension/verify/no-validate.schema.json"
)

var ctx = context.Background()

func TestSchemaExtensionVerification(t *testing.T) {
	net.IP{}.IsPrivate()
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
				schema: emailSchemaPath,
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
				schema: emailSchemaPath,
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
				schema: emailSchemaPath,
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
				schema: emailSchemaPath,
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
				schema:    emailSchemaPath,
				doc:       `{"emails":["foo@ory.sh","bar@ory.sh"], "username": "foobar"}`,
				expectErr: errors.New("I[#/username] S[#/properties/username/format] \"foobar\" is not valid \"email\""),
			},
			{
				name:   "email:must merge email addresses from multiple attributes",
				schema: emailSchemaPath,
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
				schema: phoneSchemaPath,
				doc:    `{"username":"+18004444444"}`,
				expect: []VerifiableAddress{
					{
						Value:      "+18004444444",
						Verified:   false,
						Status:     VerifiableAddressStatusPending,
						Via:        ChannelTypeSMS,
						IdentityID: iid,
					},
				},
			},
			{
				name:   "phone:must create new address because new and existing doesn't match",
				schema: phoneSchemaPath,
				doc:    `{"username":"+18004444444"}`,
				existing: []VerifiableAddress{
					{
						Value:      "+442087599036",
						Verified:   false,
						Status:     VerifiableAddressStatusPending,
						Via:        ChannelTypeSMS,
						IdentityID: iid,
					},
				},
				expect: []VerifiableAddress{
					{
						Value:      "+18004444444",
						Verified:   false,
						Status:     VerifiableAddressStatusPending,
						Via:        ChannelTypeSMS,
						IdentityID: iid,
					},
				},
			},
			{
				name:   "phone:must find existing addresses in case of match",
				schema: phoneSchemaPath,
				doc:    `{"phones":["+18004444444","+442087599036"]}`,
				existing: []VerifiableAddress{
					{
						Value:      "+442087599036",
						Verified:   true,
						Status:     VerifiableAddressStatusCompleted,
						Via:        ChannelTypeSMS,
						IdentityID: iid,
					},
					{
						Value:      "+380634872774",
						Verified:   true,
						Status:     VerifiableAddressStatusCompleted,
						Via:        ChannelTypeSMS,
						IdentityID: iid,
					},
				},
				expect: []VerifiableAddress{
					{
						Value:      "+442087599036",
						Verified:   true,
						Status:     VerifiableAddressStatusCompleted,
						Via:        ChannelTypeSMS,
						IdentityID: iid,
					},
					{
						Value:      "+18004444444",
						Verified:   false,
						Status:     VerifiableAddressStatusPending,
						Via:        ChannelTypeSMS,
						IdentityID: iid,
					},
				},
			},
			{
				name:   "phone:must return only one address in case of duplication",
				schema: phoneSchemaPath,
				doc:    `{"phones": ["+18004444444","+18004444444","+442087599036"]}`,
				existing: []VerifiableAddress{
					{
						Value:      "+18004444444",
						Verified:   false,
						Status:     VerifiableAddressStatusPending,
						Via:        ChannelTypeSMS,
						IdentityID: iid,
					},
					{
						Value:      "+380634872774",
						Verified:   true,
						Status:     VerifiableAddressStatusCompleted,
						Via:        ChannelTypeSMS,
						IdentityID: iid,
					},
				},
				expect: []VerifiableAddress{
					{
						Value:      "+18004444444",
						Verified:   false,
						Status:     VerifiableAddressStatusPending,
						Via:        ChannelTypeSMS,
						IdentityID: iid,
					},
					{
						Value:      "+442087599036",
						Verified:   false,
						Status:     VerifiableAddressStatusPending,
						Via:        ChannelTypeSMS,
						IdentityID: iid,
					},
				},
			},
			{
				name:   "phone:must merge phones from multiple attributes",
				schema: phoneSchemaPath,
				doc:    `{"phones": ["+18004444444","+18004444444","+442087599036"], "username": "+380634872774"}`,
				expect: []VerifiableAddress{
					{
						Value:      "+18004444444",
						Verified:   false,
						Status:     VerifiableAddressStatusPending,
						Via:        ChannelTypeSMS,
						IdentityID: iid,
					},
					{
						Value:      "+442087599036",
						Verified:   false,
						Status:     VerifiableAddressStatusPending,
						Via:        ChannelTypeSMS,
						IdentityID: iid,
					},
					{
						Value:      "+380634872774",
						Verified:   false,
						Status:     VerifiableAddressStatusPending,
						Via:        ChannelTypeSMS,
						IdentityID: iid,
					},
				},
			},
			{
				name:      "phone:must return error for malformed input",
				schema:    phoneSchemaPath,
				doc:       `{"phones":["+18004444444","+18004444444","12112112"], "username": "+380634872774"}`,
				expectErr: errors.New("I[#/phones/2] S[#/properties/phones/items/format] \"12112112\" is not valid \"tel\""),
			},
			{
				name:      "missing format returns an error",
				schema:    missingFormatSchemaPath,
				doc:       `{"phone": "+380634872774"}`,
				expectErr: errors.New("I[#/phone] S[#/properties/phone/format] no format specified. A format is required if verification is enabled. If this was intentional, please set \"format\" to \"no-validate\""),
			},
			{
				name:   "missing format works for email if format is missing",
				schema: legacyEmailMissingFormatSchemaPath,
				doc:    `{"email": "user@ory.sh"}`,
				expect: []VerifiableAddress{
					{
						Value:      "user@ory.sh",
						Verified:   false,
						Status:     VerifiableAddressStatusPending,
						Via:        ChannelTypeEmail,
						IdentityID: iid,
					},
				},
			},
			{
				name:   "format: no-validate works",
				schema: noValidateSchemaPath,
				doc:    `{"phone": "not a phone number"}`,
				expect: []VerifiableAddress{
					{
						Value:      "not a phone number",
						Verified:   false,
						Status:     VerifiableAddressStatusPending,
						Via:        ChannelTypeSMS,
						IdentityID: iid,
					},
				},
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

				mustContainAddress(t, tc.expect, addresses)
			})
		}
	})
}

func mustContainAddress(t *testing.T, expected, actual []VerifiableAddress) {
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

func TestMergeVerifiableAddresses(t *testing.T) {
	for _, tt := range []struct {
		name                      string
		base, overrides, expected []VerifiableAddress
	}{
		{
			name: "empty base",
			base: []VerifiableAddress{},
			overrides: []VerifiableAddress{{
				Value: "override@ory.sh",
				Via:   "email",
			}},
			expected: []VerifiableAddress{},
		}, {
			name: "no overlap",
			base: []VerifiableAddress{{
				Value: "base@ory.sh",
				Via:   "email",
			}},
			overrides: []VerifiableAddress{{
				Value: "override@ory.sh",
				Via:   "email",
			}},
			expected: []VerifiableAddress{{
				Value: "base@ory.sh",
				Via:   "email",
			}},
		}, {
			name: "overrides",
			base: []VerifiableAddress{
				{Value: "base@ory.sh", Via: "email"},
				{Value: "common-1-is-overwritten@ory.sh", Via: "email"},
				{Value: "common-2-is-ignored@ory.sh", Via: "no match"},
			},
			overrides: []VerifiableAddress{
				{Value: "common-1-is-overwritten@ory.sh", Via: "email", Verified: true, Status: VerifiableAddressStatusCompleted},
				{Value: "common-2-is-ignored@ory.sh", Via: "wrong via"},
				{Value: "override-only-is-ignored@ory.sh", Via: "email"},
			},
			expected: []VerifiableAddress{
				{Value: "base@ory.sh", Via: "email"},
				{Value: "common-1-is-overwritten@ory.sh", Via: "email", Verified: true, Status: VerifiableAddressStatusCompleted},
				{Value: "common-2-is-ignored@ory.sh", Via: "no match"},
			},
		},
	} {
		t.Run("case="+tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, merge(tt.base, tt.overrides))
		})
	}
}
