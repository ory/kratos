// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package driver

import (
	"context"

	"github.com/ory/kratos/schema"
)

func (m *RegistryDefault) IdentityTraitsSchemas(ctx context.Context) (schema.IdentitySchemaList, error) {
	if m.identitySchemaProvider == nil {
		m.identitySchemaProvider = schema.NewDefaultIdentityTraitsProvider(m)
	}
	return m.identitySchemaProvider.IdentityTraitsSchemas(ctx)
}
