package oidc_test

import (
	"encoding/json"
	"testing"

	"github.com/bmizerany/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/viper"

	"github.com/ory/hive/driver/configuration"
	"github.com/ory/hive/identity"
	"github.com/ory/hive/internal"
	. "github.com/ory/hive/selfservice/oidc"
)

func TestConfig(t *testing.T) {
	viper.Set(configuration.ViperKeySelfServiceStrategyConfig+"."+string(identity.CredentialsTypeOIDC), json.RawMessage(`{"config":{"providers": [{"provider": "generic"}]}}`))

	conf, reg := internal.NewMemoryRegistry(t)
	s := NewStrategy(reg, conf)

	collection, err := s.Config()
	require.NoError(t, err)

	require.Len(t, collection.Providers, 1)
	assert.Equal(t, "generic", collection.Providers[0].Provider)
}
