// Copyright © 2025 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package test

import (
	"encoding/base64"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/schema"
	"github.com/ory/x/contextx"
	"github.com/ory/x/urlx"
)

func TestIdentitySchemaProvider(t *testing.T, provider schema.IdentitySchemaProvider) {
	urlFromID := func(id string) string {
		return fmt.Sprintf("http://%s.com", id)
	}

	schemas := schema.Schemas{
		{ID: "foo"},
		{ID: "preset://email"},
		{ID: config.DefaultIdentityTraitsSchemaID},
	}

	for i := range schemas {
		raw := urlFromID(schemas[i].ID)
		schemas[i].RawURL = raw
		schemas[i].URL = urlx.ParseOrPanic(raw)
	}

	ctx := contextx.WithConfigValues(t.Context(), map[string]any{
		config.ViperKeyIdentitySchemas: func() (cs config.Schemas) {
			for _, s := range schemas {
				cs = append(cs, config.Schema{
					ID:  s.ID,
					URL: s.RawURL,
				})
			}
			return cs
		}(),
	})

	list, err := provider.IdentityTraitsSchemas(ctx)
	require.NoError(t, err)

	t.Run("GetByID", func(t *testing.T) {
		t.Run("case=get with raw schemaID", func(t *testing.T) {
			for _, schema := range schemas {
				actual, err := list.GetByID(schema.ID)
				require.NoError(t, err)
				require.Equal(t, schema, *actual)
			}
		})

		t.Run("case=get with encoded schemaID", func(t *testing.T) {
			for _, schema := range schemas {
				encodedID := base64.RawURLEncoding.EncodeToString([]byte(schema.ID))

				actual, err := list.GetByID(encodedID)
				require.NoError(t, err)
				require.Equal(t, schema, *actual)
			}
		})

		t.Run("case=get default schema", func(t *testing.T) {
			s1, err := list.GetByID("")
			require.NoError(t, err)
			s2, err := list.GetByID(config.DefaultIdentityTraitsSchemaID)
			require.NoError(t, err)
			assert.Equal(t, &schemas[2], s1)
			assert.Equal(t, &schemas[2], s2)
		})

		t.Run("case=should return error on not existing id", func(t *testing.T) {
			s, err := list.GetByID("not existing id")
			require.Error(t, err)
			assert.Equal(t, (*schema.Schema)(nil), s)
		})
	})

	t.Run("List", func(t *testing.T) {
		t.Run("case=get all schemas", func(t *testing.T) {
			p0 := list.List(0, 4)
			assert.Equal(t, schemas, p0)
		})

		t.Run("case=smaller pages", func(t *testing.T) {
			p0, p1 := list.List(0, 2), list.List(1, 2)
			assert.Equal(t, schemas, append(p0, p1...))
		})

		t.Run("case=indexes out of range", func(t *testing.T) {
			p0 := list.List(-1, 10)
			assert.Equal(t, schemas, p0)
		})
	})
}
