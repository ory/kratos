package hash_test

import (
	"context"
	"crypto/rand"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/hash"
	"github.com/ory/kratos/internal"
)

func mkpw(t *testing.T, length int) []byte {
	pw := make([]byte, length)
	_, err := rand.Read(pw)
	require.NoError(t, err)
	return pw
}

func TestArgonHasher(t *testing.T) {
	for k, pw := range [][]byte{
		mkpw(t, 8),
		mkpw(t, 16),
		mkpw(t, 32),
		mkpw(t, 64),
		mkpw(t, 128),
	} {
		t.Run(fmt.Sprintf("case=%d", k), func(t *testing.T) {
			_, reg := internal.NewFastRegistryWithMocks(t)
			for kk, h := range []hash.Hasher{
				hash.NewHasherArgon2(reg),
			} {
				t.Run(fmt.Sprintf("hasher=%T/password=%d", h, kk), func(t *testing.T) {
					hs, err := h.Generate(context.Background(), pw)
					require.NoError(t, err)
					assert.NotEqual(t, pw, hs)

					t.Logf("hash: %s", hs)
					require.NoError(t, h.Compare(context.Background(), pw, hs))

					mod := make([]byte, len(pw))
					copy(mod, pw)
					mod[len(pw)-1] = ^pw[len(pw)-1]
					require.Error(t, h.Compare(context.Background(), mod, hs))
				})
			}
		})
	}
}

func TestBcryptHasherGeneratesErrorWhenPasswordIsLong(t *testing.T) {
	_, reg := internal.NewFastRegistryWithMocks(t)
	hasher := hash.NewHasherBcrypt(reg)

	password := mkpw(t, 73)
	res, err := hasher.Generate(context.Background(), password)

	assert.Error(t, err, "password is too long")
	assert.Nil(t, res)
}

func TestBcryptHasherGeneratesHash(t *testing.T) {
	for k, pw := range [][]byte{
		mkpw(t, 8),
		mkpw(t, 16),
		mkpw(t, 32),
		mkpw(t, 64),
		mkpw(t, 72),
	} {
		t.Run(fmt.Sprintf("case=%d", k), func(t *testing.T) {
			_, reg := internal.NewFastRegistryWithMocks(t)
			hasher := hash.NewHasherBcrypt(reg)
			res, err := hasher.Generate(context.Background(), pw)

			assert.Nil(t, err)

			// Valid format: $2a$12$[22 character salt][31 character hash]
			assert.Equal(t, 60, len(string(res)), "invalid bcrypt hash length")
			assert.Equal(t, "$2a$12$", string(res)[:7], "invalid bcrypt identifier")
		})
	}
}


func TestBcryptHasherComparesFailsWhenPasswordIsTooLong(t *testing.T) {
	_, reg := internal.NewFastRegistryWithMocks(t)
	hasher := hash.NewHasherBcrypt(reg)

	password := mkpw(t, 73)
	err := hasher.Compare(context.Background(), password, []byte("hash"))

	assert.Error(t, err, "password is too long")
}


func TestBcryptHasherComparesHashSuccess(t *testing.T) {
	for k, pw := range [][]byte{
		mkpw(t, 8),
		mkpw(t, 16),
		mkpw(t, 32),
		mkpw(t, 64),
		mkpw(t, 72),
	} {
		t.Run(fmt.Sprintf("case=%d", k), func(t *testing.T) {
			_, reg := internal.NewFastRegistryWithMocks(t)
			hasher := hash.NewHasherBcrypt(reg)

			hs, err := hasher.Generate(context.Background(), pw)

			assert.Nil(t, err)

			err = hasher.Compare(context.Background(), pw, hs)
			assert.Nil(t, err, "hash validation fails")
		})
	}
}

func TestBcryptHasherComparesHashFail(t *testing.T) {
	for k, pw := range [][]byte{
		mkpw(t, 8),
		mkpw(t, 16),
		mkpw(t, 32),
		mkpw(t, 64),
		mkpw(t, 72),
	} {
		t.Run(fmt.Sprintf("case=%d", k), func(t *testing.T) {
			_, reg := internal.NewFastRegistryWithMocks(t)
			hasher := hash.NewHasherBcrypt(reg)

			mod := make([]byte, len(pw))
			copy(mod, pw)
			mod[len(pw)-1] = ^pw[len(pw)-1]

			err := hasher.Compare(context.Background(), pw, mod)
			assert.Error(t, err)
		})
	}
}
