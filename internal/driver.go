package internal

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ory/viper"

	"github.com/ory/x/logrusx"

	"github.com/ory/hive-cloud/hive/driver"
	"github.com/ory/hive-cloud/hive/driver/configuration"
)

func resetConfig() {
	viper.Set(configuration.ViperKeyDSN, nil)

	viper.Set("LOG_LEVEL", "debug")
	viper.Set(configuration.ViperKeyHasherArgon2ConfigMemory, 64)
	viper.Set(configuration.ViperKeyHasherArgon2ConfigIterations, 1)
	viper.Set(configuration.ViperKeyHasherArgon2ConfigParallelism, 1)
	viper.Set(configuration.ViperKeyHasherArgon2ConfigSaltLength, 2)
	viper.Set(configuration.ViperKeyHasherArgon2ConfigKeyLength, 2)
}

func NewConfigurationWithDefaults() *configuration.ViperProvider {
	resetConfig()
	return configuration.NewViperProvider(logrusx.New())
}

func NewMemoryRegistry(t *testing.T) (*configuration.ViperProvider, *driver.RegistryMemory) {
	conf := NewConfigurationWithDefaults()
	viper.Set(configuration.ViperKeyDSN, "memory")

	reg, err := driver.NewRegistry(conf)
	require.NoError(t, err)
	return conf, reg.WithConfig(conf).(*driver.RegistryMemory)
}
