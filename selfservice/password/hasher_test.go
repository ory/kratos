package password_test

import (
	"crypto/rand"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/hive/internal"
	. "github.com/ory/hive/selfservice/password"
)

func password(t *testing.T, length int) []byte {
	pw := make([]byte, length)
	_, err := rand.Read(pw)
	require.NoError(t, err)
	return pw
}

func TestHasher(t *testing.T) {
	for k, pw := range [][]byte{
		password(t, 8),
		password(t, 16),
		password(t, 32),
		password(t, 64),
		password(t, 128),
	} {
		t.Run(fmt.Sprintf("case=%d", k), func(t *testing.T) {
			conf := internal.NewConfigurationWithDefaults()
			for kk, h := range []Hasher{
				NewHasherArgon2(conf),
			} {
				t.Run(fmt.Sprintf("hasher=%T/password=%d", h, kk), func(t *testing.T) {
					hs, err := h.Generate(pw)
					require.NoError(t, err)
					assert.NotEqual(t, pw, hs)

					require.NoError(t, h.Compare(pw, hs))

					mod := make([]byte, len(pw))
					copy(mod, pw)
					mod[len(pw)-1] = ^pw[len(pw)-1]
					require.Error(t, h.Compare(mod, hs))
				})
			}
		})
	}
}
