// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package identity

import (
	"bytes"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/ory/x/snapshotx"

	"github.com/ory/kratos/x"

	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	"github.com/gofrs/uuid"

	"github.com/ory/x/sqlxx"

	"github.com/ory/kratos/driver/config"

	"github.com/stretchr/testify/assert"
)

func TestNewIdentity(t *testing.T) {
	i := NewIdentity(config.DefaultIdentityTraitsSchemaID)
	assert.Equal(t, uuid.Nil, i.ID)
	assert.NotEmpty(t, i.Traits)
	assert.NotNil(t, i.Credentials)
	assert.True(t, i.IsActive())
}

func TestIdentityCredentialsOr(t *testing.T) {
	i := NewIdentity(config.DefaultIdentityTraitsSchemaID)
	i.Credentials = nil

	expected := &Credentials{ID: x.NewUUID(), Type: CredentialsTypePassword}
	assert.Equal(t, expected, i.GetCredentialsOr(CredentialsTypePassword, expected))

	expected = &Credentials{ID: x.NewUUID(), Type: CredentialsTypeWebAuthn}
	i.SetCredentials(CredentialsTypeWebAuthn, *expected)

	assert.Equal(t, expected, i.GetCredentialsOr(CredentialsTypeWebAuthn, nil))
}

func TestIdentityCredentialsOrCreate(t *testing.T) {
	i := NewIdentity(config.DefaultIdentityTraitsSchemaID)
	i.Credentials = nil

	expected := &Credentials{Config: []byte("true"), IdentityID: i.ID, Type: CredentialsTypePassword}
	i.UpsertCredentialsConfig(CredentialsTypePassword, []byte("true"), 0)
	actual, ok := i.GetCredentials(CredentialsTypePassword)
	assert.True(t, ok)
	assert.Equal(t, expected, actual)
}

func TestIdentityCredentials(t *testing.T) {
	i := NewIdentity(config.DefaultIdentityTraitsSchemaID)
	i.Credentials = nil

	// Shouldn't error if map is nil
	i.DeleteCredentialsType(CredentialsTypeTOTP)

	expectedTOTP := Credentials{ID: x.NewUUID(), Type: CredentialsTypeTOTP}
	i.SetCredentials(CredentialsTypeTOTP, expectedTOTP)
	actual, found := i.GetCredentials(CredentialsTypeTOTP)
	assert.True(t, found, "should set and find the credential if map was nil")
	assert.Equal(t, &expectedTOTP, actual)

	expectedPassword := Credentials{ID: x.NewUUID(), Type: CredentialsTypePassword}
	i.SetCredentials(CredentialsTypePassword, expectedPassword)
	actual, found = i.GetCredentials(CredentialsTypePassword)
	assert.True(t, found, "should set and find the credential if map was not nil")
	assert.Equal(t, &expectedPassword, actual)

	expectedOIDC := Credentials{ID: x.NewUUID()}
	i.SetCredentials(CredentialsTypeOIDC, expectedOIDC)
	actual, found = i.GetCredentials(CredentialsTypeOIDC)
	assert.True(t, found)
	assert.Equal(t, expectedOIDC.ID, actual.ID)
	assert.Equal(t, CredentialsTypeOIDC, actual.Type, "should set the type if we forgot to set it in the credentials struct")

	i.DeleteCredentialsType(CredentialsTypePassword)
	_, found = i.GetCredentials(CredentialsTypePassword)
	assert.False(t, found, "should delete a credential properly")

	actual, found = i.GetCredentials(CredentialsTypeTOTP)
	assert.True(t, found, "but not alter other credentials")
	assert.Equal(t, &expectedTOTP, actual)
}

func TestMarshalExcludesCredentials(t *testing.T) {
	i := NewIdentity(config.DefaultIdentityTraitsSchemaID)
	i.Credentials = map[CredentialsType]Credentials{
		CredentialsTypePassword: {
			ID: uuid.UUID{},
		},
	}

	var b bytes.Buffer
	require.Nil(t, json.NewEncoder(&b).Encode(i))

	assert.False(t, gjson.Get(b.String(), "credentials").Exists(), "Credentials should not be rendered to json")

	// To ensure the original identity is not changed / Unmarshal has no side effects:
	require.NotEmpty(t, i.Credentials)
}

