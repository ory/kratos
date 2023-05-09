// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package totp_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/ory/kratos/selfservice/strategy/totp"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/internal"
)

func TestCountActiveCredentials(t *testing.T) {
	_, reg := internal.NewFastRegistryWithMocks(t)
	strategy := totp.NewStrategy(reg)

	key, err := totp.NewKey(context.Background(), "foo", reg)
	require.NoError(t, err)

	t.Run("first factor", func(t *testing.T) {
		actual, err := strategy.CountActiveFirstFactorCredentials(nil)
		require.NoError(t, err)
		assert.Equal(t, 0, actual)
	})

	t.Run("multi factor", func(t *testing.T) {
		for k, tc := range []struct {
			in       map[identity.CredentialsType]identity.Credentials
			expected int
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
					Config: []byte(`{"totp_url": ""}`),
				}},
				expected: 0,
			},
			{
				in: map[identity.CredentialsType]identity.Credentials{strategy.ID(): {
					Type:        strategy.ID(),
					Identifiers: []string{"foo"},
					Config:      []byte(`{"totp_url": "` + key.URL() + `"}`),
				}},
				expected: 1,
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
				cc := map[identity.CredentialsType]identity.Credentials{}
				for _, c := range tc.in {
					cc[c.Type] = c
				}

				actual, err := strategy.CountActiveMultiFactorCredentials(cc)
				require.NoError(t, err)
				assert.Equal(t, tc.expected, actual)
			})
		}
	})
}
