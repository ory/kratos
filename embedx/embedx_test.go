package embedx

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ory/jsonschema/v3"
)

func TestAddSchemaResources(t *testing.T) {

	t.Run("case=validate schema resources are loading schema id correctly", func(t *testing.T) {
		c := jsonschema.NewCompiler()
		assert.NoError(t, AddSchemaResources(c, Config, IdentityMeta, IdentityExtension))

		assert.EqualValues(t, ConfigSchemaID.ToString(), Config.GetSchemaID())
		assert.EqualValues(t, IdentityMetaSchemaID, IdentityMeta.GetSchemaID())
		assert.EqualValues(t, IdentityExtensionSchemaID, IdentityExtension.GetSchemaID())

		assert.EqualValues(t, ConfigSchemaID.ToString(), "https://github.com/ory/kratos/embedx/config.schema.json")
		assert.EqualValues(t, IdentityMetaSchemaID.ToString(), "ory://identity-meta")
		assert.EqualValues(t, IdentityExtensionSchemaID.ToString(), "ory://identity-extension")
	})

	t.Run("case=add config schema resources", func(t *testing.T) {
		c := jsonschema.NewCompiler()
		assert.NoError(t, AddSchemaResources(c, Config))

		_, err := c.Compile(ConfigSchemaID.ToString())
		assert.NoError(t, err)
	})

	t.Run("case=add identity schema resources", func(t *testing.T) {
		c := jsonschema.NewCompiler()
		assert.NoError(t, AddSchemaResources(c, IdentityMeta))

		_, err := c.Compile(IdentityMetaSchemaID.ToString())
		assert.NoError(t, err)
	})

	t.Run("case=add multiple schema resources", func(t *testing.T) {
		c := jsonschema.NewCompiler()
		assert.NoError(t, AddSchemaResources(c, Config, IdentityMeta))

		_, err := c.Compile(IdentityMetaSchemaID.ToString())
		assert.NoError(t, err)

		_, err = c.Compile(ConfigSchemaID.ToString())
	})

	t.Run("case=dynamic multiple schema resourcces", func(t *testing.T) {
		c := jsonschema.NewCompiler()

		assert.NoError(t, AddSchemaResources(c, Config, IdentityExtension, IdentityMeta))

		_, err := c.Compile(Config.GetSchemaID())
		assert.NoError(t, err)

		_, err = c.Compile(IdentityExtension.GetSchemaID())
		assert.NoError(t, err)

		_, err = c.Compile(IdentityMeta.GetSchemaID())
		assert.NoError(t, err)
	})

	t.Run("case=specifying an incorrect schema type must throw an error", func(t *testing.T) {
		c := jsonschema.NewCompiler()
		err := AddSchemaResources(c, 4, 10)
		assert.Errorf(t, err, "an error must be thrown on an invalid schema type")
		assert.EqualError(t, err, "the specified schema type (4) is not supported")
	})
}
