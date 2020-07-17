package testhelpers

import (
	"github.com/ory/viper"

	"github.com/ory/kratos/driver/configuration"
)

// SetDefaultIdentitySchema sets the identity schemas in viper config:
//
//	testhelpers.SetDefaultIdentitySchema("file://customer.json")
func SetDefaultIdentitySchema(location string) {
	viper.Set(configuration.ViperKeyDefaultIdentitySchemaURL, location)
}

// SetIdentitySchemas sets the identity schemas in viper config:
//
//	testhelpers.SetIdentitySchemas(map[string]string{"customer": "file://customer.json"})
func SetIdentitySchemas(schemas map[string]string) {
	var s []configuration.SchemaConfig
	for id, location := range schemas {
		s = append(s, configuration.SchemaConfig{ID: id, URL: location})
	}

	viper.Set(configuration.ViperKeyIdentitySchemas, s)
}
