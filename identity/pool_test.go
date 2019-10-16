package identity_test

import (
	"context"
	"encoding/json"
	"flag"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ory/x/sqlcon/dockertest"

	"github.com/ory/herodot"

	"github.com/stretchr/testify/require"

	. "github.com/ory/hive/identity"
	"github.com/ory/hive/internal"
)

func init() {
	internal.RegisterFakes()
}

// nolint: staticcheck
func TestMain(m *testing.M) {
	flag.Parse()
	runner := dockertest.Register()
	runner.Exit(m.Run())
}

func TestPool(t *testing.T) {
	conf, reg := internal.NewMemoryRegistry(t)
	pools := map[string]Pool{
		"memory": NewPoolMemory(conf, reg),
	}

	if !testing.Short() {
		var l sync.Mutex
		dockertest.Parallel([]func(){
			func() {
				db, err := dockertest.ConnectToTestPostgreSQL()
				require.NoError(t, err)

				_, reg := internal.NewRegistrySQL(t, db)

				l.Lock()
				pools["postgres"] = reg.IdentityPool()
				l.Unlock()
			},
		})
	}

	var newid = func(schemaURL string, credentialsID string) *Identity {
		i := NewIdentity(schemaURL)
		i.SetCredentials(CredentialsTypePassword, Credentials{
			ID: CredentialsTypePassword, Identifiers: []string{credentialsID},
			Config: json.RawMessage(`{}`),
		})
		return i
	}

	var assertEqual = func(t *testing.T, expected, actual *Identity) {
		assert.Empty(t, actual.Credentials)
		require.Equal(t, expected.Traits, actual.Traits)
		require.Equal(t, expected.ID, actual.ID)
	}

	for name, pool := range pools {
		t.Run("dbal="+name, func(t *testing.T) {
			t.Run("case=get-not-exist", func(t *testing.T) {
				_, err := pool.Get(context.Background(), "does-not-exist")
				require.EqualError(t, err, herodot.ErrNotFound.Error(), "%+v", err)
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
				assertEqual(t, i, got)
			})

			t.Run("case=create and keep set values", func(t *testing.T) {
				i := newid("file://./stub/identity-2.schema.json", "id-2")
				i.ID = "id-2"

				_, err := pool.Create(context.Background(), i)
				require.NoError(t, err)

				got, err := pool.Get(context.Background(), i.ID)
				require.NoError(t, err)
				assert.Equal(t, "file://./stub/identity-2.schema.json", got.TraitsSchemaURL)
				assertEqual(t, i, got)
			})

			t.Run("case=fail on duplicate credential identifiers", func(t *testing.T) {
				i := newid("", "id-1")

				_, err := pool.Create(context.Background(), i)
				require.Error(t, err)

				_, err = pool.Get(context.Background(), i.ID)
				require.Error(t, err)
			})

			t.Run("case=create with invalid traits data", func(t *testing.T) {
				i := newid("", "id-3")
				i.Traits = json.RawMessage(`{"bar":123}`) // bar should be a string

				_, err := pool.Create(context.Background(), i)
				require.Error(t, err)
			})

			t.Run("case=update an identity", func(t *testing.T) {
				toUpdate, err := pool.GetClassified(context.Background(), "id-1")
				require.NoError(t, err)
				require.NotEmpty(t, toUpdate.ID)
				require.NotEmpty(t, toUpdate.Credentials)

				toUpdate.TraitsSchemaURL = "file://./stub/identity-2.schema.json"
				toUpdate, err = pool.UpdateConfidential(context.Background(), toUpdate, toUpdate.Credentials)
				require.NoError(t, err)
				require.Empty(t, toUpdate.Credentials)

				updatedConfidential, err := pool.GetClassified(context.Background(), "id-1")
				require.NoError(t, err, "%+v", toUpdate)
				assert.Equal(t, "file://./stub/identity-2.schema.json", updatedConfidential.TraitsSchemaURL)
				assert.NotEmpty(t, updatedConfidential.Credentials)

				updatedConfidential.Traits = json.RawMessage(`{"bar":"bazbar"}`)
				toUpdate, err = pool.Update(context.Background(), toUpdate)
				require.NoError(t, err)

				updatedWithoutCredentials, err := pool.GetClassified(context.Background(), "id-1")
				require.NoError(t, err, "%+v", toUpdate)
				assert.Equal(t, "file://./stub/identity-2.schema.json", updatedConfidential.TraitsSchemaURL)
				assert.Equal(t, `{"bar":"bazbar"}`, string(updatedConfidential.Traits))
				assert.NotEmpty(t, updatedConfidential.Credentials)
				assert.Equal(t, updatedConfidential.Credentials, updatedWithoutCredentials.Credentials)
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
					CredentialsTypePassword: {
						ID: CredentialsTypePassword, Identifiers: []string{"new-id-1", "new-id-2"},
						Config: json.RawMessage(`{}`),
					},
				}

				_, err = pool.UpdateConfidential(context.Background(), toUpdate, toUpdate.Credentials)
				require.NoError(t, err)

				got, err := pool.GetClassified(context.Background(), toUpdate.ID)
				require.NoError(t, err)

				assert.Equal(t, []string{"new-id-1", "new-id-2"}, got.Credentials[CredentialsTypePassword].Identifiers)
			})

			t.Run("case=should fail to insert identity because credentials from traits exist", func(t *testing.T) {
				i := newid("file://./stub/identity.schema.json", "should-not-matter")
				i.Traits = json.RawMessage(`{"email":"id-2"}`)
				_, err := pool.Create(context.Background(), i)
				require.Error(t, err)

				i = newid("file://./stub/identity.schema.json", "id-4")
				_, err = pool.Create(context.Background(), i)
				require.NoError(t, err)

				i.Traits = json.RawMessage(`{"email":"id-2"}`)
				_, err = pool.UpdateConfidential(context.Background(), i, i.Credentials)
				require.Error(t, err)
			})

			t.Run("case=create and update an identity with credentials from traits", func(t *testing.T) {
				i := newid("file://./stub/identity.schema.json", "id-3")
				i.Traits = json.RawMessage(`{"email":"email-id-3"}`)

				_, err := pool.Create(context.Background(), i)
				require.NoError(t, err)

				got, err := pool.GetClassified(context.Background(), i.ID)
				require.NoError(t, err)
				assert.Equal(t, []string{"email-id-3"}, got.Credentials[CredentialsTypePassword].Identifiers)

				i.Traits = json.RawMessage(`{"email":"email-id-4"}`)
				_, err = pool.UpdateConfidential(context.Background(), i, i.Credentials)
				require.NoError(t, err)

				got, err = pool.GetClassified(context.Background(), i.ID)
				require.NoError(t, err)
				assert.Equal(t, []string{"email-id-4"}, got.Credentials[CredentialsTypePassword].Identifiers)
			})

			t.Run("case=list", func(t *testing.T) {
				is, err := pool.List(context.Background(), 10, 0)
				require.NoError(t, err)
				require.Len(t, is, 4)
				assert.Equal(t, "id-1", is[0].ID)
				assert.Equal(t, "id-2", is[1].ID)
			})

			t.Run("case=find identity by its credentials identifier", func(t *testing.T) {
				expected := newid("file://./stub/identity.schema.json", "id-5")
				expected.Traits = json.RawMessage(`{"email": "email-id-5"}`)
				ct := expected.Credentials[CredentialsTypePassword]
				ct.Identifiers = []string{"email-id-5"}
				expected.Credentials[CredentialsTypePassword] = ct

				_, err := pool.Create(context.Background(), expected)
				require.NoError(t, err)

				actual, creds, err := pool.FindByCredentialsIdentifier(context.Background(), CredentialsTypePassword, "email-id-5")
				require.NoError(t, err)

				assert.EqualValues(t, ct, *creds)

				expected.Credentials = nil
				assertEqual(t, expected, actual)
			})

			t.Run("case=delete an identity", func(t *testing.T) {
				err := pool.Delete(context.Background(), "id-1")
				require.NoError(t, err)

				_, err = pool.GetClassified(context.Background(), "id-1")
				require.Error(t, err)
			})

			t.Run("case=create with empty credentials config", func(t *testing.T) {
				// This test covers a case where the config value of a credentials setting is empty. This causes
				// issues with postgres' json field.
				i := newid("", "id-missing-creds-config")
				i.SetCredentials(CredentialsTypePassword, Credentials{
					ID: CredentialsTypePassword, Identifiers: []string{"id-missing-creds-config"},
					Config: json.RawMessage(``),
				})

				_, err := pool.Create(context.Background(), i)
				require.NoError(t, err)
			})
		})
	}
}
