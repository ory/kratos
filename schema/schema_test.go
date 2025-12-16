// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package schema_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/schema"
	schematest "github.com/ory/kratos/schema/test"
)

func TestDefaultIdentityTraitsProvider(t *testing.T) {
	_, reg := internal.NewFastRegistryWithMocks(t)
	schematest.TestIdentitySchemaProvider(t, schema.NewDefaultIdentityTraitsProvider(reg))
}

func TestGetKeysInOrder(t *testing.T) {
	for i, tc := range []struct {
		schemaRef string
		keys      []string
		path      string
	}{
		{schemaRef: "file://./stub/identity.schema.json", keys: []string{"bar", "email"}},
		{schemaRef: "file://./stub/complex.schema.json", keys: []string{
			"meal.name", "meal.chef", "traits.email",
			"traits.stringy", "traits.numby", "traits.booly", "traits.should_big_number", "traits.should_long_string",
			"fruits", "vegetables",
		}},
	} {
		t.Run(fmt.Sprintf("case=%d schemaRef=%s", i, tc.schemaRef), func(t *testing.T) {
			actual, err := schema.GetKeysInOrder(context.Background(), tc.schemaRef)
			require.NoError(t, err)
			assert.Equal(t, tc.keys, actual)
		})
	}
}
