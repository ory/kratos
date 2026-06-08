// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package driver

import (
	"context"

	"github.com/ory/kratos/schema"
)

// IdentityTraitsSchemas returns the identity traits schemas. The provider is
// eagerly initialized in initCheapMembers (or replaced in Init), so this getter
// only reads the field and never lazily initializes it.
func (m *RegistryDefault) IdentityTraitsSchemas(ctx context.Context) (schema.IdentitySchemaList, error) {
	return m.identitySchemaProvider.IdentityTraitsSchemas(ctx)
}
