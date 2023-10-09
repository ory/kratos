// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package testhelpers

import (
	"context"
	"testing"

	"github.com/ory/kratos/driver/config"
)

// SetIdentitySchemas sets the identity schemas in viper config:
//
//	testhelpers.SetIdentitySchemas(map[string]string{"customer": "file://customer.json"})
func SetIdentitySchemas(t *testing.T, conf *config.Config, schemas map[string]string) {
	ctx := context.Background()
	var s []config.Schema
	for id, location := range schemas {
		s = append(s, config.Schema{ID: id, URL: location})
	}

	conf.MustSet(ctx, config.ViperKeyIdentitySchemas, s)
}
