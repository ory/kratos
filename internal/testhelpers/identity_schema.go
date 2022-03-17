package testhelpers

import (
	"testing"

	"github.com/ory/kratos/driver/config"
)

// SetIdentitySchemas sets the identity schemas in viper config:
//
//	testhelpers.SetIdentitySchemas(map[string]string{"customer": "file://customer.json"})
func SetIdentitySchemas(t *testing.T, conf *config.Config, schemas map[string]string) {
	var s []config.Schema
	for id, location := range schemas {
		s = append(s, config.Schema{ID: id, URL: location})
	}

	conf.MustSet(config.ViperKeyIdentitySchemas, s)
}
