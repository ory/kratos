package identity

import (
	"encoding/json"
	"testing"

	"github.com/gofrs/uuid"

	"github.com/ory/x/sqlxx"

	"github.com/ory/kratos/driver/config"

	"github.com/stretchr/testify/assert"
)

func TestNewIdentity(t *testing.T) {
	i := NewIdentity(config.DefaultIdentityTraitsSchemaID)
	assert.NotEmpty(t, i.ID)
	// assert.NotEmpty(t, i.Metadata)
	assert.NotEmpty(t, i.Traits)
	assert.NotNil(t, i.Credentials)
}

func TestMarshalExcludesCredentials(t *testing.T) {
	i := NewIdentity(config.DefaultIdentityTraitsSchemaID)
	credentials := map[CredentialsType]Credentials{
		CredentialsTypePassword: Credentials{
			ID: uuid.UUID{},
		},
	}

	i.Credentials = credentials
	jsonBytes, err := json.Marshal(i)
	assert.Nil(t, err)
	var jsonMap = map[string]json.RawMessage{}
	err = json.Unmarshal(jsonBytes, &jsonMap)
	assert.Nil(t, err)
	assert.Nil(t, jsonMap["credentials"])
	assert.Equal(t, credentials, i.Credentials, "Original credentials should not be touched by marshalling")
}

func TestMarshalExcludesCredentialsByReference(t *testing.T) {
	i := NewIdentity(config.DefaultIdentityTraitsSchemaID)
	credentials := map[CredentialsType]Credentials{
		CredentialsTypePassword: Credentials{
			Type: CredentialsTypePassword,
		},
	}
	i.Credentials = credentials
	jsonBytes, err := json.Marshal(&i)
	assert.Nil(t, err)
	var jsonMap = map[string]json.RawMessage{}
	err = json.Unmarshal(jsonBytes, &jsonMap)
	assert.Nil(t, err)
	assert.Nil(t, jsonMap["credentials"])
	assert.Equal(t, credentials, i.Credentials, "Original credentials should not be touched by marshalling")
}

func TestUnMarshallIgnoresCredentials(t *testing.T) {
	jsonText := "{\"id\":\"3234ad11-49c6-49e2-bfac-537f3e06cd85\",\"schema_id\":\"default\",\"schema_url\":\"\",\"traits\":{}, \"credentials\" : {\"password\":{\"type\":\"\",\"identifiers\":null,\"config\":null,\"updatedAt\":\"0001-01-01T00:00:00Z\"}}}"
	var i Identity
	err := json.Unmarshal([]byte(jsonText), &i)
	assert.Nil(t, err)

	assert.Nil(t, i.Credentials)
	assert.Equal(t, "3234ad11-49c6-49e2-bfac-537f3e06cd85", i.ID.String())
}

func TestMarshalIdentityWithCredentialsWhenCredentialsNil(t *testing.T) {
	i := NewIdentity(config.DefaultIdentityTraitsSchemaID)
	i.Credentials = nil
	jsonBytes, err := json.Marshal(IdentityWithCredentialsMetadataInJSON(*i))
	assert.Nil(t, err)
	var jsonMap = map[string]json.RawMessage{}
	err = json.Unmarshal(jsonBytes, &jsonMap)
	assert.Nil(t, err)
	assert.Nil(t, jsonMap["credentials"])
}

func TestMarshalIdentityWithCredentials(t *testing.T) {
	i := NewIdentity(config.DefaultIdentityTraitsSchemaID)
	credentials := map[CredentialsType]Credentials{
		CredentialsTypePassword: Credentials{
			Type:   CredentialsTypePassword,
			Config: sqlxx.JSONRawMessage("{\"some\" : \"secret\"}"),
		},
	}
	i.Credentials = credentials

	jsonBytes, err := json.Marshal(IdentityWithCredentialsMetadataInJSON(*i))
	assert.Nil(t, err)
	var jsonMap = map[string]json.RawMessage{}
	err = json.Unmarshal(jsonBytes, &jsonMap)
	assert.Nil(t, err)
	assert.NotNil(t, jsonMap["credentials"])
	assert.JSONEq(t, "{\"password\":{\"type\":\"password\",\"identifiers\":null,\"updated_at\":\"0001-01-01T00:00:00Z\",\"created_at\":\"0001-01-01T00:00:00Z\"}}", string(jsonMap["credentials"]))
	assert.Equal(t, credentials, i.Credentials, "Original credentials should not be touched by marshalling")
}
