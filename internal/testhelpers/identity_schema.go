package testhelpers

import (
	"testing"

	"github.com/ory/kratos/driver/config"
)

// SetDefaultIdentitySchema sets the identity schemas in viper config:
//
//	testhelpers.SetDefaultIdentitySchema("file://customer.json")
func SetDefaultIdentitySchema(t *testing.T, conf *config.Config, location string) {
	conf.MustSet(config.ViperKeyDefaultIdentitySchemaURL, location)
}

// SetIdentitySchemas sets the identity schemas in viper config:
//
//	testhelpers.SetIdentitySchemas(map[string]string{"customer": "file://customer.json"})
func SetIdentitySchemas(t *testing.T, conf *config.Config, schemas map[string]string) {
	var s []config.SchemaConfig
	for id, location := range schemas {
		s = append(s, config.SchemaConfig{ID: id, URL: location})
	}

	conf.MustSet(config.ViperKeyIdentitySchemas, s)
}
