package identity_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/ory/herodot"

	"github.com/ory/hive/selfservice/password"

	"github.com/ory/hive/driver"

	"github.com/bxcodec/faker"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	. "github.com/ory/hive/identity"
	"github.com/ory/hive/internal"
)

func init() {
	internal.RegisterFakes()
}

func TestPool(t *testing.T) {
	conf := internal.NewConfigurationWithDefaults()
	r := new(driver.RegistryMemory).WithConfig(conf)

	pools := map[string]Pool{
		"memory": NewPoolMemory(r),
	}

	var identities []Identity
	for k := 0; k < 5; k++ {
		var i Identity
		require.NoError(t, faker.FakeData(&i))
		i.Credentials = map[CredentialsType]Credentials{
			password.CredentialsType: {
				ID:          password.CredentialsType,
				Identifiers: []string{fmt.Sprintf("id-%d", k)},
				Options:     json.RawMessage(`{}`),
			},
		}
		identities = append(identities, i)
	}

	ii := Identity{ID: identities[0].ID}
	ii.Credentials = map[CredentialsType]Credentials{
		password.CredentialsType: {
			ID:          password.CredentialsType,
			Identifiers: []string{fmt.Sprintf("id-1")},
			Options:     json.RawMessage(`{}`),
		},
	}

	ctx := context.Background()
	for name, p := range pools {
		t.Run("adapter="+name, func(t *testing.T) {
			t.Run("case=create", func(t *testing.T) {
				_, err := p.Get(ctx, "does-not-exist")
				require.Error(t, err)

				_, err = p.Get(ctx, identities[0].ID)
				require.Error(t, err)
			})

			t.Run("case=create", func(t *testing.T) {
				i := identities[0]
				_, err := p.Create(ctx, &i)
				require.NoError(t, err)

				g, err := p.Get(ctx, identities[0].ID)
				require.NoError(t, err)
				require.EqualValues(t, g, &i)

				g, err = p.Create(ctx, &identities[1])
				require.NoError(t, err)
				require.EqualValues(t, g, &identities[1])

				// violates uniqueness
				_, err = p.Create(ctx, &i)
				require.EqualError(t, err, herodot.ErrConflict.Error())
			})

			t.Run("case=update", func(t *testing.T) {
				i := identities[0]
				i.Traits = json.RawMessage(`["a"]`)
				_, err := p.Update(ctx, &i)
				require.NoError(t, err)

				g, err := p.Get(ctx, identities[0].ID)
				require.NoError(t, err)
				require.EqualValues(t, g, &i)

				i2 := identities[2]
				_, err = p.Update(ctx, &i2)
				require.Error(t, err)

				// violates uniqueness
				_, err = p.Update(ctx, &ii)
				require.EqualError(t, err, herodot.ErrConflict.Error())
			})

			t.Run("case=list", func(t *testing.T) {
				for limit := 1; limit <= 2; limit++ {
					t.Run(fmt.Sprintf("limit=%d", limit), func(t *testing.T) {
						is, err := p.List(ctx, limit, 0)
						require.NoError(t, err)
						assert.Len(t, is, limit)
						var prev Identity
						for _, i := range is {
							assert.NotEqual(t, prev, i)
							prev = i
						}
					})
				}
			})

			t.Run("case=delete", func(t *testing.T) {
				require.NoError(t, p.Delete(ctx, identities[0].ID))
				_, err := p.Get(ctx, identities[0].ID)
				require.Error(t, err)
			})
		})
	}
}
