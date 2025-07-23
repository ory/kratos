// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package identity

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	"github.com/ory/kratos/cipher"
	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/x"
	"github.com/ory/x/snapshotx"
	"github.com/ory/x/sqlxx"
)

func TestNewIdentity(t *testing.T) {
	t.Parallel()

	i := NewIdentity(config.DefaultIdentityTraitsSchemaID)
	assert.Equal(t, uuid.Nil, i.ID)
	assert.NotEmpty(t, i.Traits)
	assert.NotNil(t, i.Credentials)
	assert.True(t, i.IsActive())
}

func TestIdentityCredentialsOr(t *testing.T) {
	t.Parallel()

	i := NewIdentity(config.DefaultIdentityTraitsSchemaID)
	i.Credentials = nil

	expected := &Credentials{ID: x.NewUUID(), Type: CredentialsTypePassword}
	assert.Equal(t, expected, i.GetCredentialsOr(CredentialsTypePassword, expected))

	expected = &Credentials{ID: x.NewUUID(), Type: CredentialsTypeWebAuthn}
	i.SetCredentials(CredentialsTypeWebAuthn, *expected)

	assert.Equal(t, expected, i.GetCredentialsOr(CredentialsTypeWebAuthn, nil))
}

func TestIdentityCredentialsOrCreate(t *testing.T) {
	t.Parallel()

	i := NewIdentity(config.DefaultIdentityTraitsSchemaID)
	i.Credentials = nil

	expected := &Credentials{Config: []byte("true"), IdentityID: i.ID, Type: CredentialsTypePassword}
	i.UpsertCredentialsConfig(CredentialsTypePassword, []byte("true"), 0)
	actual, ok := i.GetCredentials(CredentialsTypePassword)
	assert.True(t, ok)
	assert.Equal(t, expected, actual)
}

func TestIdentityCredentials(t *testing.T) {
	t.Parallel()

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
	t.Parallel()

	i := NewIdentity(config.DefaultIdentityTraitsSchemaID)
	i.Credentials = map[CredentialsType]Credentials{
		CredentialsTypePassword: {
			ID: uuid.UUID{},
		},
	}

	rawJSON, err := json.Marshal(i)
	require.NoError(t, err)

	creds := gjson.GetBytes(rawJSON, "credentials")
	assert.Falsef(t, creds.Exists(), "Credentials should not be rendered to JSON, but got: %q", creds.Raw)

	// To ensure the original identity is not changed / Unmarshal has no side effects:
	require.NotEmpty(t, i.Credentials)
}

func TestMarshalExcludesCredentialsByReference(t *testing.T) {
	t.Parallel()

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
	t.Parallel()

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
	t.Parallel()

	jsonText := `{"id":"3234ad11-49c6-49e2-bfac-537f3e06cd85","schema_id":"default","schema_url":"","traits":{}, "credentials" : {"password":{"type":"","identifiers":null,"config":null,"updatedAt":"0001-01-01T00:00:00Z"}}}`
	var i Identity
	err := json.Unmarshal([]byte(jsonText), &i)
	assert.Nil(t, err)

	assert.Nil(t, i.Credentials)
	assert.Equal(t, "3234ad11-49c6-49e2-bfac-537f3e06cd85", i.ID.String())
}

func TestUnMarshallIgnoresAdminMetadata(t *testing.T) {
	t.Parallel()

	jsonText := `{"id":"3234ad11-49c6-49e2-bfac-537f3e06cd85","schema_id":"default","schema_url":"","traits":{}, "admin_metadata" : {"foo":"bar"}}`
	var i Identity
	err := json.Unmarshal([]byte(jsonText), &i)
	assert.Nil(t, err)

	assert.Nil(t, i.MetadataAdmin)
}

func TestMarshalIdentityWithCredentialsWhenCredentialsNil(t *testing.T) {
	t.Parallel()

	i := NewIdentity(config.DefaultIdentityTraitsSchemaID)
	i.Credentials = nil

	var b bytes.Buffer
	require.Nil(t, json.NewEncoder(&b).Encode(WithCredentialsNoConfigAndAdminMetadataInJSON(*i)))

	assert.False(t, gjson.Get(b.String(), "credentials").Exists())
}

