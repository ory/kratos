// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package internal

import (
	"cmp"
	"context"
	"runtime"
	"testing"

	"github.com/gofrs/uuid"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/driver"
	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/embedx"
	"github.com/ory/kratos/selfservice/hook"
	"github.com/ory/kratos/x/nosurfx"
	"github.com/ory/x/configx"
	"github.com/ory/x/contextx"
	"github.com/ory/x/dbal"
	"github.com/ory/x/jsonnetsecure"
	"github.com/ory/x/logrusx"
	"github.com/ory/x/randx"
)

func NewConfigurationWithDefaults(t testing.TB, opts ...configx.OptionModifier) *config.Config {
	configOpts := append([]configx.OptionModifier{
		configx.WithValues(map[string]interface{}{
			"log.level":                                      "error",
			config.ViperKeyDSN:                               dbal.NewSQLiteTestDatabase(t),
			config.ViperKeyHasherArgon2ConfigMemory:          16384,
			config.ViperKeyHasherArgon2ConfigIterations:      1,
			config.ViperKeyHasherArgon2ConfigParallelism:     1,
			config.ViperKeyHasherArgon2ConfigSaltLength:      16,
			config.ViperKeyHasherBcryptCost:                  4,
			config.ViperKeyHasherArgon2ConfigKeyLength:       16,
			config.ViperKeyCourierSMTPURL:                    "smtp://foo:bar@baz.com/",
			config.ViperKeySelfServiceBrowserDefaultReturnTo: "https://www.ory.sh/redirect-not-set",
			config.ViperKeySecretsCipher:                     []string{"secret-thirty-two-character-long"},
			config.ViperKeySecretsPagination:                 []string{uuid.Must(uuid.NewV4()).String()},
			config.ViperKeySelfServiceLoginFlowStyle:         "unified",
		}),
		configx.SkipValidation(),
	}, opts...)
	return config.MustNew(t, logrusx.New("", ""),
		contextx.NewTestConfigProvider(embedx.ConfigSchema, configOpts...),
		configOpts...,
	)
}

// NewFastRegistryWithMocks returns a registry with several mocks and an SQLite in memory database that make testing
// easier and way faster. This suite does not work for e2e or advanced integration tests.
func NewFastRegistryWithMocks(t *testing.T, opts ...configx.OptionModifier) (*config.Config, *driver.RegistryDefault) {
	conf, reg := NewRegistryDefaultWithDSN(t, "", opts...)
	reg.WithCSRFTokenGenerator(nosurfx.FakeCSRFTokenGenerator)
	reg.WithCSRFHandler(nosurfx.NewFakeCSRFHandler(""))
	reg.WithHooks(map[string]func(config.SelfServiceHook) interface{}{
		"err": func(c config.SelfServiceHook) interface{} {
			return &hook.Error{Config: c.Config}
		},
	})
	reg.SetJSONNetVMProvider(jsonnetsecure.NewTestProvider(t))

	require.NoError(t, reg.Persister().MigrateUp(context.Background()))
	require.NotEqual(t, uuid.Nil, reg.Persister().NetworkID(context.Background()))
	return conf, reg
}

// NewRegistryDefaultWithDSN returns a more standard registry without mocks. Good for e2e and advanced integration testing!
func NewRegistryDefaultWithDSN(t testing.TB, dsn string, opts ...configx.OptionModifier) (*config.Config, *driver.RegistryDefault) {
	ctx := context.Background()
	c := NewConfigurationWithDefaults(t, append([]configx.OptionModifier{configx.WithValues(map[string]interface{}{
		config.ViperKeyDSN:             cmp.Or(dsn, dbal.NewSQLiteTestDatabase(t)+"&lock=false&max_conns=1"),
		"dev":                          true,
		config.ViperKeySecretsCipher:   []string{randx.MustString(32, randx.AlphaNum)},
		config.ViperKeySecretsCookie:   []string{randx.MustString(32, randx.AlphaNum)},
		config.ViperKeySecretsDefault:  []string{randx.MustString(32, randx.AlphaNum)},
		config.ViperKeyCipherAlgorithm: "xchacha20-poly1305",
	})}, opts...)...)
	reg, err := driver.NewRegistryFromDSN(ctx, c, logrusx.New("", "", logrusx.ForceLevel(logrus.ErrorLevel)))
	require.NoError(t, err)
	pool := jsonnetsecure.NewProcessPool(runtime.GOMAXPROCS(0))
	t.Cleanup(pool.Close)
	require.NoError(t, reg.Init(context.Background(), contextx.NewTestConfigProvider(embedx.ConfigSchema), driver.SkipNetworkInit, driver.WithDisabledMigrationLogging(), driver.WithJsonnetPool(pool)))
	require.NoError(t, reg.Persister().MigrateUp(context.Background())) // always migrate up

	actual, err := reg.Persister().DetermineNetwork(context.Background())
	require.NoError(t, err)
	reg.SetPersister(reg.Persister().WithNetworkID(actual.ID))

	require.EqualValues(t, reg.Persister().NetworkID(context.Background()), actual.ID)
	require.NotEqual(t, uuid.Nil, reg.Persister().NetworkID(context.Background()))
	reg.Persister()

	return c, reg
}

func NewVeryFastRegistryWithoutDB(t *testing.T) (*config.Config, *driver.RegistryDefault) {
	c := NewConfigurationWithDefaults(t)
	reg, err := driver.NewRegistryFromDSN(context.Background(), c, logrusx.New("", ""))
	require.NoError(t, err)
	return c, reg
}
