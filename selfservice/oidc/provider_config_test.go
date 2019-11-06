package oidc_test

import (
	"encoding/json"
	"testing"

	"github.com/bmizerany/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/viper"

	"github.com/ory/kratos/driver/configuration"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/internal"
	. "github.com/ory/kratos/selfservice/oidc"
)

func TestConfig(t *testing.T) {
	conf, reg := internal.NewMemoryRegistry(t)

	viper.Set(configuration.ViperKeySelfServiceStrategyConfig+"."+string(identity.CredentialsTypeOIDC), json.RawMessage(`{"config":{"providers": [{"provider": "generic"}]}}`))
	s := NewStrategy(reg, conf)

	collection, err := s.Config()
	require.NoError(t, err)

	require.Len(t, collection.Providers, 1)
	assert.Equal(t, "generic", collection.Providers[0].Provider)
}
