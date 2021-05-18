package hash_test

import (
	"context"
	"crypto/rand"
	"fmt"
	"testing"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/x/configx"
	"github.com/ory/x/logrusx"
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

					algorithm, realHash, err := hash.ParsePasswordHash(hs)
					assert.NoError(t, err)
					assert.Equal(t, hash.Argon2AlgorithmId, algorithm)

					require.NoError(t, hash.CompareArgon2id(context.Background(), reg.Config(context.Background()), pw, realHash))

					mod := make([]byte, len(pw))
					copy(mod, pw)
					mod[len(pw)-1] = ^pw[len(pw)-1]
					require.Error(t, hash.CompareArgon2id(context.Background(), reg.Config(context.Background()), mod, realHash))
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

			algorithm, realHash, err := hash.ParsePasswordHash(res)
			assert.NoError(t, err)
			assert.Equal(t, hash.BcryptAlgorithmId, algorithm)

			// Valid format: $2a$12$[22 character salt][31 character hash]
			assert.Equal(t, 60, len(realHash), "invalid bcrypt hash length")
			assert.Equal(t, []byte("$2a$04$"), realHash[:7], "invalid bcrypt identifier")
		})
	}
}

func TestComparatorBcryptFailsWhenPasswordIsTooLong(t *testing.T) {
	_, reg := internal.NewFastRegistryWithMocks(t)
	password := mkpw(t, 73)
	err := hash.CompareBcrypt(context.Background(), reg.Config(context.Background()), password, []byte("hash"))

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

			algorithm, realHash, err := hash.ParsePasswordHash(hs)
			assert.NoError(t, err)
			assert.Equal(t, hash.BcryptAlgorithmId, algorithm)

			err = hash.CompareBcrypt(context.Background(), reg.Config(context.Background()), pw, realHash)
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
			_, reg := internal.NewFastRegistryWithMocks(t)
			mod := make([]byte, len(pw))
			copy(mod, pw)
			mod[len(pw)-1] = ^pw[len(pw)-1]

			err := hash.CompareBcrypt(context.Background(), reg.Config(context.Background()), pw, mod)
			assert.Error(t, err)
		})
	}
}

func TestComparatorBcryptAesSuccess(t *testing.T) {
	for k, pw := range [][]byte{
		mkpw(t, 8),
		mkpw(t, 16),
		mkpw(t, 32),
		mkpw(t, 64),
		mkpw(t, 72),
		mkpw(t, 96),
		mkpw(t, 128),
	} {
		t.Run(fmt.Sprintf("case=%d", k), func(t *testing.T) {
			_, reg := internal.NewFastRegistryWithMocks(t)
			hasher := hash.NewHasherBcryptAES(reg)

			hs, err := hasher.Generate(context.Background(), pw)

			assert.Nil(t, err)

			algorithm, realHash, err := hash.ParsePasswordHash(hs)
			assert.NoError(t, err)
			assert.Equal(t, hash.BcryptAESAlgorithmId, algorithm)

			err = hash.CompareBcryptAes(context.Background(), reg.Config(context.Background()), pw, realHash)
			assert.Nil(t, err, "hash validation fails")
		})
	}
}

func TestComparatorBcryptAesFail(t *testing.T) {
	p := config.MustNew(t, logrusx.New("", ""),
		configx.WithConfigFiles("../internal/.kratos.yaml"))

	for k, pw := range [][]byte{
		mkpw(t, 8),
		mkpw(t, 16),
		mkpw(t, 32),
		mkpw(t, 64),
		mkpw(t, 72),
		mkpw(t, 96),
		mkpw(t, 128),
	} {
		t.Run(fmt.Sprintf("case=%d", k), func(t *testing.T) {
			mod := make([]byte, len(pw))
			copy(mod, pw)
			mod[len(pw)-1] = ^mod[len(pw)-1]

			err := hash.CompareBcryptAes(context.Background(), p, pw, mod)
			assert.Error(t, err)
		})
	}
}