func TestMarshalExcludesCredentialsByReference(t *testing.T) {
	i := NewIdentity(config.DefaultIdentityTraitsSchemaID)
	i.Credentials = map[CredentialsType]Credentials{
		CredentialsTypePassword: {
			ID: uuid.UUID{},
		},
	}

	var b bytes.Buffer
	require.Nil(t, json.NewEncoder(&b).Encode(&i))

	assert.False(t, gjson.Get(b.String(), "credentials").Exists(), "Credentials should not be rendered to json")

	// To ensure the original identity is not changed / Unmarshal has no side effects:
	require.NotEmpty(t, i.Credentials)
}

func TestMarshalIgnoresAdminMetadata(t *testing.T) {
	i := NewIdentity(config.DefaultIdentityTraitsSchemaID)
	i.MetadataAdmin = []byte(`{"admin":"bar"}`)
	i.MetadataPublic = []byte(`{"public":"bar"}`)

	var b bytes.Buffer
	require.Nil(t, json.NewEncoder(&b).Encode(&i))

	assert.False(t, gjson.Get(b.String(), "metadata_admin.admin").Exists(), "Admin metadata should not be rendered to json but got: %s", b.String())
	assert.Equal(t, "bar", gjson.Get(b.String(), "metadata_public.public").String(), "Public metadata should be rendered to json")

	// To ensure the original identity is not changed / Unmarshal has no side effects:
	require.NotEmpty(t, i.MetadataAdmin)
	require.NotEmpty(t, i.MetadataPublic)
}

func TestUnMarshallIgnoresCredentials(t *testing.T) {
	jsonText := "{\"id\":\"3234ad11-49c6-49e2-bfac-537f3e06cd85\",\"schema_id\":\"default\",\"schema_url\":\"\",\"traits\":{}, \"credentials\" : {\"password\":{\"type\":\"\",\"identifiers\":null,\"config\":null,\"updatedAt\":\"0001-01-01T00:00:00Z\"}}}"
	var i Identity
	err := json.Unmarshal([]byte(jsonText), &i)
	assert.Nil(t, err)

	assert.Nil(t, i.Credentials)
	assert.Equal(t, "3234ad11-49c6-49e2-bfac-537f3e06cd85", i.ID.String())
}

func TestUnMarshallIgnoresAdminMetadata(t *testing.T) {
	jsonText := "{\"id\":\"3234ad11-49c6-49e2-bfac-537f3e06cd85\",\"schema_id\":\"default\",\"schema_url\":\"\",\"traits\":{}, \"admin_metadata\" : {\"foo\":\"bar\"}}"
	var i Identity
	err := json.Unmarshal([]byte(jsonText), &i)
	assert.Nil(t, err)

	assert.Nil(t, i.MetadataAdmin)
}

func TestMarshalIdentityWithCredentialsWhenCredentialsNil(t *testing.T) {
	i := NewIdentity(config.DefaultIdentityTraitsSchemaID)
	i.Credentials = nil

	var b bytes.Buffer
	require.Nil(t, json.NewEncoder(&b).Encode(WithCredentialsMetadataAndAdminMetadataInJSON(*i)))

	assert.False(t, gjson.Get(b.String(), "credentials").Exists())
}

func TestMarshalIdentityWithAdminMetadata(t *testing.T) {
	i := NewIdentity(config.DefaultIdentityTraitsSchemaID)
	i.MetadataAdmin = []byte(`{"some":"metadata"}`)

	var b bytes.Buffer
	require.Nil(t, json.NewEncoder(&b).Encode(WithAdminMetadataInJSON(*i)))
	assert.Equal(t, "metadata", gjson.GetBytes(i.MetadataAdmin, "some").String(), "Original metadata_admin should not be touched by marshalling")
}

func TestMarshalIdentityWithCredentialsMetadata(t *testing.T) {
	i := NewIdentity(config.DefaultIdentityTraitsSchemaID)
	credentials := map[CredentialsType]Credentials{
		CredentialsTypePassword: {
			Type:   CredentialsTypePassword,
			Config: sqlxx.JSONRawMessage("{\"some\" : \"secret\"}"),
		},
	}
	i.Credentials = credentials
	i.MetadataAdmin = []byte(`{"some":"metadata"}`)

	var b bytes.Buffer
	require.Nil(t, json.NewEncoder(&b).Encode(WithCredentialsMetadataAndAdminMetadataInJSON(*i)))

	credentialsInJson := gjson.Get(b.String(), "credentials")
	assert.True(t, credentialsInJson.Exists())

	assert.JSONEq(t, "{\"password\":{\"type\":\"password\",\"identifiers\":null,\"updated_at\":\"0001-01-01T00:00:00Z\",\"created_at\":\"0001-01-01T00:00:00Z\",\"version\":0}}", credentialsInJson.Raw)
	assert.Equal(t, credentials, i.Credentials, "Original credentials should not be touched by marshalling")
	assert.Equal(t, "metadata", gjson.GetBytes(i.MetadataAdmin, "some").String(), "Original metadata_admin should not be touched by marshalling")
}