func TestMarshalIdentityWithAdminMetadata(t *testing.T) {
	t.Parallel()

	i := NewIdentity(config.DefaultIdentityTraitsSchemaID)
	i.MetadataAdmin = []byte(`{"some":"metadata"}`)

	var b bytes.Buffer
	require.Nil(t, json.NewEncoder(&b).Encode(WithAdminMetadataInJSON(*i)))
	assert.Equal(t, "metadata", gjson.GetBytes(i.MetadataAdmin, "some").String(), "Original metadata_admin should not be touched by marshalling")
}

func TestMarshalIdentityWithCredentialsNoConfig(t *testing.T) {
	t.Parallel()

	i := NewIdentity(config.DefaultIdentityTraitsSchemaID)
	credentials := map[CredentialsType]Credentials{
		CredentialsTypePassword: {
			Type:   CredentialsTypePassword,
			Config: sqlxx.JSONRawMessage(`{"some":"secret"}`),
		},
	}
	i.Credentials = credentials
	i.MetadataAdmin = []byte(`{"some":"metadata"}`)

	rawJSON, err := json.Marshal((*WithCredentialsNoConfigAndAdminMetadataInJSON)(i))
	require.NoError(t, err)

	credentialsInJSON := gjson.GetBytes(rawJSON, "credentials")
	assert.Truef(t, credentialsInJSON.Exists(), "Credentials should be rendered to JSON, but got: %q", credentialsInJSON.Raw)

	assert.JSONEq(t, `{"password":{"type":"password","identifiers":null,"updated_at":"0001-01-01T00:00:00Z","created_at":"0001-01-01T00:00:00Z","version":0}}`, credentialsInJSON.Raw)
	assert.Equal(t, credentials, i.Credentials, "Original credentials should not be touched by marshalling")
	assert.Equal(t, "metadata", gjson.GetBytes(i.MetadataAdmin, "some").String(), "Original metadata_admin should not be touched by marshalling")
}

func TestMarshalIdentityWithAll(t *testing.T) {
	t.Parallel()

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

	credentialsInJSON := gjson.Get(b.String(), "credentials")
	assert.True(t, credentialsInJSON.Exists())

	snapshotx.SnapshotT(t, json.RawMessage(credentialsInJSON.Raw))
	assert.Equal(t, credentials, i.Credentials, "Original credentials should not be touched by marshalling")
	assert.Equal(t, "metadata", gjson.GetBytes(i.MetadataAdmin, "some").String(), "Original credentials should not be touched by marshalling")
}