func TestCompare(t *testing.T) {
	p := config.MustNew(t, logrusx.New("", ""),
		configx.WithConfigFiles("../internal/.kratos.yaml"))

	assert.Nil(t, hash.Compare(context.Background(), p, []byte("test"), []byte("$bcrypt$2a$12$o6hx.Wog/wvFSkT/Bp/6DOxCtLRTDj7lm9on9suF/WaCGNVHbkfL6")))
	assert.Nil(t, hash.CompareBcrypt(context.Background(), p, []byte("test"), []byte("$2a$12$o6hx.Wog/wvFSkT/Bp/6DOxCtLRTDj7lm9on9suF/WaCGNVHbkfL6")))
	assert.Error(t, hash.Compare(context.Background(), p, []byte("test"), []byte("$bcrypt$2a$12$o6hx.Wog/wvFSkT/Bp/6DOxCtLRTDj7lm9on9suF/WaCGNVHbkfL7")))

	assert.Nil(t, hash.Compare(context.Background(), p, []byte("test"), []byte("$bcrypt$2a$15$GRvRO2nrpYTEuPQX6AieaOlZ4.7nMGsXpt.QWMev1zrP86JNspZbO")))
	assert.Nil(t, hash.CompareBcrypt(context.Background(), p, []byte("test"), []byte("$2a$15$GRvRO2nrpYTEuPQX6AieaOlZ4.7nMGsXpt.QWMev1zrP86JNspZbO")))
	assert.Error(t, hash.Compare(context.Background(), p, []byte("test"), []byte("$bcrypt$2a$15$GRvRO2nrpYTEuPQX6AieaOlZ4.7nMGsXpt.QWMev1zrP86JNspZb1")))

	// legacy bcrypt hash format
	assert.Nil(t, hash.Compare(context.Background(), p, []byte("test"), []byte("$2a$15$GRvRO2nrpYTEuPQX6AieaOlZ4.7nMGsXpt.QWMev1zrP86JNspZbO")))
	assert.Nil(t, hash.CompareBcrypt(context.Background(), p, []byte("test"), []byte("$2a$15$GRvRO2nrpYTEuPQX6AieaOlZ4.7nMGsXpt.QWMev1zrP86JNspZbO")))
	assert.Error(t, hash.Compare(context.Background(), p, []byte("test"), []byte("$2a$15$GRvRO2nrpYTEuPQX6AieaOlZ4.7nMGsXpt.QWMev1zrP86JNspZb1")))

	assert.Nil(t, hash.Compare(context.Background(), p, []byte("test"), []byte("$argon2id$v=19$m=32,t=2,p=4$cm94YnRVOW5jZzFzcVE4bQ$MNzk5BtR2vUhrp6qQEjRNw")))
	assert.Nil(t, hash.CompareArgon2id(context.Background(), p, []byte("test"), []byte("$v=19$m=32,t=2,p=4$cm94YnRVOW5jZzFzcVE4bQ$MNzk5BtR2vUhrp6qQEjRNw")))
	assert.Error(t, hash.Compare(context.Background(), p, []byte("test"), []byte("$argon2id$v=19$m=32,t=2,p=4$cm94YnRVOW5jZzFzcVE4bQ$MNzk5BtR2vUhrp6qQEjRN2")))

	assert.Nil(t, hash.Compare(context.Background(), p, []byte("test"), []byte("$argon2id$v=19$m=32,t=5,p=4$cm94YnRVOW5jZzFzcVE4bQ$fBxypOL0nP/zdPE71JtAV71i487LbX3fJI5PoTN6Lp4")))
	assert.Nil(t, hash.CompareArgon2id(context.Background(), p, []byte("test"), []byte("$v=19$m=32,t=5,p=4$cm94YnRVOW5jZzFzcVE4bQ$fBxypOL0nP/zdPE71JtAV71i487LbX3fJI5PoTN6Lp4")))
	assert.Error(t, hash.Compare(context.Background(), p, []byte("test"), []byte("$argon2id$v=19$m=32,t=5,p=4$cm94YnRVOW5jZzFzcVE4bQ$fBxypOL0nP/zdPE71JtAV71i487LbX3fJI5PoTN6Lp5")))

	assert.Nil(t, hash.Compare(context.Background(), p, []byte("test"), []byte("$bcryptaes$f5298ce2aed0e76e9f8d2edb905a6c42da007357d2c1778bd12ca5351d7085f0d72330b9b321383a50cd142a3ede42223b4dcbf8efc4674a86b1daac55b7161d5f57d3fd2b58a02c9995bb0394d3ab795fcf11aded25899c")))
	assert.Nil(t, hash.CompareBcryptAes(context.Background(), p, []byte("test"), []byte("$f5298ce2aed0e76e9f8d2edb905a6c42da007357d2c1778bd12ca5351d7085f0d72330b9b321383a50cd142a3ede42223b4dcbf8efc4674a86b1daac55b7161d5f57d3fd2b58a02c9995bb0394d3ab795fcf11aded25899c")))
	assert.Error(t, hash.Compare(context.Background(), p, []byte("test"), []byte("$bcryptaes$f5298ce2aed0e76e9f8d2edb905a6c42da007357d2c1778bd12ca5351d7085f0d72330b9b321383a50cd142a3ede42223b4dcbf8efc4674a86b1daac55b7161d5f57d3fd2b58a02c9995bb0394d3ab795fcf11aded25899d")))

	assert.Nil(t, hash.Compare(context.Background(), p, []byte("test"), []byte("$bcryptaes$485b35689563835d1d41b7e520822167e1645f36a237805c95a465cfea1e722cd5118d8983717b1dc26534f8806808e5c108cedf5a4744f961b851159dd2c9cd4d8d5cd27434ec1df2d84b78a0cdcdcb2ae874b9cc014665")))
	assert.Nil(t, hash.CompareBcryptAes(context.Background(), p, []byte("test"), []byte("$485b35689563835d1d41b7e520822167e1645f36a237805c95a465cfea1e722cd5118d8983717b1dc26534f8806808e5c108cedf5a4744f961b851159dd2c9cd4d8d5cd27434ec1df2d84b78a0cdcdcb2ae874b9cc014665")))
	assert.Error(t, hash.Compare(context.Background(), p, []byte("test"), []byte("$bcryptaes$485b35689563835d1d41b7e520822167e1645f36a237805c95a465cfea1e722cd5118d8983717b1dc26534f8806808e5c108cedf5a4744f961b851159dd2c9cd4d8d5cd27434ec1df2d84b78a0cdcdcb2ae874b9cc014666")))

	// use secondary encryption key (test key rotation)
	assert.Nil(t, hash.Compare(context.Background(), p, []byte("test"), []byte("$bcryptaes$afc9b874b712ac06ab838ef6f1d50cb1fd804d9ad2e73c6713cbfc2bead081b38d7944cd1329447ac90d86b142d48daef9fb9e1d0c9b5fc22b041b3d4d087d50a0b6a7d4ae641bea3d6ddbd5b272b04dfad504e49073dc48")))
	assert.Nil(t, hash.CompareBcryptAes(context.Background(), p, []byte("test"), []byte("$afc9b874b712ac06ab838ef6f1d50cb1fd804d9ad2e73c6713cbfc2bead081b38d7944cd1329447ac90d86b142d48daef9fb9e1d0c9b5fc22b041b3d4d087d50a0b6a7d4ae641bea3d6ddbd5b272b04dfad504e49073dc48")))
	assert.Error(t, hash.Compare(context.Background(), p, []byte("test"), []byte("$bcryptaes$afc9b874b712ac06ab838ef6f1d50cb1fd804d9ad2e73c6713cbfc2bead081b38d7944cd1329447ac90d86b142d48daef9fb9e1d0c9b5fc22b041b3d4d087d50a0b6a7d4ae641bea3d6ddbd5b272b04dfad504e49073dc49")))
}
