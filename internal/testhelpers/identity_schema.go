// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package testhelpers

import (
	"github.com/ory/kratos/driver/config"
)

func IdentitySchemasConfig(schemas map[string]string) map[string]any {
	var s []config.Schema
	for id, location := range schemas {
		s = append(s, config.Schema{ID: id, URL: location})
	}
	return map[string]any{config.ViperKeyIdentitySchemas: s}
}