func TestMarshalIdentityWithAll(t *testing.T) {
	i := NewIdentity(config.DefaultIdentityTraitsSchemaID)
	credentials := map[CredentialsType]Credentials{
		CredentialsTypePassword: {
			Type:   CredentialsTypePassword,
			Config: sqlxx.JSONRawMessage("{\"some\" : \"secret\"}"),
		},
	}
	i.Credentials = credentials
	i.MetadataAdmin = []byte(`{"some":"metadata"}`)

	var b bytes.Buffer
	require.Nil(t, json.NewEncoder(&b).Encode(WithCredentialsAndAdminMetadataInJSON(*i)))

	credentialsInJson := gjson.Get(b.String(), "credentials")
	assert.True(t, credentialsInJson.Exists())

	snapshotx.SnapshotTExcept(t, json.RawMessage(credentialsInJson.Raw), nil)
	assert.Equal(t, credentials, i.Credentials, "Original credentials should not be touched by marshalling")
	assert.Equal(t, "metadata", gjson.GetBytes(i.MetadataAdmin, "some").String(), "Original credentials should not be touched by marshalling")
}

func TestValidateNID(t *testing.T) {
	nid := x.NewUUID()
	for k, tc := range []struct {
		i           *Identity
		expect      *Identity
		expectedErr bool
	}{
		{i: &Identity{}, expectedErr: true},
		{i: &Identity{NID: nid}},
		{
			i:      &Identity{NID: nid, RecoveryAddresses: []RecoveryAddress{{NID: x.NewUUID()}}},
			expect: &Identity{NID: nid, VerifiableAddresses: []VerifiableAddress{}, RecoveryAddresses: []RecoveryAddress{}},
		},
		{
			i:      &Identity{NID: nid, VerifiableAddresses: []VerifiableAddress{{NID: x.NewUUID()}}},
			expect: &Identity{NID: nid, VerifiableAddresses: []VerifiableAddress{}, RecoveryAddresses: []RecoveryAddress{}},
		},
		{
			i:      &Identity{NID: nid, Credentials: map[CredentialsType]Credentials{CredentialsTypePassword: {NID: x.NewUUID()}}},
			expect: &Identity{NID: nid, Credentials: map[CredentialsType]Credentials{}, VerifiableAddresses: []VerifiableAddress{}, RecoveryAddresses: []RecoveryAddress{}},
		},
		{
			i:      &Identity{NID: nid, VerifiableAddresses: []VerifiableAddress{{NID: x.NewUUID()}}, RecoveryAddresses: []RecoveryAddress{{NID: nid}}},
			expect: &Identity{NID: nid, VerifiableAddresses: []VerifiableAddress{}, RecoveryAddresses: []RecoveryAddress{{NID: nid}}},
		},
		{
			i:      &Identity{NID: nid, VerifiableAddresses: []VerifiableAddress{{NID: nid}}, RecoveryAddresses: []RecoveryAddress{{NID: x.NewUUID()}}},
			expect: &Identity{NID: nid, VerifiableAddresses: []VerifiableAddress{{NID: nid}}, RecoveryAddresses: []RecoveryAddress{}},
		},
		{
			i:      &Identity{NID: nid, VerifiableAddresses: []VerifiableAddress{{NID: nid}}, RecoveryAddresses: []RecoveryAddress{{NID: nid}}},
			expect: &Identity{NID: nid, VerifiableAddresses: []VerifiableAddress{{NID: nid}}, RecoveryAddresses: []RecoveryAddress{{NID: nid}}},
		},
		{
			i:      &Identity{NID: nid, Credentials: map[CredentialsType]Credentials{CredentialsTypePassword: {NID: x.NewUUID()}}, RecoveryAddresses: []RecoveryAddress{{NID: nid}}, VerifiableAddresses: []VerifiableAddress{{NID: nid}}},
			expect: &Identity{NID: nid, Credentials: map[CredentialsType]Credentials{}, RecoveryAddresses: []RecoveryAddress{{NID: nid}}, VerifiableAddresses: []VerifiableAddress{{NID: nid}}},
		},
		{
			i:      &Identity{NID: nid, Credentials: map[CredentialsType]Credentials{CredentialsTypePassword: {NID: nid}}, RecoveryAddresses: []RecoveryAddress{{NID: nid}}, VerifiableAddresses: []VerifiableAddress{{NID: nid}}},
			expect: &Identity{NID: nid, Credentials: map[CredentialsType]Credentials{CredentialsTypePassword: {NID: nid}}, RecoveryAddresses: []RecoveryAddress{{NID: nid}}, VerifiableAddresses: []VerifiableAddress{{NID: nid}}},
		},
	} {
		t.Run(fmt.Sprintf("case=%d", k), func(t *testing.T) {
			err := tc.i.Validate()
			if tc.expectedErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				if tc.expect != nil {
					assert.EqualValues(t, tc.expect, tc.i)
				}
			}
		})
	}
}

