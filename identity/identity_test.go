package identity

import (
	"encoding/json"
	"testing"

	"github.com/ory/kratos/driver/configuration"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewIdentity(t *testing.T) {
	i := NewIdentity(configuration.DefaultIdentityTraitsSchemaID)
	assert.NotEmpty(t, i.ID)
	// assert.NotEmpty(t, i.Metadata)
	assert.NotEmpty(t, i.Traits)
	assert.NotNil(t, i.Credentials)
}

func TestCopyCredentials(t *testing.T) {
	i := NewIdentity(configuration.DefaultIdentityTraitsSchemaID)
	i.Credentials = map[CredentialsType]Credentials{
		"foo": {
			Type:        "foo",
			Identifiers: []string{"bar"},
			Config:      json.RawMessage(`{"foo":"bar"}`),
		},
	}

	creds := i.CopyCredentials()
	creds["bar"] = Credentials{
		Type:        "bar",
		Identifiers: []string{"bar"},
		Config:      json.RawMessage(`{"foo":"bar"}`),
	}

	conf := creds["foo"].Config
	require.NoError(t, json.Unmarshal([]byte(`{"bar":"bar"}`), &conf))

	assert.Empty(t, i.Credentials["bar"])
	assert.Equal(t, `{"foo":"bar"}`, string(i.Credentials["foo"].Config))
	assert.Equal(t, `{"bar":"bar"}`, string(creds["foo"].Config))
}