func TestValidateNID(t *testing.T) {
	t.Parallel()

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
	t.Parallel()

	var addresses []RecoveryAddress

	for i := range 10 {
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
	t.Parallel()

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

type cipherProvider struct{}

func (c *cipherProvider) Cipher(context.Context) cipher.Cipher {
	return cipher.NewNoop()
}

func TestWithDeclassifiedCredentials(t *testing.T) {
	t.Parallel()

	i := NewIdentity(config.DefaultIdentityTraitsSchemaID)
	credentials := map[CredentialsType]Credentials{
		CredentialsTypePassword: {
			Identifiers: []string{"zab", "bar"},
			Type:        CredentialsTypePassword,
			Config:      sqlxx.JSONRawMessage(`{"some": "secret"}`),
		},
		CredentialsTypeOIDC: {
			Type:        CredentialsTypeOIDC,
			Identifiers: []string{"bar", "baz"},
			// hint:
			//	echo '666f6f' | xxd -r -p
			Config: sqlxx.JSONRawMessage(`{"providers": [{"subject":"bar","provider":"oidc1","initial_id_token":"666f6f"}]}`),
		},
		CredentialsTypeSAML: {
			Type:        CredentialsTypeSAML,
			Identifiers: []string{"qux", "quz"},
			// hint:
			//	echo 'this should not appear in output' | xxd -ps -c 0
			Config: sqlxx.JSONRawMessage(`{"providers": [{"subject":"qux","provider":"saml1","initial_id_token":"746869732073686f756c64206e6f742061707065617220696e206f75747075740a"}]}`),
		},
		CredentialsTypeWebAuthn: {
			Type:        CredentialsTypeWebAuthn,
			Identifiers: []string{"foo", "bar"},
			Config:      sqlxx.JSONRawMessage(`{"some": "secret"}`),
		},
	}
	i.Credentials = credentials

	t.Run("case=no-include", func(t *testing.T) {
		actualIdentity, err := i.WithDeclassifiedCredentials(ctx, &cipherProvider{}, nil)
		require.NoError(t, err)

		for ct, actual := range actualIdentity.Credentials {
			t.Run("credential="+string(ct), func(t *testing.T) {
				snapshotx.SnapshotT(t, actual)
			})
		}
	})

	t.Run("case=include-webauthn", func(t *testing.T) {
		actualIdentity, err := i.WithDeclassifiedCredentials(ctx, &cipherProvider{}, []CredentialsType{CredentialsTypeWebAuthn})
		require.NoError(t, err)

		for ct, actual := range actualIdentity.Credentials {
			t.Run("credential="+string(ct), func(t *testing.T) {
				snapshotx.SnapshotT(t, actual)
			})
		}
	})

	t.Run("case=include-multi", func(t *testing.T) {
		actualIdentity, err := i.WithDeclassifiedCredentials(ctx, &cipherProvider{}, []CredentialsType{CredentialsTypeWebAuthn, CredentialsTypePassword})
		require.NoError(t, err)

		for ct, actual := range actualIdentity.Credentials {
			t.Run("credential="+string(ct), func(t *testing.T) {
				snapshotx.SnapshotT(t, actual)
			})
		}
	})

	t.Run("case=oidc", func(t *testing.T) {
		actualIdentity, err := i.WithDeclassifiedCredentials(ctx, &cipherProvider{}, []CredentialsType{CredentialsTypeOIDC})
		require.NoError(t, err)

		for ct, actual := range actualIdentity.Credentials {
			t.Run("credential="+string(ct), func(t *testing.T) {
				snapshotx.SnapshotT(t, actual)
			})
		}
	})
	t.Run("case=saml", func(t *testing.T) {
		actualIdentity, err := i.WithDeclassifiedCredentials(ctx, &cipherProvider{}, []CredentialsType{CredentialsTypeSAML})
		require.NoError(t, err)

		for ct, actual := range actualIdentity.Credentials {
			t.Run("credential="+string(ct), func(t *testing.T) {
				snapshotx.SnapshotT(t, actual)
			})
		}
	})
}

func TestDeleteCredentialOIDCSAMLFromIdentity(t *testing.T) {
	t.Parallel()

	i := NewIdentity(config.DefaultIdentityTraitsSchemaID)

	err := i.deleteCredentialOIDCSAMLFromIdentity(CredentialsTypeOIDC, "")
	assert.Error(t, err)
	err = i.deleteCredentialOIDCSAMLFromIdentity(CredentialsTypeOIDC, "does-not-exist")
	assert.Error(t, err)
	err = i.deleteCredentialOIDCSAMLFromIdentity(CredentialsTypeSAML, "")
	assert.Error(t, err)
	err = i.deleteCredentialOIDCSAMLFromIdentity(CredentialsTypeSAML, "does-not-exist")
	assert.Error(t, err)

	credentials := map[CredentialsType]Credentials{
		CredentialsTypePassword: {
			Identifiers: []string{"zab", "bar"},
			Type:        CredentialsTypePassword,
			Config:      sqlxx.JSONRawMessage(`{"some" : "secret"}`),
		},
		CredentialsTypeOIDC: {
			Type:        CredentialsTypeOIDC,
			Identifiers: []string{"bar:1234", "baz:5678"},
			Config:      sqlxx.JSONRawMessage(`{"providers": [{"provider": "bar", "subject": "1234"}, {"provider": "baz", "subject": "5678"}]}`),
		},
		CredentialsTypeSAML: {
			Type:        CredentialsTypeSAML,
			Identifiers: []string{"bar:1234", "baz:5678"},
			Config:      sqlxx.JSONRawMessage(`{"providers": [{"provider": "bar", "subject": "1234"}, {"provider": "baz", "subject": "5678"}]}`),
		},
		CredentialsTypeWebAuthn: {
			Type:        CredentialsTypeWebAuthn,
			Identifiers: []string{"foo", "bar"},
			Config:      sqlxx.JSONRawMessage(`{"some" : "secret"}`),
		},
	}
	i.Credentials = credentials

	err = i.deleteCredentialOIDCSAMLFromIdentity(CredentialsTypeOIDC, "zab")
	assert.Error(t, err)
	err = i.deleteCredentialOIDCSAMLFromIdentity(CredentialsTypeOIDC, "foo")
	assert.Error(t, err)
	err = i.deleteCredentialOIDCSAMLFromIdentity(CredentialsTypeOIDC, "bar")
	assert.Error(t, err, "matches multiple OIDC credentials")

	require.NoError(t, i.deleteCredentialOIDCSAMLFromIdentity(CredentialsTypeOIDC, "bar:1234"))

	assert.Len(t, i.Credentials, 4)

	assert.Contains(t, i.Credentials, CredentialsTypePassword)
	assert.EqualValues(t, i.Credentials[CredentialsTypePassword].Identifiers, []string{"zab", "bar"})

	assert.Contains(t, i.Credentials, CredentialsTypeWebAuthn)
	assert.EqualValues(t, i.Credentials[CredentialsTypeWebAuthn].Identifiers, []string{"foo", "bar"})

	assert.Contains(t, i.Credentials, CredentialsTypeSAML)
	assert.EqualValues(t, i.Credentials[CredentialsTypeSAML].Identifiers, []string{"bar:1234", "baz:5678"})

	assert.Contains(t, i.Credentials, CredentialsTypeOIDC)

	oidc, ok := i.GetCredentials(CredentialsTypeOIDC)
	require.True(t, ok)
	assert.EqualValues(t, oidc.Identifiers, []string{"baz:5678"})
	var cfg CredentialsOIDC
	_, err = i.ParseCredentials(CredentialsTypeOIDC, &cfg)
	require.NoError(t, err)
	assert.EqualValues(t, CredentialsOIDC{Providers: []CredentialsOIDCProvider{{Provider: "baz", Subject: "5678"}}}, cfg)

	require.NoError(t, i.deleteCredentialOIDCSAMLFromIdentity(CredentialsTypeSAML, "baz:5678"))

	assert.Len(t, i.Credentials, 4)
	assert.Contains(t, i.Credentials, CredentialsTypeOIDC)
	assert.EqualValues(t, i.Credentials[CredentialsTypeOIDC].Identifiers, []string{"baz:5678"})

	assert.Contains(t, i.Credentials, CredentialsTypeSAML)
	saml, ok := i.GetCredentials(CredentialsTypeSAML)
	require.True(t, ok)
	assert.EqualValues(t, saml.Identifiers, []string{"bar:1234"})
	var samlCfg CredentialsOIDC
	_, err = i.ParseCredentials(CredentialsTypeSAML, &samlCfg)
	require.NoError(t, err)
	assert.EqualValues(t, CredentialsOIDC{Providers: []CredentialsOIDCProvider{{Provider: "bar", Subject: "1234"}}}, samlCfg)
}

func TestMergeOIDCCredentials(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name           string
		identity       *Identity
		newCredentials Credentials

		expectedIdentity *Identity
		assertErr        assert.ErrorAssertionFunc
	}{
		{
			name:     "adds OIDC credential if not exists",
			identity: &Identity{},
			newCredentials: Credentials{
				Type:        CredentialsTypeOIDC,
				Identifiers: []string{"oidc:1234"},
				Config:      sqlxx.JSONRawMessage(`{"providers":[{"provider":"oidc","subject":"1234"}]}`),
			},

			expectedIdentity: &Identity{
				Credentials: map[CredentialsType]Credentials{
					CredentialsTypeOIDC: {
						Type:        CredentialsTypeOIDC,
						Identifiers: []string{"oidc:1234"},
						Config:      sqlxx.JSONRawMessage(`{"providers":[{"provider":"oidc","subject":"1234"}]}`),
					},
				},
			},
		},
		{
			name: "merges OIDC credential if exists",
			identity: &Identity{
				Credentials: map[CredentialsType]Credentials{
					CredentialsTypePassword: {
						Type:        CredentialsTypePassword,
						Identifiers: []string{"user@example.com"},
					},
					CredentialsTypeOIDC: {
						Type:        CredentialsTypeOIDC,
						Identifiers: []string{"foo", "replace:1234", "bar", "baz", "replace:abc", "replace:dont-replace"},
						Config: sqlxx.JSONRawMessage(`{"providers": [
	{"provider": "replace", "subject": "1234", "use_auto_link": true},
	{"provider": "dont-touch", "subject": "foo"},
	{"provider": "replace", "subject": "abc", "use_auto_link": true},
	{"provider": "also-dont-touch", "subject": "bar", "use_auto_link": true},
	{"provider": "replace", "subject": "dont-replace", "use_auto_link": false}
]}`),
					},
				},
			},
			newCredentials: Credentials{
				Type:        CredentialsTypeOIDC,
				Identifiers: []string{},
				Config:      sqlxx.JSONRawMessage(`{"providers": [{"provider": "replace", "subject": "new-subject"}]}`),
			},

			expectedIdentity: &Identity{
				Credentials: map[CredentialsType]Credentials{
					CredentialsTypeOIDC: {
						Type:        CredentialsTypeOIDC,
						Identifiers: []string{"foo", "bar", "baz", "replace:dont-replace", "replace:new-subject"},
						Config: sqlxx.JSONRawMessage(`{
  "providers" : [ {
    "subject" : "foo",
    "provider" : "dont-touch",
    "initial_id_token" : "",
    "initial_access_token" : "",
    "initial_refresh_token" : ""
  }, {
    "subject" : "bar",
    "provider" : "also-dont-touch",
    "initial_id_token" : "",
    "initial_access_token" : "",
    "initial_refresh_token" : "",
    "use_auto_link": true
  }, {
    "subject" : "dont-replace",
    "provider" : "replace",
    "initial_id_token" : "",
    "initial_access_token" : "",
    "initial_refresh_token" : ""
  }, {
    "subject" : "new-subject",
    "provider" : "replace",
    "initial_id_token" : "",
    "initial_access_token" : "",
    "initial_refresh_token" : ""
  } ]
}`),
					},
					CredentialsTypePassword: {
						Type:        CredentialsTypePassword,
						Identifiers: []string{"user@example.com"},
					},
				},
			},
		},
		{
			name: "errs if new credential has no provider",
			identity: &Identity{
				Credentials: map[CredentialsType]Credentials{
					CredentialsTypeOIDC: {
						Type:        CredentialsTypeOIDC,
						Identifiers: []string{"oidc:1234"},
						Config:      sqlxx.JSONRawMessage(`{"providers": [{"provider": "oidc", "subject": "1234"}]}`),
					},
				},
			},
			newCredentials: Credentials{
				Type:        CredentialsTypeOIDC,
				Identifiers: []string{"oidc:1234"},
				Config:      sqlxx.JSONRawMessage(`{"providers": []}`),
			},

			assertErr: assert.Error,
		},
		{
			name: "errs if identity credentials are invalid",
			identity: &Identity{
				Credentials: map[CredentialsType]Credentials{
					CredentialsTypeOIDC: {
						Type:        CredentialsTypeOIDC,
						Identifiers: []string{"oidc:1234"},
						Config:      sqlxx.JSONRawMessage("invalid"),
					},
				},
			},
			newCredentials: Credentials{
				Type:        CredentialsTypeOIDC,
				Identifiers: []string{"oidc:1234"},
				Config:      sqlxx.JSONRawMessage(`{"providers": [{"provider": "replace", "subject": "new-subject"}]}`),
			},

			assertErr: assert.Error,
		},
		{
			name: "errs if new credential config is invalid",
			identity: &Identity{
				Credentials: map[CredentialsType]Credentials{
					CredentialsTypeOIDC: {
						Type:        CredentialsTypeOIDC,
						Identifiers: []string{"oidc:1234"},
						Config:      sqlxx.JSONRawMessage(`{"providers": [{"provider": "oidc", "subject": "1234"}]}`),
					},
				},
			},
			newCredentials: Credentials{
				Type:        CredentialsTypeOIDC,
				Identifiers: []string{"oidc:1234"},
				Config:      sqlxx.JSONRawMessage(`invalid`),
			},

			assertErr: assert.Error,
		},
	} {
		t.Run("case="+tc.name, func(t *testing.T) {
			err := tc.identity.MergeOIDCCredentials(CredentialsTypeOIDC, tc.newCredentials)

			if tc.assertErr != nil {
				tc.assertErr(t, err)
			} else {
				require.NoError(t, err)
			}

			if tc.expectedIdentity != nil {
				var buf bytes.Buffer
				require.NoError(t, json.Compact(&buf, tc.expectedIdentity.Credentials[CredentialsTypeOIDC].Config))
				tc.expectedIdentity.UpsertCredentialsConfig(CredentialsTypeOIDC, buf.Bytes(), 0)
				assert.EqualExportedValues(t, tc.expectedIdentity, tc.identity)
			}
		})
	}
}
