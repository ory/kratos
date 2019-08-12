package identity_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/ory/herodot"
	"github.com/stretchr/testify/assert"

	"github.com/stretchr/testify/require"

	. "github.com/ory/hive/identity"
	"github.com/ory/hive/internal"
)

func init() {
	internal.RegisterFakes()
}

func TestPool(t *testing.T) {
	conf, reg := internal.NewMemoryRegistry(t)
	pools := map[string]Pool{
		"memory": NewPoolMemory(conf, reg),
	}

	var newid = func(schemaURL string, credentialsID string) *Identity {
		i := NewIdentity(schemaURL)
		i.SetCredentials(CredentialsTypePassword, Credentials{
			ID: CredentialsTypePassword, Identifiers:
			[]string{credentialsID},
			Options: json.RawMessage(`{}`),
		})
		return i
	}

	var assertequal = func(t *testing.T, expected, actual *Identity) {
		assert.Empty(t, actual.Credentials)
		require.Equal(t, expected.Traits, actual.Traits)
		require.Equal(t, expected.ID, actual.ID)
	}

	for name, pool := range pools {
		t.Run("dbal="+name, func(t *testing.T) {
			t.Run("case=get-not-exist", func(t *testing.T) {
				_, err := pool.Get(context.Background(), "does-not-exist")
				require.EqualError(t, err, herodot.ErrNotFound.Error())
			})
		})

		t.Run("case=create with default values", func(t *testing.T) {
			i := newid("", "id-1")
			i.ID = "id-1"

			got, err := pool.Create(context.Background(), i)
			require.NoError(t, err)
			require.Empty(t, got.Credentials)

			got, err = pool.Get(context.Background(), i.ID)
			require.NoError(t, err)

			assert.Equal(t, "file://./stub/identity.schema.json", got.TraitsSchemaURL)
			assertequal(t, i, got)
		})

		t.Run("case=create and keep set values", func(t *testing.T) {
			i := newid("file://./stub/identity-2.schema.json", "id-2")
			i.ID = "id-2"

			_, err := pool.Create(context.Background(), i)
			require.NoError(t, err)

			got, err := pool.Get(context.Background(), i.ID)
			require.NoError(t, err)
			assert.Equal(t, "file://./stub/identity-2.schema.json", got.TraitsSchemaURL)
			assertequal(t, i, got)
		})

		t.Run("case=fail on duplicate credential identifiers", func(t *testing.T) {
			i := newid("", "id-1")

			_, err := pool.Create(context.Background(), i)
			require.Error(t, err)
		})

		t.Run("case=create with default values", func(t *testing.T) {
			i := newid("", "id-3")
			i.Traits = json.RawMessage(`{"bar":123}`)

			_, err := pool.Create(context.Background(), i)
			require.Error(t, err)
		})

		t.Run("case=update an identity", func(t *testing.T) {
			got, err := pool.Get(context.Background(), "id-1")
			require.NoError(t, err)

			got.TraitsSchemaURL = "file://./stub/identity-2.schema.json"
			got, err = pool.Update(context.Background(), got)
			require.NoError(t, err)
			require.Empty(t, got.Credentials)

			got, err = pool.Get(context.Background(), "id-1")
			require.NoError(t, err)
			assert.Equal(t, "file://./stub/identity-2.schema.json", got.TraitsSchemaURL)
		})

		t.Run("case=fail to update because validation fails", func(t *testing.T) {
			got, err := pool.Get(context.Background(), "id-1")
			require.NoError(t, err)

			got.Traits = json.RawMessage(`{"bar":123}`)
			_, err = pool.Update(context.Background(), got)
			require.Error(t, err)
		})

		t.Run("case=updating credentials should work", func(t *testing.T) {
			toUpdate, err := pool.Get(context.Background(), "id-1")
			require.NoError(t, err)

			toUpdate.Credentials = map[CredentialsType]Credentials{
				CredentialsTypeOIDC: {
					ID: CredentialsTypeOIDC, Identifiers:
					[]string{"id-2"},
					Options: json.RawMessage(`{}`),
				},
			}

			_, err = pool.Update(context.Background(), toUpdate)
			require.NoError(t, err)

			got, err := pool.GetClassified(context.Background(), toUpdate.ID)
			require.NoError(t, err)

			assert.Equal(t, []string{"id-2"}, got.Credentials[CredentialsTypeOIDC].Identifiers)
			assert.Empty(t, got.Credentials[CredentialsTypePassword])
		})

		t.Run("case=create and update an identity with credentials from traits", func(t *testing.T) {
			i := newid("file://./stub/identity.schema.json", "id-3")
			i.Traits = json.RawMessage(`{"email":"email-id-3"}`)

			_, err :=  pool.Create(context.Background(), i)
			require.NoError(t,err)

			got, err := pool.GetClassified(context.Background(), i.ID)
			require.NoError(t, err)
			assert.Equal(t, []string{"email-id-3"}, got.Credentials[CredentialsTypePassword].Identifiers)

			i.Traits = json.RawMessage(`{"email":"email-id-4"}`)
			_, err =  pool.Update(context.Background(), i)
			require.NoError(t,err)

			got, err = pool.GetClassified(context.Background(), i.ID)
			require.NoError(t, err)
			assert.Equal(t, []string{"email-id-4"}, got.Credentials[CredentialsTypePassword].Identifiers)
		})

		t.Run("case=list", func(t *testing.T) {
			is, err := pool.List(context.Background(),10,0)
			require.NoError(t, err)
			assert.Equal(t, "id-1", is[0].ID)
			assert.Equal(t, "id-2", is[1].ID)
		})

		t.Run("case=delete an identity", func(t *testing.T) {
			err := pool.Delete(context.Background(), "id-1")
			require.NoError(t, err)

			_, err = pool.GetClassified(context.Background(), "id-1")
			require.Error(t, err)
		})
	}
}
