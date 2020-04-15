package driver_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ory/kratos/driver/configuration"
	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/schema"
	"github.com/ory/viper"
	"github.com/ory/x/urlx"
)

func TestRegistryDefault_IdentityTraitsSchemas(t *testing.T) {
	_, reg := internal.NewFastRegistryWithMocks(t)
	defaultSchema := schema.Schema{
		ID:     "default",
		URL:    urlx.ParseOrPanic("file://default.schema.json"),
		RawURL: "file://default.schema.json",
	}
	altSchema := schema.Schema{
		ID:     "alt",
		URL:    urlx.ParseOrPanic("file://other.schema.json"),
		RawURL: "file://other.schema.json",
	}

	viper.Set(configuration.ViperKeyDefaultIdentityTraitsSchemaURL, defaultSchema.RawURL)
	viper.Set(configuration.ViperKeyIdentityTraitsSchemas, []configuration.SchemaConfig{{ID: altSchema.ID, URL: altSchema.RawURL}})

	ss := reg.IdentityTraitsSchemas()
	assert.Equal(t, 2, len(ss))
	assert.Contains(t, ss, defaultSchema)
	assert.Contains(t, ss, altSchema)
}
