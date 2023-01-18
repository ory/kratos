// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package hash_test

import (
	"context"
	"crypto/rand"
	"encoding/base64"
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
			hs, err := hasher.Generate(context.Background(), pw)

			assert.Nil(t, err)
			assert.True(t, hasher.Understands(hs))

			// Valid format: $2a$12$[22 character salt][31 character hash]
			assert.Equal(t, 60, len(string(hs)), "invalid bcrypt hash length")
			assert.Equal(t, "$2a$04$", string(hs)[:7], "invalid bcrypt identifier")
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
			assert.True(t, hasher.Understands(hs))

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

func TestPbkdf2Hasher(t *testing.T) {
	for k, pw := range [][]byte{
		mkpw(t, 8),
		mkpw(t, 16),
		mkpw(t, 32),
		mkpw(t, 64),
		mkpw(t, 128),
	} {
		t.Run(fmt.Sprintf("case=%d", k), func(t *testing.T) {
			for kk, hasher := range []hash.Hasher{
				&hash.Pbkdf2{
					Algorithm:  "sha1",
					Iterations: 100000,
					SaltLength: 32,
					KeyLength:  32,
				},
				&hash.Pbkdf2{
					Algorithm:  "sha224",
					Iterations: 100000,
					SaltLength: 32,
					KeyLength:  32,
				},
				&hash.Pbkdf2{
					Algorithm:  "sha256",
					Iterations: 100000,
					SaltLength: 32,
					KeyLength:  32,
				},
				&hash.Pbkdf2{
					Algorithm:  "sha384",
					Iterations: 100000,
					SaltLength: 32,
					KeyLength:  32,
				},
				&hash.Pbkdf2{
					Algorithm:  "sha512",
					Iterations: 100000,
					SaltLength: 32,
					KeyLength:  32,
				},
			} {
				t.Run(fmt.Sprintf("hasher=%T/password=%d", hasher, kk), func(t *testing.T) {
					hs, err := hasher.Generate(context.Background(), pw)
					require.NoError(t, err)
					assert.NotEqual(t, pw, hs)

					t.Logf("hash: %s", hs)
					require.NoError(t, hash.ComparePbkdf2(context.Background(), pw, hs))

					assert.True(t, hasher.Understands(hs))

					mod := make([]byte, len(pw))
					copy(mod, pw)
					mod[len(pw)-1] = ^pw[len(pw)-1]
					require.Error(t, hash.ComparePbkdf2(context.Background(), mod, hs))
				})
			}
		})
	}
}

func TestCompare(t *testing.T) {
	assert.Error(t, hash.Compare(context.Background(), []byte("test"), []byte("$unknown$12$o6hx.Wog/wvFSkT/Bp/6DOxCtLRTDj7lm9on9suF/WaCGNVHbkfL6")))

	assert.Nil(t, hash.Compare(context.Background(), []byte("test"), []byte("$2a$12$o6hx.Wog/wvFSkT/Bp/6DOxCtLRTDj7lm9on9suF/WaCGNVHbkfL6")))
	assert.Nil(t, hash.CompareBcrypt(context.Background(), []byte("test"), []byte("$2a$12$o6hx.Wog/wvFSkT/Bp/6DOxCtLRTDj7lm9on9suF/WaCGNVHbkfL6")))
	assert.Error(t, hash.Compare(context.Background(), []byte("test"), []byte("$2a$12$o6hx.Wog/wvFSkT/Bp/6DOxCtLRTDj7lm9on9suF/WaCGNVHbkfL7")))

	assert.Nil(t, hash.Compare(context.Background(), []byte("test"), []byte("$2a$15$GRvRO2nrpYTEuPQX6AieaOlZ4.7nMGsXpt.QWMev1zrP86JNspZbO")))
	assert.Nil(t, hash.CompareBcrypt(context.Background(), []byte("test"), []byte("$2a$15$GRvRO2nrpYTEuPQX6AieaOlZ4.7nMGsXpt.QWMev1zrP86JNspZbO")))
	assert.Error(t, hash.Compare(context.Background(), []byte("test"), []byte("$2a$15$GRvRO2nrpYTEuPQX6AieaOlZ4.7nMGsXpt.QWMev1zrP86JNspZb1")))

	assert.Nil(t, hash.Compare(context.Background(), []byte("test"), []byte("$argon2id$v=19$m=32,t=2,p=4$cm94YnRVOW5jZzFzcVE4bQ$MNzk5BtR2vUhrp6qQEjRNw")))
	assert.Nil(t, hash.CompareArgon2id(context.Background(), []byte("test"), []byte("$argon2id$v=19$m=32,t=2,p=4$cm94YnRVOW5jZzFzcVE4bQ$MNzk5BtR2vUhrp6qQEjRNw")))
	assert.Error(t, hash.Compare(context.Background(), []byte("test"), []byte("$argon2id$v=19$m=32,t=2,p=4$cm94YnRVOW5jZzFzcVE4bQ$MNzk5BtR2vUhrp6qQEjRN2")))

	assert.Nil(t, hash.Compare(context.Background(), []byte("test"), []byte("$argon2i$v=19$m=65536,t=3,p=4$kk51rW/vxIVCYn+EG4kTSg$NyT88uraJ6im6dyha/M5jhXvpqlEdlS/9fEm7ScMb8c")))
	assert.Nil(t, hash.CompareArgon2i(context.Background(), []byte("test"), []byte("$argon2i$v=19$m=65536,t=3,p=4$kk51rW/vxIVCYn+EG4kTSg$NyT88uraJ6im6dyha/M5jhXvpqlEdlS/9fEm7ScMb8c")))
	assert.Error(t, hash.Compare(context.Background(), []byte("test"), []byte("$argon2i$v=19$m=65536,t=3,p=4$pZ+27D6B0bCi0DwSmANF1w$4RNCUu4Uyu7eTIvzIdSuKz+I9idJlX/ykn6J10/W0EU")))

	assert.Nil(t, hash.Compare(context.Background(), []byte("test"), []byte("$argon2id$v=19$m=32,t=5,p=4$cm94YnRVOW5jZzFzcVE4bQ$fBxypOL0nP/zdPE71JtAV71i487LbX3fJI5PoTN6Lp4")))
	assert.Nil(t, hash.CompareArgon2id(context.Background(), []byte("test"), []byte("$argon2id$v=19$m=32,t=5,p=4$cm94YnRVOW5jZzFzcVE4bQ$fBxypOL0nP/zdPE71JtAV71i487LbX3fJI5PoTN6Lp4")))
	assert.Error(t, hash.Compare(context.Background(), []byte("test"), []byte("$argon2id$v=19$m=32,t=5,p=4$cm94YnRVOW5jZzFzcVE4bQ$fBxypOL0nP/zdPE71JtAV71i487LbX3fJI5PoTN6Lp5")))

	assert.Nil(t, hash.Compare(context.Background(), []byte("test"), []byte("$pbkdf2-sha256$i=100000,l=32$1jP+5Zxpxgtee/iPxGgOz0RfE9/KJuDElP1ley4VxXc$QJxzfvdbHYBpydCbHoFg3GJEqMFULwskiuqiJctoYpI")))
	assert.Nil(t, hash.ComparePbkdf2(context.Background(), []byte("test"), []byte("$pbkdf2-sha256$i=100000,l=32$1jP+5Zxpxgtee/iPxGgOz0RfE9/KJuDElP1ley4VxXc$QJxzfvdbHYBpydCbHoFg3GJEqMFULwskiuqiJctoYpI")))
	assert.Error(t, hash.Compare(context.Background(), []byte("test"), []byte("$pbkdf2-sha256$i=100000,l=32$1jP+5Zxpxgtee/iPxGgOz0RfE9/KJuDElP1ley4VxXc$QJxzfvdbHYBpydCbHoFg3GJEqMFULwskiuqiJctoYpp")))

	assert.Nil(t, hash.Compare(context.Background(), []byte("test"), []byte("$pbkdf2-sha512$i=100000,l=32$bdHBpn7OWOivJMVJypy2UqR0UnaD5prQXRZevj/05YU$+wArTfv1a+bNGO1iZrmEdVjhA+lL11wF4/IxpgYfPwc")))
	assert.Nil(t, hash.ComparePbkdf2(context.Background(), []byte("test"), []byte("$pbkdf2-sha512$i=100000,l=32$bdHBpn7OWOivJMVJypy2UqR0UnaD5prQXRZevj/05YU$+wArTfv1a+bNGO1iZrmEdVjhA+lL11wF4/IxpgYfPwc")))
	assert.Error(t, hash.Compare(context.Background(), []byte("test"), []byte("$pbkdf2-sha512$i=100000,l=32$bdHBpn7OWOivJMVJypy2UqR0UnaD5prQXRZevj/05YU$+wArTfv1a+bNGO1iZrmEdVjhA+lL11wF4/IxpgYfPww")))

	assert.Error(t, hash.Compare(context.Background(), []byte("test"), []byte("$pbkdf2-sha256$1jP+5Zxpxgtee/iPxGgOz0RfE9/KJuDElP1ley4VxXc$QJxzfvdbHYBpydCbHoFg3GJEqMFULwskiuqiJctoYpI")))
	assert.Error(t, hash.Compare(context.Background(), []byte("test"), []byte("$pbkdf2-sha256$aaaa$1jP+5Zxpxgtee/iPxGgOz0RfE9/KJuDElP1ley4VxXc$QJxzfvdbHYBpydCbHoFg3GJEqMFULwskiuqiJctoYpI")))
	assert.Error(t, hash.Compare(context.Background(), []byte("test"), []byte("$pbkdf2-sha256$i=100000,l=32$1jP+5Zxpxgtee/iPxGgOz0RfE9/KJuDElP1ley4VxXcc$QJxzfvdbHYBpydCbHoFg3GJEqMFULwskiuqiJctoYpI")))
	assert.Error(t, hash.Compare(context.Background(), []byte("test"), []byte("$pbkdf2-sha256$i=100000,l=32$1jP+5Zxpxgtee/iPxGgOz0RfE9/KJuDElP1ley4VxXc$QJxzfvdbHYBpydCbHoFg3GJEqMFULwskiuqiJctoYpII")))
	assert.Error(t, hash.Compare(context.Background(), []byte("test"), []byte("$pbkdf2-sha512$I=100000,l=32$bdHBpn7OWOivJMVJypy2UqR0UnaD5prQXRZevj/05YU$+wArTfv1a+bNGO1iZrmEdVjhA+lL11wF4/IxpgYfPwc")))

	assert.Nil(t, hash.Compare(context.Background(), []byte("test"), []byte("$scrypt$ln=16384,r=8,p=1$2npRo7P03Mt8keSoMbyD/tKFWyUzjiQf2svUaNDSrhA=$MiCzNcIplSMqSBrm4HckjYqYhaVPPjTARTzwB1cVNYE=")))
	assert.Nil(t, hash.CompareScrypt(context.Background(), []byte("test"), []byte("$scrypt$ln=16384,r=8,p=1$2npRo7P03Mt8keSoMbyD/tKFWyUzjiQf2svUaNDSrhA=$MiCzNcIplSMqSBrm4HckjYqYhaVPPjTARTzwB1cVNYE=")))
	assert.Error(t, hash.Compare(context.Background(), []byte("test"), []byte("$scrypt$ln=16384,r=8,p=1$2npRo7P03Mt8keSoMbyD/tKFWyUzjiQf2svUaNDSrhA=$MiCzNcIplSMqSBrm4HckjYqYhaVPPjTARTzwB1cVNYF=")))

	assert.Nil(t, hash.Compare(context.Background(), []byte("8x4WjoDbSxJZdR"), []byte("$firescrypt$ln=14,r=8,p=1$sPtDhWcd1MfdAw==$xbSou7FOl6mChCyzpCPIQ7tku7nsQMTFtyOZSXXd7tjBa4NtimOx7v42Gv2SfzPQu1oxM2/k4SsbOu73wlKe1A==$Bw==$YE0dO4bwD4JnJafh6lZZfkp1MtKzuKAXQcDCJNJNyeCHairWHKENOkbh3dzwaCdizzOspwr/FITUVlnOAwPKyw==")))
	assert.Nil(t, hash.CompareFirebaseScrypt(context.Background(), []byte("8x4WjoDbSxJZdR"), []byte("$firescrypt$ln=14,r=8,p=1$sPtDhWcd1MfdAw==$xbSou7FOl6mChCyzpCPIQ7tku7nsQMTFtyOZSXXd7tjBa4NtimOx7v42Gv2SfzPQu1oxM2/k4SsbOu73wlKe1A==$Bw==$YE0dO4bwD4JnJafh6lZZfkp1MtKzuKAXQcDCJNJNyeCHairWHKENOkbh3dzwaCdizzOspwr/FITUVlnOAwPKyw==")))
	assert.Error(t, hash.Compare(context.Background(), []byte("8x4WjoDbSxJZdR"), []byte("$firescrypt$ln=14,r=8,p=1$sPtDhWcd1MfdAw==$xbSou7FOl6mChCyzpCPIQ7tku7nsQMTFtyOZSXXd7tjBa4NtimOx7v42Gv2SfzPQu1oxM2/k4SsbOu73wlKe1A==$Bw==$YE0dO4bwD4JnJafh6lZZfkp1MtKzuKAXQcDCJNJNyeCHairWHKENOkbh3dzwaCdizzOspwr/FITUVlnOAwPKyc==")))

	assert.Error(t, hash.Compare(context.Background(), []byte("test"), []byte("$scrypt$2npRo7P03Mt8keSoMbyD/tKFWyUzjiQf2svUaNDSrhA=$MiCzNcIplSMqSBrm4HckjYqYhaVPPjTARTzwB1cVNYE=")))
	assert.Error(t, hash.Compare(context.Background(), []byte("test"), []byte("$scrypt$ln=16384,r=8,p=1$(2npRo7P03Mt8keSoMbyD/tKFWyUzjiQf2svUaNDSrhA=$MiCzNcIplSMqSBrm4HckjYqYhaVPPjTARTzwB1cVNYE=")))
	assert.Error(t, hash.Compare(context.Background(), []byte("test"), []byte("$scrypt$ln=16384,r=8,p=1$(2npRo7P03Mt8keSoMbyD/tKFWyUzjiQf2svUaNDSrhA=$(MiCzNcIplSMqSBrm4HckjYqYhaVPPjTARTzwB1cVNYE=")))
	assert.Error(t, hash.Compare(context.Background(), []byte("test"), []byte("$scrypt$ln=16385,r=8,p=1$2npRo7P03Mt8keSoMbyD/tKFWyUzjiQf2svUaNDSrhA=$MiCzNcIplSMqSBrm4HckjYqYhaVPPjTARTzwB1cVNYE=")))
	assert.Error(t, hash.Compare(context.Background(), []byte("tesu"), []byte("$scrypt$ln=16384,r=8,p=1$2npRo7P03Mt8keSoMbyD/tKFWyUzjiQf2svUaNDSrhA=$MiCzNcIplSMqSBrm4HckjYqYhaVPPjTARTzwB1cVNYE=")))
	assert.Error(t, hash.Compare(context.Background(), []byte("tesu"), []byte("$scrypt$ln=abc,r=8,p=1$2npRo7P03Mt8keSoMbyD/tKFWyUzjiQf2svUaNDSrhA=$MiCzNcIplSMqSBrm4HckjYqYhaVPPjTARTzwB1cVNYE=")))

	assert.Nil(t, hash.Compare(context.Background(), []byte("test"), []byte("$md5$CY9rzUYh03PK3k6DJie09g==")))
	assert.Nil(t, hash.CompareMD5(context.Background(), []byte("test"), []byte("$md5$CY9rzUYh03PK3k6DJie09g==")))
	assert.Error(t, hash.Compare(context.Background(), []byte("test"), []byte("$md5$WhBei51A4TKXgNYuoiZdig==")))
	assert.Error(t, hash.Compare(context.Background(), []byte("test"), []byte("$md5$Dk/E5LQLsx4yt8QbUbvpdg==")))

	assert.Nil(t, hash.Compare(context.Background(), []byte("ory"), []byte("$md5$ptoWyof5SobW+pbZu2QXoQ==")))
	assert.Nil(t, hash.CompareMD5(context.Background(), []byte("ory"), []byte("$md5$ptoWyof5SobW+pbZu2QXoQ==")))
	assert.Error(t, hash.Compare(context.Background(), []byte("ory"), []byte("$md5$4skj967KRHFsnPFoL5dMMw==")))

	assert.ErrorIs(t, hash.Compare(context.Background(), []byte("ory"), []byte("$md5$$")), hash.ErrInvalidHash)
	assert.Error(t, hash.Compare(context.Background(), []byte("ory"), []byte("$md5$$$")))
	assert.Error(t, hash.Compare(context.Background(), []byte("ory"), []byte("$md5$pf=$$")))
	assert.ErrorIs(t, hash.Compare(context.Background(), []byte("ory"), []byte("$md5$pf=MTIz$Z$")), base64.CorruptInputError(0))
	assert.ErrorIs(t, hash.Compare(context.Background(), []byte("ory"), []byte("$md5$pf=MTIz$Z$")), base64.CorruptInputError(0))
	assert.ErrorIs(t, hash.Compare(context.Background(), []byte("ory"), []byte("$md5$pf=MTIz$MTIz$Z")), base64.CorruptInputError(0))

	assert.Nil(t, hash.Compare(context.Background(), []byte("test"), []byte("$md5$pf=e1NBTFR9e1BBU1NXT1JEfQ==$MTIz$q+RdKCgc+ipCAcm5ChQwlQ=="))) // pf={SALT}{PASSWORD} salt=123
	assert.Error(t, hash.Compare(context.Background(), []byte("test"), []byte("$md5$pf=e1NBTFR9e1BBU1NXT1JEfQ==$MTIz$hh8ZTp1hGPPZQqcr4+UXSQ==")))

	assert.Nil(t, hash.CompareMD5(context.Background(), []byte("test"), []byte("$md5$pf=e1NBTFR9JCR7UEFTU1dPUkR9$MTIzNA==$ud392Z8rfZ+Ou7ZFXYLKbA=="))) // pf={SALT}$${PASSWORD} salt=1234
	assert.Error(t, hash.CompareMD5(context.Background(), []byte("test1"), []byte("$md5$pf=e1NBTFR9JCR7UEFTU1dPUkR9$MTIzNA==$ud392Z8rfZ+Ou7ZFXYLKbA==")))

	assert.Nil(t, hash.CompareMD5(context.Background(), []byte("ory"), []byte("$md5$pf=e1BBU1NXT1JEfXtTQUxUfSQ/$MTIzNDU2Nzg5$8PhwWanVRnpJAFK4NUjR0w=="))) // pf={PASSWORD}{SALT}$? salt=123456789
	assert.Error(t, hash.CompareMD5(context.Background(), []byte("ory1"), []byte("$md5$pf=e1BBU1NXT1JEfXtTQUxUfSQ/$MTIzNDU2Nzg5$8PhwWanVRnpJAFK4NUjR0w==")))
}
