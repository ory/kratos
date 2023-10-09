// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package testhelpers

import (
	"context"
	"encoding/base64"
	"testing"

	"github.com/ory/kratos/driver/config"

	"github.com/spf13/pflag"
	"github.com/stretchr/testify/require"

	"github.com/ory/x/configx"
	"github.com/ory/x/randx"
)

func UseConfigFile(t *testing.T, path string) *pflag.FlagSet {
	flags := pflag.NewFlagSet("config", pflag.ContinueOnError)
	configx.RegisterFlags(flags)
	require.NoError(t, flags.Parse([]string{"--config", path}))
	return flags
}

func SetDefaultIdentitySchema(conf *config.Config, url string) {
	conf.MustSet(context.Background(), config.ViperKeyDefaultIdentitySchemaID, "default")
	conf.MustSet(context.Background(), config.ViperKeyIdentitySchemas, config.Schemas{
		{ID: "default", URL: url},
	})
}

// UseIdentitySchema registeres an identity schema in the config with a random ID and returns the ID
//
// It also registeres a test cleanup function, to reset the schemas to the original values, after the test finishes
func UseIdentitySchema(t *testing.T, conf *config.Config, url string) (id string) {
	id = randx.MustString(16, randx.Alpha)
	schemas, err := conf.IdentityTraitsSchemas(context.Background())
	require.NoError(t, err)
	conf.MustSet(context.Background(), config.ViperKeyIdentitySchemas, append(schemas, config.Schema{
		ID:  id,
		URL: url,
	}))
	t.Cleanup(func() {
		conf.MustSet(context.Background(), config.ViperKeyIdentitySchemas, schemas)
	})
	return id
}

// SetDefaultIdentitySchemaFromRaw allows setting the default identity schema from a raw JSON string.
func SetDefaultIdentitySchemaFromRaw(conf *config.Config, schema []byte) {
	conf.MustSet(context.Background(), config.ViperKeyDefaultIdentitySchemaID, "default")
	conf.MustSet(context.Background(), config.ViperKeyIdentitySchemas, config.Schemas{
		{ID: "default", URL: "base64://" + base64.URLEncoding.EncodeToString(schema)},
	})
}
