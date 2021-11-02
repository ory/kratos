package password_test

import (
	"context"
	"fmt"
	"testing"

	hash2 "github.com/ory/kratos/hash"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/selfservice/strategy/password"
)

func TestCountActiveCredentials(t *testing.T) {
	_, reg := internal.NewFastRegistryWithMocks(t)
	strategy := password.NewStrategy(reg)

	h1, err := hash2.NewHasherBcrypt(reg).Generate(context.Background(), []byte("a password"))
	require.NoError(t, err)
	h2, err := reg.Hasher().Generate(context.Background(), []byte("a password"))
	require.NoError(t, err)

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
				Config: []byte(`{"hashed_password": "` + string(h1) + `"}`),
			}},
			expected: 0,
		},
		{
			in: identity.CredentialsCollection{{
				Type:        strategy.ID(),
				Identifiers: []string{""},
				Config:      []byte(`{"hashed_password": "` + string(h1) + `"}`),
			}},
			expected: 0,
		},
		{
			in: identity.CredentialsCollection{{
				Type:        strategy.ID(),
				Identifiers: []string{"foo"},
				Config:      []byte(`{"hashed_password": "` + string(h1) + `"}`),
			}},
			expected: 1,
		},
		{
			in: identity.CredentialsCollection{{
				Type:        strategy.ID(),
				Identifiers: []string{"foo"},
				Config:      []byte(`{"hashed_password": "` + string(h2) + `"}`),
			}},
			expected: 1,
		},
		{
			in: identity.CredentialsCollection{{
				Type:   strategy.ID(),
				Config: []byte(`{"hashed_password": "asdf"}`),
			}},
			expected: 0,
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

			actual, err := strategy.CountActiveCredentials(cc)
			require.NoError(t, err)
			assert.Equal(t, tc.expected, actual)
		})
	}
}