// TestRecoveryAddresses tests the CollectRecoveryAddresses are collected from all identities.
func TestRecoveryAddresses(t *testing.T) {
	var addresses []RecoveryAddress

	for i := 0; i < 10; i++ {
		addresses = append(addresses, RecoveryAddress{
			Value: fmt.Sprintf("address-%d", i),
		})
	}

	id1 := &Identity{RecoveryAddresses: addresses[:5]}
	id2 := &Identity{}
	id3 := &Identity{RecoveryAddresses: addresses[5:]}

	assert.Equal(t, addresses, CollectRecoveryAddresses([]*Identity{id1, id2, id3}))
}

// TestVerifiableAddresses tests the VerfifableAddresses are collected from all identities.
func TestVerifiableAddresses(t *testing.T) {
	var addresses []VerifiableAddress

	for i := 0; i < 10; i++ {
		addresses = append(addresses, VerifiableAddress{
			Value: fmt.Sprintf("address-%d", i),
		})
	}

	id1 := &Identity{VerifiableAddresses: addresses[:5]}
	id2 := &Identity{}
	id3 := &Identity{VerifiableAddresses: addresses[5:]}

	assert.Equal(t, addresses, CollectVerifiableAddresses([]*Identity{id1, id2, id3}))
}

func TestWithDeclassifiedCredentials(t *testing.T) {
	i := NewIdentity(config.DefaultIdentityTraitsSchemaID)
	credentials := map[CredentialsType]Credentials{
		CredentialsTypePassword: {
			Identifiers: []string{"zab", "bar"},
			Type:        CredentialsTypePassword,
			Config:      sqlxx.JSONRawMessage("{\"some\" : \"secret\"}"),
		},
		CredentialsTypeOIDC: {
			Type:        CredentialsTypeOIDC,
			Identifiers: []string{"bar", "baz"},
			Config:      sqlxx.JSONRawMessage("{\"some\" : \"secret\"}"),
		},
		CredentialsTypeWebAuthn: {
			Type:        CredentialsTypeWebAuthn,
			Identifiers: []string{"foo", "bar"},
			Config:      sqlxx.JSONRawMessage("{\"some\" : \"secret\"}"),
		},
	}
	i.Credentials = credentials

	t.Run("case=no-include", func(t *testing.T) {
		actualIdentity, err := i.WithDeclassifiedCredentials(ctx, nil, nil)
		require.NoError(t, err)

		for ct, actual := range actualIdentity.Credentials {
			t.Run("credential="+string(ct), func(t *testing.T) {
				snapshotx.SnapshotT(t, actual)
			})
		}
	})

	t.Run("case=include-webauthn", func(t *testing.T) {
		actualIdentity, err := i.WithDeclassifiedCredentials(ctx, nil, []CredentialsType{CredentialsTypeWebAuthn})
		require.NoError(t, err)

		for ct, actual := range actualIdentity.Credentials {
			t.Run("credential="+string(ct), func(t *testing.T) {
				snapshotx.SnapshotT(t, actual)
			})
		}
	})

	t.Run("case=include-multi", func(t *testing.T) {
		actualIdentity, err := i.WithDeclassifiedCredentials(ctx, nil, []CredentialsType{CredentialsTypeWebAuthn, CredentialsTypePassword})
		require.NoError(t, err)

		for ct, actual := range actualIdentity.Credentials {
			t.Run("credential="+string(ct), func(t *testing.T) {
				snapshotx.SnapshotT(t, actual)
			})
		}
	})
}
