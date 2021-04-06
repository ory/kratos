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
					require.NoError(t, hash.CompareArgon2id(context.Background(), pw, hs))

					mod := make([]byte, len(pw))
					copy(mod, pw)
					mod[len(pw)-1] = ^pw[len(pw)-1]
					require.Error(t, hash.CompareArgon2id(context.Background(), mod, hs))
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

func TestComparatorBcryptFailsWhenPasswordIsTooLong(t *testing.T) {
	password := mkpw(t, 73)
	err := hash.CompareBcrypt(context.Background(), password, []byte("hash"))

	assert.Error(t, err, "password is too long")
}

func TestComparatorBcryptSuccess(t *testing.T) {
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

			err = hash.CompareBcrypt(context.Background(), pw, hs)
			assert.Nil(t, err, "hash validation fails")
		})
	}
}

func TestComparatorBcryptFail(t *testing.T) {
	for k, pw := range [][]byte{
		mkpw(t, 8),
		mkpw(t, 16),
		mkpw(t, 32),
		mkpw(t, 64),
		mkpw(t, 72),
	} {
		t.Run(fmt.Sprintf("case=%d", k), func(t *testing.T) {
			mod := make([]byte, len(pw))
			copy(mod, pw)
			mod[len(pw)-1] = ^pw[len(pw)-1]

			err := hash.CompareBcrypt(context.Background(), pw, mod)
			assert.Error(t, err)
		})
	}
}

func TestCompare(t *testing.T) {
	assert.Nil(t, hash.Compare(context.Background(), []byte("test"), []byte("$2a$12$o6hx.Wog/wvFSkT/Bp/6DOxCtLRTDj7lm9on9suF/WaCGNVHbkfL6")))
	assert.Nil(t, hash.CompareBcrypt(context.Background(), []byte("test"), []byte("$2a$12$o6hx.Wog/wvFSkT/Bp/6DOxCtLRTDj7lm9on9suF/WaCGNVHbkfL6")))
	assert.Error(t, hash.Compare(context.Background(), []byte("test"), []byte("$2a$12$o6hx.Wog/wvFSkT/Bp/6DOxCtLRTDj7lm9on9suF/WaCGNVHbkfL7")))

	assert.Nil(t, hash.Compare(context.Background(), []byte("test"), []byte("$2a$15$GRvRO2nrpYTEuPQX6AieaOlZ4.7nMGsXpt.QWMev1zrP86JNspZbO")))
	assert.Nil(t, hash.CompareBcrypt(context.Background(), []byte("test"), []byte("$2a$15$GRvRO2nrpYTEuPQX6AieaOlZ4.7nMGsXpt.QWMev1zrP86JNspZbO")))
	assert.Error(t, hash.Compare(context.Background(), []byte("test"), []byte("$2a$15$GRvRO2nrpYTEuPQX6AieaOlZ4.7nMGsXpt.QWMev1zrP86JNspZb1")))

	assert.Nil(t, hash.Compare(context.Background(), []byte("test"), []byte("$argon2id$v=19$m=32,t=2,p=4$cm94YnRVOW5jZzFzcVE4bQ$MNzk5BtR2vUhrp6qQEjRNw")))
	assert.Nil(t, hash.CompareArgon2id(context.Background(), []byte("test"), []byte("$argon2id$v=19$m=32,t=2,p=4$cm94YnRVOW5jZzFzcVE4bQ$MNzk5BtR2vUhrp6qQEjRNw")))
	assert.Error(t, hash.Compare(context.Background(), []byte("test"), []byte("$argon2id$v=19$m=32,t=2,p=4$cm94YnRVOW5jZzFzcVE4bQ$MNzk5BtR2vUhrp6qQEjRN2")))

	assert.Nil(t, hash.Compare(context.Background(), []byte("test"), []byte("$argon2id$v=19$m=32,t=5,p=4$cm94YnRVOW5jZzFzcVE4bQ$fBxypOL0nP/zdPE71JtAV71i487LbX3fJI5PoTN6Lp4")))
	assert.Nil(t, hash.CompareArgon2id(context.Background(), []byte("test"), []byte("$argon2id$v=19$m=32,t=5,p=4$cm94YnRVOW5jZzFzcVE4bQ$fBxypOL0nP/zdPE71JtAV71i487LbX3fJI5PoTN6Lp4")))
	assert.Error(t, hash.Compare(context.Background(), []byte("test"), []byte("$argon2id$v=19$m=32,t=5,p=4$cm94YnRVOW5jZzFzcVE4bQ$fBxypOL0nP/zdPE71JtAV71i487LbX3fJI5PoTN6Lp5")))
}
