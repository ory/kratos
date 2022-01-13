package testhelpers

import (
	"testing"

	"github.com/ory/kratos/driver/config"

	"github.com/spf13/pflag"
	"github.com/stretchr/testify/require"

	"github.com/ory/x/configx"
)

func UseConfigFile(t *testing.T, path string) *pflag.FlagSet {
	flags := pflag.NewFlagSet("config", pflag.ContinueOnError)
	configx.RegisterFlags(flags)
	require.NoError(t, flags.Parse([]string{"--config", path}))
	return flags
}

func SetDefaultIdentitySchema(conf *config.Config, url string) {
	conf.MustSet(config.ViperKeyDefaultIdentitySchemaID, "default")
	conf.MustSet(config.ViperKeyIdentitySchemas, config.Schemas{
		{ID: "default", URL: url},
	})
}
