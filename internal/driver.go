package internal

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ory/viper"

	"github.com/ory/x/logrusx"

	"github.com/ory/kratos/driver"
	"github.com/ory/kratos/driver/configuration"
	"github.com/ory/kratos/x"
)

func resetConfig() {
	viper.Set(configuration.ViperKeyDSN, nil)

	viper.Set("LOG_LEVEL", "trace")
	viper.Set(configuration.ViperKeyHasherArgon2ConfigMemory, 64)
	viper.Set(configuration.ViperKeyHasherArgon2ConfigIterations, 1)
	viper.Set(configuration.ViperKeyHasherArgon2ConfigParallelism, 1)
	viper.Set(configuration.ViperKeyHasherArgon2ConfigSaltLength, 2)
	viper.Set(configuration.ViperKeyHasherArgon2ConfigKeyLength, 2)
}

func NewConfigurationWithDefaults() *configuration.ViperProvider {
	viper.Reset()
	resetConfig()
	return configuration.NewViperProvider(logrusx.New())
}

func NewRegistryDefault(t *testing.T) (*configuration.ViperProvider, *driver.RegistryDefault) {
	conf, reg := NewRegistryDefaultWithDSN(t, "")
	require.NoError(t, reg.Persister().MigrateUp(context.Background()))
	return conf, reg
}

func NewRegistryDefaultWithDSN(t *testing.T, dsn string) (*configuration.ViperProvider, *driver.RegistryDefault) {
	viper.Reset()
	resetConfig()

	viper.Set(configuration.ViperKeyDSN, "sqlite3://"+filepath.Join(os.TempDir(), x.NewUUID().String())+".sql?mode=memory&_fk=true")
	if dsn != "" {
		viper.Set(configuration.ViperKeyDSN, dsn)
	}

	d, err := driver.NewDefaultDriver(logrusx.New(), "test", "test", "test")
	require.NoError(t, err)
	return d.Configuration().(*configuration.ViperProvider), d.Registry().(*driver.RegistryDefault)
}
