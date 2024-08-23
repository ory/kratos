// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package password_test

import (
	"cmp"
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/ory/kratos/driver/config"
	confighelpers "github.com/ory/kratos/driver/config/testhelpers"

	hash2 "github.com/ory/kratos/hash"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/go-faker/faker/v4"

	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/selfservice/strategy/password"
)

func generateRandomConfig(t *testing.T) (identity.CredentialsPassword, []byte) {
	t.Helper()
	var cred identity.CredentialsPassword
	require.NoError(t, faker.FakeData(&cred))
	c, err := json.Marshal(cred)
	require.NoError(t, err)
	return cred, c
}

func TestCountActiveFirstFactorCredentials(t *testing.T) {
	ctx := context.Background()
	_, reg := internal.NewFastRegistryWithMocks(t)
	strategy := password.NewStrategy(reg)

	h1, err := hash2.NewHasherBcrypt(reg).Generate(context.Background(), []byte("a password"))
	require.NoError(t, err)
	h2, err := reg.Hasher(ctx).Generate(context.Background(), []byte("a password"))
	require.NoError(t, err)

	t.Run("test regressions fixtures", func(t *testing.T) {
		// This test ensures we do not add regressions to this method by, for example, adding a new field.
		for k := 0; k < 100; k++ {
			t.Run(fmt.Sprintf("run=%d", k), func(t *testing.T) {
				cred, c := generateRandomConfig(t)
				actual, err := strategy.CountActiveFirstFactorCredentials(ctx, map[identity.CredentialsType]identity.Credentials{strategy.ID(): {
					Type:        strategy.ID(),
					Identifiers: []string{"foo"},
					Config:      c,
				}})
				assert.NoError(t, err)

				if len(cred.HashedPassword) == 0 && cred.UsePasswordMigrationHook {
					// This case is OK
					assert.Equal(t, 0, actual)
					return
				}

				assert.Equal(t, 1, actual)
			})
		}
	})

	t.Run("with fixtures", func(t *testing.T) {
		for k, tc := range []struct {
			in       map[identity.CredentialsType]identity.Credentials
			expected int
			ctx      context.Context
		}{
			{
				in: map[identity.CredentialsType]identity.Credentials{strategy.ID(): {
					Type:   strategy.ID(),
					Config: []byte{},
				}},
				expected: 0,
			},
			{
				in: map[identity.CredentialsType]identity.Credentials{strategy.ID(): {
					Type:   strategy.ID(),
					Config: []byte(`{"hashed_password": "` + string(h1) + `"}`),
				}},
				expected: 0,
			},
			{
				in: map[identity.CredentialsType]identity.Credentials{strategy.ID(): {
					Type:        strategy.ID(),
					Identifiers: []string{""},
					Config:      []byte(`{"hashed_password": "` + string(h1) + `"}`),
				}},
				expected: 0,
			},
			{
				in: map[identity.CredentialsType]identity.Credentials{strategy.ID(): {
					Type:        strategy.ID(),
					Identifiers: []string{"foo"},
					Config:      []byte(`{"hashed_password": "` + string(h1) + `"}`),
				}},
				expected: 1,
			},
			{
				in: map[identity.CredentialsType]identity.Credentials{strategy.ID(): {
					Type:        strategy.ID(),
					Identifiers: []string{"foo"},
					Config:      []byte(`{"hashed_password": "` + string(h2) + `"}`),
				}},
				expected: 1,
			},
			{
				in: map[identity.CredentialsType]identity.Credentials{strategy.ID(): {
					Type:        strategy.ID(),
					Identifiers: []string{"foo"},
					Config:      []byte(`{"use_password_migration_hook":true}`),
				}},
				expected: 1,
				ctx:      confighelpers.WithConfigValue(ctx, config.ViperKeyPasswordMigrationHook+".enabled", true),
			},
			{
				in: map[identity.CredentialsType]identity.Credentials{strategy.ID(): {
					Type:        strategy.ID(),
					Identifiers: []string{"foo"},
					Config:      []byte(`{"use_password_migration_hook":true}`),
				}},
				expected: 0,
				ctx:      confighelpers.WithConfigValue(ctx, config.ViperKeyPasswordMigrationHook+".enabled", false),
			},
			{
				in: map[identity.CredentialsType]identity.Credentials{strategy.ID(): {
					Type:        strategy.ID(),
					Identifiers: []string{"foo"},
					Config:      []byte(`{"use_password_migration_hook":false}`),
				}},
				expected: 0,
			},
			{
				in: map[identity.CredentialsType]identity.Credentials{strategy.ID(): {
					Type:   strategy.ID(),
					Config: []byte(`{"hashed_password": "asdf"}`),
				}},
				expected: 0,
			},
			{
				in: map[identity.CredentialsType]identity.Credentials{strategy.ID(): {
					Type:   strategy.ID(),
					Config: []byte(`{}`),
				}},
				expected: 0,
			},
			{
				in:       nil,
				expected: 0,
			},
		} {
			t.Run(fmt.Sprintf("case=%d", k), func(t *testing.T) {
				actual, err := strategy.CountActiveFirstFactorCredentials(cmp.Or(tc.ctx, ctx), tc.in)
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, actual)

				actual, err = strategy.CountActiveMultiFactorCredentials(cmp.Or(tc.ctx, ctx), tc.in)
				assert.NoError(t, err)
				assert.Equal(t, 0, actual)
			})
		}
	})
}
