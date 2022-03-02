package lookup_test

import (
	"fmt"
	"testing"

	"github.com/ory/kratos/selfservice/strategy/lookup"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/internal"
)

func TestCountActiveFirstFactorCredentials(t *testing.T) {
	_, reg := internal.NewFastRegistryWithMocks(t)
	strategy := lookup.NewStrategy(reg)

	t.Run("first factor", func(t *testing.T) {
		actual, err := strategy.CountActiveFirstFactorCredentials(nil)
		require.NoError(t, err)
		assert.Equal(t, 0, actual)
	})

	t.Run("multi factor", func(t *testing.T) {
		for k, tc := range []struct {
			in       identity.CredentialsCollection
			expected int
		}{
			{
				in: identity.CredentialsCollection{{
					Type:   strategy.ID(),
					Config: []byte{},
				}},
				expected: 0,
			},
			{
				in: identity.CredentialsCollection{{
					Type:   strategy.ID(),
					Config: []byte(`{"recovery_codes": []}`),
				}},
				expected: 0,
			},
			{
				in: identity.CredentialsCollection{{
					Type:        strategy.ID(),
					Identifiers: []string{"foo"},
					Config:      []byte(`{"recovery_codes": [{}]}`),
				}},
				expected: 1,
			},
			{
				in: identity.CredentialsCollection{{
					Type:   strategy.ID(),
					Config: []byte(`{}`),
				}},
				expected: 0,
			},
			{
				in:       identity.CredentialsCollection{{}, {}},
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
