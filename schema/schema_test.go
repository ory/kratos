package schema

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/driver/configuration"
	"github.com/ory/x/urlx"
)

func TestSchemas_GetByID(t *testing.T) {
	urlFromID := func(id string) string {
		return fmt.Sprintf("http://%s.com", id)
	}

	ss := Schemas{
		Schema{
			ID: "foo",
		},
		Schema{
			ID: "bar",
		},
		Schema{
			ID: "foobar",
		},
		Schema{
			ID: configuration.DefaultIdentityTraitsSchemaID,
		},
	}

	for _, s := range ss {
		s.RawURL = urlFromID(s.ID)
		s.URL = urlx.ParseOrPanic(s.RawURL)
	}

	t.Run("case=get first schema", func(t *testing.T) {
		s, err := ss.GetByID("foo")
		require.NoError(t, err)
		assert.Equal(t, &ss[0], s)
	})

	t.Run("case=get second schema", func(t *testing.T) {
		s, err := ss.GetByID("bar")
		require.NoError(t, err)
		assert.Equal(t, &ss[1], s)
	})

	t.Run("case=get third schema", func(t *testing.T) {
		s, err := ss.GetByID("foobar")
		require.NoError(t, err)
		assert.Equal(t, &ss[2], s)
	})

	t.Run("case=get default schema", func(t *testing.T) {
		s1, err := ss.GetByID("")
		require.NoError(t, err)
		s2, err := ss.GetByID(configuration.DefaultIdentityTraitsSchemaID)
		require.NoError(t, err)
		assert.Equal(t, &ss[3], s1)
		assert.Equal(t, &ss[3], s2)
	})

	t.Run("case=should return error on not existing id", func(t *testing.T) {
		s, err := ss.GetByID("not existing id")
		require.Error(t, err)
		assert.Equal(t, (*Schema)(nil), s)
	})
}

func TestGetKeysInOrder(t *testing.T) {
	for i, tc := range []struct {
		schemaRef string
		keys      []string
		path      string
	}{
		{schemaRef: "file://./stub/identity.schema.json", keys: []string{"bar", "email"}},
		{schemaRef: "file://./stub/complex.schema.json", keys: []string{"meal.name", "meal.chef", "traits.email",
			"traits.stringy", "traits.numby", "traits.booly", "traits.should_big_number", "traits.should_long_string",
			"fruits", "vegetables"}},
	} {
		t.Run(fmt.Sprintf("case=%d schemaRef=%s", i, tc.schemaRef), func(t *testing.T) {
			actual, err := GetKeysInOrder(tc.schemaRef)
			require.NoError(t, err)
			assert.Equal(t, tc.keys, actual)
		})
	}
}
