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

	"github.com/gofrs/uuid"

	"github.com/ory/kratos/schema"
	"github.com/ory/kratos/x"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSchemaExtensionRecovery(t *testing.T) {
	iid := x.NewUUID()
	breakGlassOrgID := uuid.NullUUID{UUID: x.NewUUID(), Valid: true}
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
					Via:        AddressTypeEmail,
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
					Via:        AddressTypeEmail,
					IdentityID: iid,
				},
			},
			existing: []RecoveryAddress{
				{
					Value:      "bar@ory.sh",
					Via:        AddressTypeEmail,
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
					Via:        AddressTypeEmail,
					IdentityID: iid,
				},
				{
					Value:      "baz@ory.sh",
					Via:        AddressTypeEmail,
					IdentityID: iid,
				},
			},
			existing: []RecoveryAddress{
				{
					Value:      "foo@ory.sh",
					Via:        AddressTypeEmail,
					IdentityID: iid,
				},
				{
					Value:      "bar@ory.sh",
					Via:        AddressTypeEmail,
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
					Via:        AddressTypeEmail,
					IdentityID: iid,
				},
				{
					Value:      "baz@ory.sh",
					Via:        AddressTypeEmail,
					IdentityID: iid,
				},
			},
			existing: []RecoveryAddress{
				{
					Value:      "foo@ory.sh",
					Via:        AddressTypeEmail,
					IdentityID: iid,
				},
				{
					Value:      "bar@ory.sh",
					Via:        AddressTypeEmail,
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
					Via:        AddressTypeEmail,
					IdentityID: iid,
				},
				{
					Value:      "bar@ory.sh",
					Via:        AddressTypeEmail,
					IdentityID: iid,
				},
				{
					Value:      "foobar@ory.sh",
					Via:        AddressTypeEmail,
					IdentityID: iid,
				},
			},
		},
		{
			description: "valid phone number normalization",
			doc:         `{"telephoneNumber":"+33 6 12 34 56 78"}`,
			schema:      "file://./stub/extension/recovery/sms.schema.json",
			expect: []RecoveryAddress{
				{
					Value:      "+33612345678",
					Via:        AddressTypeSMS,
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
					Via:        AddressTypeSMS,
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
					Via:        AddressTypeSMS,
					IdentityID: iid,
				},
			},
			existing: []RecoveryAddress{
				{
					Value:      "+12 345 67890123",
					Via:        AddressTypeSMS,
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
					Via:        AddressTypeSMS,
					IdentityID: iid,
				},
			},
			existing: []RecoveryAddress{
				{
					Value:      "+68672098006",
					Via:        AddressTypeSMS,
					IdentityID: iid,
				},
				{
					Value:      "+33 07856952",
					Via:        AddressTypeSMS,
					IdentityID: iid,
				},
			},
		},
		{
			description: "invalid phone number",
			doc:         `{"telephoneNumber": "foobar"}`,
			schema:      "file://./stub/extension/recovery/sms.schema.json",
			// We get 2 errors: one from the JSON schema `format` validation and one from the normalization error.
			expectErr: errors.New("I[#/telephoneNumber] S[#/properties/telephoneNumber] validation failed\n  I[#/telephoneNumber] S[#/properties/telephoneNumber/format] \"foobar\" is not valid \"tel\"\n  I[#/telephoneNumber] S[#/properties/telephoneNumber/format] \"foobar\" is not valid \"tel\" for \"sms\""),
		},
		{
			description: "phone number that fails normalization",
			doc:         `{"telephoneNumber": "+1234"}`, // Too short to be a valid E.164 number
			schema:      "file://./stub/extension/recovery/sms.schema.json",
			// The JSON schema format validator rejects it, and normalization also fails
			expectErr: errors.New("I[#/telephoneNumber] S[#/properties/telephoneNumber] validation failed\n  I[#/telephoneNumber] S[#/properties/telephoneNumber/format] \"+1234\" is not valid \"tel\"\n  I[#/telephoneNumber] S[#/properties/telephoneNumber/format] \"+1234\" is not valid \"tel\" for \"sms\""),
		},
		{
			description: "break_glass_for_organization preserved on existing recovery address",
			doc:         `{"username":"foo@ory.sh"}`,
			schema:      "file://./stub/extension/recovery/email.schema.json",
			expect: []RecoveryAddress{
				{
					Value:                     "foo@ory.sh",
					Via:                       AddressTypeEmail,
					IdentityID:                iid,
					BreakGlassForOrganization: breakGlassOrgID,
				},
			},
			existing: []RecoveryAddress{
				{
					Value:                     "foo@ory.sh",
					Via:                       AddressTypeEmail,
					IdentityID:                iid,
					BreakGlassForOrganization: breakGlassOrgID,
				},
			},
		},
		{
			description: "break_glass null preserved on existing recovery address",
			doc:         `{"username":"foo@ory.sh"}`,
			schema:      "file://./stub/extension/recovery/email.schema.json",
			expect: []RecoveryAddress{
				{
					Value:      "foo@ory.sh",
					Via:        AddressTypeEmail,
					IdentityID: iid,
				},
			},
			existing: []RecoveryAddress{
				{
					Value:      "foo@ory.sh",
					Via:        AddressTypeEmail,
					IdentityID: iid,
				},
			},
		},
	} {
		t.Run(fmt.Sprintf("case=%d description=%s", k, tc.description), func(t *testing.T) {
			id := &Identity{ID: iid, RecoveryAddresses: tc.existing}
			c := jsonschema.NewCompiler()
			runner, err := schema.NewExtensionRunner(t.Context())
			require.NoError(t, err)

			e := NewSchemaExtensionRecovery(id)
			runner.AddRunner(e).Register(c)

			err = c.MustCompile(t.Context(), tc.schema).Validate(bytes.NewBufferString(tc.doc))
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

// TestSchemaExtensionRecovery_PreservesExistingRecordWhenValueIsNotNormalized
// is a regression test for the same root cause as #11283's verification bug,
// applied to recovery addresses: identities whose persisted
// RecoveryAddress.Value is in pre-normalization raw form get their existing
// record (ID, BreakGlassForOrganization, timestamps) dropped on every
// re-validation because the extension now emits the E.164-normalized form
// and has() compares values byte-wise.
func TestSchemaExtensionRecovery_PreservesExistingRecordWhenValueIsNotNormalized(t *testing.T) {
	iid := x.NewUUID()
	existingID := x.NewUUID()
	breakGlassOrgID := uuid.NullUUID{UUID: x.NewUUID(), Valid: true}

	rawPhone := "+49 0176 67111638"
	id := &Identity{
		ID: iid,
		RecoveryAddresses: []RecoveryAddress{{
			ID:                        existingID,
			Value:                     rawPhone,
			Via:                       AddressTypeSMS,
			IdentityID:                iid,
			BreakGlassForOrganization: breakGlassOrgID,
		}},
	}

	c := jsonschema.NewCompiler()
	runner, err := schema.NewExtensionRunner(t.Context())
	require.NoError(t, err)

	e := NewSchemaExtensionRecovery(id)
	runner.AddRunner(e).Register(c)

	err = c.MustCompile(t.Context(), "file://./stub/extension/recovery/sms.schema.json").
		Validate(bytes.NewBufferString(fmt.Sprintf(`{"telephoneNumber":%q}`, rawPhone)))
	require.NoError(t, err)
	require.NoError(t, e.Finish())

	require.Len(t, id.RecoveryAddresses, 1)
	assert.Equal(t, existingID, id.RecoveryAddresses[0].ID,
		"existing recovery record ID must be retained on re-validation with raw phone")
	assert.Equal(t, breakGlassOrgID, id.RecoveryAddresses[0].BreakGlassForOrganization,
		"break_glass_for_organization must be retained on re-validation with raw phone")
}
