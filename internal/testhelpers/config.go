// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package testhelpers

import (
	"context"
	"encoding/base64"
	"testing"

	confighelpers "github.com/ory/kratos/driver/config/testhelpers"

	"github.com/spf13/pflag"
	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/x/configx"
	"github.com/ory/x/randx"
)

func UseConfigFile(t *testing.T, path string) *pflag.FlagSet {
	flags := pflag.NewFlagSet("config", pflag.ContinueOnError)
	configx.RegisterFlags(flags)
	require.NoError(t, flags.Parse([]string{"--config", path}))
	return flags
}

func DefaultIdentitySchemaConfig(url string) map[string]any {
	return map[string]any{
		config.ViperKeyDefaultIdentitySchemaID: "default",
		config.ViperKeyIdentitySchemas: config.Schemas{
			{ID: "default", URL: url},
		},
	}
}

func WithDefaultIdentitySchema(ctx context.Context, url string) context.Context {
	return confighelpers.WithConfigValues(ctx, DefaultIdentitySchemaConfig(url))
}

// Deprecated: Use context-based WithDefaultIdentitySchema instead
func SetDefaultIdentitySchema(conf *config.Config, url string) func() {
	schemaUrl, _ := conf.DefaultIdentityTraitsSchemaURL(context.Background())
	conf.MustSet(context.Background(), config.ViperKeyDefaultIdentitySchemaID, "default")
	conf.MustSet(context.Background(), config.ViperKeyIdentitySchemas, config.Schemas{
		{ID: "default", URL: url},
	})
	return func() {
		conf.MustSet(context.Background(), config.ViperKeyIdentitySchemas, config.Schemas{
			{ID: "default", URL: schemaUrl.String()},
		})
	}
}

// WithAddIdentitySchema registers an identity schema in the config with a random ID and returns the ID
//
// It also registers a test cleanup function, to reset the schemas to the original values, after the test finishes
func WithAddIdentitySchema(ctx context.Context, t *testing.T, conf *config.Config, url string) (context.Context, string) {
	id := randx.MustString(16, randx.Alpha)
	schemas, err := conf.IdentityTraitsSchemas(ctx)
	require.NoError(t, err)

	return confighelpers.WithConfigValue(ctx, config.ViperKeyIdentitySchemas, append(schemas, config.Schema{
		ID:  id,
		URL: url,
	})), id
}

// UseIdentitySchema registers an identity schema in the config with a random ID and returns the ID
//
// It also registers a test cleanup function, to reset the schemas to the original values, after the test finishes
// Deprecated: Use context-based WithAddIdentitySchema instead
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

// WithDefaultIdentitySchemaFromRaw allows setting the default identity schema from a raw JSON string.
func WithDefaultIdentitySchemaFromRaw(ctx context.Context, schema []byte) context.Context {
	return WithDefaultIdentitySchema(ctx, "base64://"+base64.URLEncoding.EncodeToString(schema))
}

// Deprecated: Use context-based WithDefaultIdentitySchemaFromRaw instead
func SetDefaultIdentitySchemaFromRaw(conf *config.Config, schema []byte) {
	conf.MustSet(context.Background(), config.ViperKeyDefaultIdentitySchemaID, "default")
	conf.MustSet(context.Background(), config.ViperKeyIdentitySchemas, config.Schemas{
		{ID: "default", URL: "base64://" + base64.URLEncoding.EncodeToString(schema)},
	})
}
