// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package hash_test

import (
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
	t.Parallel()
	ctx := t.Context()
	_, reg := internal.NewVeryFastRegistryWithoutDB(t)
	h := hash.NewHasherArgon2(reg)
	for _, pwLength := range []int{
		8,
		16,
		32,
		64,
		128,
	} {
		pwLength := pwLength
		t.Run(fmt.Sprintf("length=%dchars", pwLength), func(t *testing.T) {
			t.Parallel()
			pw := mkpw(t, pwLength)
			hs, err := h.Generate(ctx, pw)
			require.NoError(t, err)
			assert.NotEqual(t, pw, hs)

			t.Logf("hash: %s", hs)
			require.NoError(t, hash.CompareArgon2id(ctx, pw, hs))

			mod := make([]byte, len(pw))
			copy(mod, pw)
			mod[len(pw)-1] = ^pw[len(pw)-1]
			require.Error(t, hash.CompareArgon2id(ctx, mod, hs))
		})
	}
}

func TestBcryptHasherGeneratesErrorWhenPasswordIsLong(t *testing.T) {
	t.Parallel()
	ctx := t.Context()

	_, reg := internal.NewVeryFastRegistryWithoutDB(t)
	hasher := hash.NewHasherBcrypt(reg)

	password := mkpw(t, 73)
	res, err := hasher.Generate(ctx, password)

	assert.Error(t, err, "password is too long")
	assert.Nil(t, res)
}

func TestBcryptHasherGeneratesHash(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	_, reg := internal.NewVeryFastRegistryWithoutDB(t)
	hasher := hash.NewHasherBcrypt(reg)
	for _, pwLength := range []int{
		8,
		16,
		32,
		64,
		72,
	} {
		pwLength := pwLength
		t.Run(fmt.Sprintf("length=%dchars", pwLength), func(t *testing.T) {
			t.Parallel()
			pw := mkpw(t, pwLength)
			hs, err := hasher.Generate(ctx, pw)

			assert.Nil(t, err)
			assert.True(t, hasher.Understands(hs))

			// Valid format: $2a$12$[22 character salt][31 character hash]
			assert.Equal(t, 60, len(string(hs)), "invalid bcrypt hash length")
			assert.Equal(t, "$2a$12$", string(hs)[:7], "invalid bcrypt identifier")
		})
	}
}

func TestComparatorBcryptFailsWhenPasswordIsTooLong(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	password := mkpw(t, 73)
	err := hash.CompareBcrypt(ctx, password, []byte("hash"))

	assert.Error(t, err, "password is too long")
}

func TestComparatorBcryptSuccess(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	_, reg := internal.NewVeryFastRegistryWithoutDB(t)
	hasher := hash.NewHasherBcrypt(reg)
	for _, pwLength := range []int{
		8,
		16,
		32,
		64,
		72,
	} {
		pwLength := pwLength
		t.Run(fmt.Sprintf("length=%dchars", pwLength), func(t *testing.T) {
			t.Parallel()
			pw := mkpw(t, pwLength)

			hs, err := hasher.Generate(ctx, pw)

			assert.Nil(t, err)
			assert.True(t, hasher.Understands(hs))

			err = hash.CompareBcrypt(ctx, pw, hs)
			assert.Nil(t, err, "hash validation fails")
		})
	}
}

func TestComparatorBcryptFail(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	for _, pwLength := range []int{
		8,
		16,
		32,
		64,
		72,
	} {
		pwLength := pwLength
		t.Run(fmt.Sprintf("length=%dchars", pwLength), func(t *testing.T) {
			t.Parallel()
			pw := mkpw(t, pwLength)
			mod := make([]byte, len(pw))
			copy(mod, pw)
			mod[len(pw)-1] = ^pw[len(pw)-1]

			err := hash.CompareBcrypt(ctx, pw, mod)
			assert.Error(t, err)
		})
	}
}

func TestPbkdf2Hasher(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	for _, pwLength := range []int{
		8,
		16,
		32,
		64,
		128,
	} {
		pwLength := pwLength
		t.Run(fmt.Sprintf("length=%dchars", pwLength), func(t *testing.T) {
			t.Parallel()
			for _, algorithm := range []string{
				"sha1",
				"sha224",
				"sha256",
				"sha384",
				"sha512",
			} {
				algorithm := algorithm
				t.Run(fmt.Sprintf("algorithm=%s", algorithm), func(t *testing.T) {
					t.Parallel()
					hasher := &hash.Pbkdf2{
						Algorithm:  algorithm,
						Iterations: 100_000,
						SaltLength: 32,
						KeyLength:  32,
					}
					pw := mkpw(t, pwLength)
					t.Logf("%d", pwLength)
					hs, err := hasher.Generate(ctx, pw)
					require.NoError(t, err)
					assert.NotEqual(t, pw, hs)

					t.Logf("hash: %s", hs)
					require.NoError(t, hash.ComparePbkdf2(ctx, pw, hs))

					assert.True(t, hasher.Understands(hs))

					mod := make([]byte, len(pw))
					copy(mod, pw)
					mod[len(pw)-1] = ^pw[len(pw)-1]
					require.Error(t, hash.ComparePbkdf2(ctx, mod, hs))
				})
			}
		})
	}
}

func TestCompare(t *testing.T) {
	t.Parallel()
	ctx := t.Context()

	t.Run("unknown", func(t *testing.T) {
		t.Parallel()
		assert.Error(t, hash.Compare(ctx, []byte("test"), []byte("$unknown$12$o6hx.Wog/wvFSkT/Bp/6DOxCtLRTDj7lm9on9suF/WaCGNVHbkfL6")))
	})

	t.Run("bcrypt", func(t *testing.T) {
		t.Parallel()
		assert.Nil(t, hash.Compare(ctx, []byte("test"), []byte("$2a$12$o6hx.Wog/wvFSkT/Bp/6DOxCtLRTDj7lm9on9suF/WaCGNVHbkfL6")))
		assert.Nil(t, hash.CompareBcrypt(ctx, []byte("test"), []byte("$2a$12$o6hx.Wog/wvFSkT/Bp/6DOxCtLRTDj7lm9on9suF/WaCGNVHbkfL6")))
		assert.Error(t, hash.Compare(ctx, []byte("test"), []byte("$2a$12$o6hx.Wog/wvFSkT/Bp/6DOxCtLRTDj7lm9on9suF/WaCGNVHbkfL7")))

		assert.Nil(t, hash.Compare(ctx, []byte("test"), []byte("$2a$15$GRvRO2nrpYTEuPQX6AieaOlZ4.7nMGsXpt.QWMev1zrP86JNspZbO")))
		assert.Nil(t, hash.CompareBcrypt(ctx, []byte("test"), []byte("$2a$15$GRvRO2nrpYTEuPQX6AieaOlZ4.7nMGsXpt.QWMev1zrP86JNspZbO")))
		assert.Error(t, hash.Compare(ctx, []byte("test"), []byte("$2a$15$GRvRO2nrpYTEuPQX6AieaOlZ4.7nMGsXpt.QWMev1zrP86JNspZb1")))
	})

	t.Run("Argon2", func(t *testing.T) {
		t.Parallel()
		assert.Nil(t, hash.Compare(ctx, []byte("test"), []byte("$argon2id$v=19$m=32,t=2,p=4$cm94YnRVOW5jZzFzcVE4bQ$MNzk5BtR2vUhrp6qQEjRNw")))
		assert.Nil(t, hash.CompareArgon2id(ctx, []byte("test"), []byte("$argon2id$v=19$m=32,t=2,p=4$cm94YnRVOW5jZzFzcVE4bQ$MNzk5BtR2vUhrp6qQEjRNw")))
		assert.Error(t, hash.Compare(ctx, []byte("test"), []byte("$argon2id$v=19$m=32,t=2,p=4$cm94YnRVOW5jZzFzcVE4bQ$MNzk5BtR2vUhrp6qQEjRN2")))

		assert.Nil(t, hash.Compare(ctx, []byte("test"), []byte("$argon2i$v=19$m=65536,t=3,p=4$kk51rW/vxIVCYn+EG4kTSg$NyT88uraJ6im6dyha/M5jhXvpqlEdlS/9fEm7ScMb8c")))
		assert.Nil(t, hash.CompareArgon2i(ctx, []byte("test"), []byte("$argon2i$v=19$m=65536,t=3,p=4$kk51rW/vxIVCYn+EG4kTSg$NyT88uraJ6im6dyha/M5jhXvpqlEdlS/9fEm7ScMb8c")))
		assert.Error(t, hash.Compare(ctx, []byte("test"), []byte("$argon2i$v=19$m=65536,t=3,p=4$pZ+27D6B0bCi0DwSmANF1w$4RNCUu4Uyu7eTIvzIdSuKz+I9idJlX/ykn6J10/W0EU")))

		assert.Nil(t, hash.Compare(ctx, []byte("test"), []byte("$argon2id$v=19$m=32,t=5,p=4$cm94YnRVOW5jZzFzcVE4bQ$fBxypOL0nP/zdPE71JtAV71i487LbX3fJI5PoTN6Lp4")))
		assert.Nil(t, hash.CompareArgon2id(ctx, []byte("test"), []byte("$argon2id$v=19$m=32,t=5,p=4$cm94YnRVOW5jZzFzcVE4bQ$fBxypOL0nP/zdPE71JtAV71i487LbX3fJI5PoTN6Lp4")))
		assert.Error(t, hash.Compare(ctx, []byte("test"), []byte("$argon2id$v=19$m=32,t=5,p=4$cm94YnRVOW5jZzFzcVE4bQ$fBxypOL0nP/zdPE71JtAV71i487LbX3fJI5PoTN6Lp5")))
	})

	t.Run("pbkdf2", func(t *testing.T) {
		t.Parallel()
		assert.Nil(t, hash.Compare(ctx, []byte("test"), []byte("$pbkdf2-sha256$i=100000,l=32$1jP+5Zxpxgtee/iPxGgOz0RfE9/KJuDElP1ley4VxXc$QJxzfvdbHYBpydCbHoFg3GJEqMFULwskiuqiJctoYpI")))
		assert.Nil(t, hash.ComparePbkdf2(ctx, []byte("test"), []byte("$pbkdf2-sha256$i=100000,l=32$1jP+5Zxpxgtee/iPxGgOz0RfE9/KJuDElP1ley4VxXc$QJxzfvdbHYBpydCbHoFg3GJEqMFULwskiuqiJctoYpI")))
		assert.Error(t, hash.Compare(ctx, []byte("test"), []byte("$pbkdf2-sha256$i=100000,l=32$1jP+5Zxpxgtee/iPxGgOz0RfE9/KJuDElP1ley4VxXc$QJxzfvdbHYBpydCbHoFg3GJEqMFULwskiuqiJctoYpp")))

		assert.Nil(t, hash.Compare(ctx, []byte("test"), []byte("$pbkdf2-sha512$i=100000,l=32$bdHBpn7OWOivJMVJypy2UqR0UnaD5prQXRZevj/05YU$+wArTfv1a+bNGO1iZrmEdVjhA+lL11wF4/IxpgYfPwc")))
		assert.Nil(t, hash.ComparePbkdf2(ctx, []byte("test"), []byte("$pbkdf2-sha512$i=100000,l=32$bdHBpn7OWOivJMVJypy2UqR0UnaD5prQXRZevj/05YU$+wArTfv1a+bNGO1iZrmEdVjhA+lL11wF4/IxpgYfPwc")))
		assert.Error(t, hash.Compare(ctx, []byte("test"), []byte("$pbkdf2-sha512$i=100000,l=32$bdHBpn7OWOivJMVJypy2UqR0UnaD5prQXRZevj/05YU$+wArTfv1a+bNGO1iZrmEdVjhA+lL11wF4/IxpgYfPww")))

		assert.Error(t, hash.Compare(ctx, []byte("test"), []byte("$pbkdf2-sha256$1jP+5Zxpxgtee/iPxGgOz0RfE9/KJuDElP1ley4VxXc$QJxzfvdbHYBpydCbHoFg3GJEqMFULwskiuqiJctoYpI")))
		assert.Error(t, hash.Compare(ctx, []byte("test"), []byte("$pbkdf2-sha256$aaaa$1jP+5Zxpxgtee/iPxGgOz0RfE9/KJuDElP1ley4VxXc$QJxzfvdbHYBpydCbHoFg3GJEqMFULwskiuqiJctoYpI")))
		assert.Error(t, hash.Compare(ctx, []byte("test"), []byte("$pbkdf2-sha256$i=100000,l=32$1jP+5Zxpxgtee/iPxGgOz0RfE9/KJuDElP1ley4VxXcc$QJxzfvdbHYBpydCbHoFg3GJEqMFULwskiuqiJctoYpI")))
		assert.Error(t, hash.Compare(ctx, []byte("test"), []byte("$pbkdf2-sha256$i=100000,l=32$1jP+5Zxpxgtee/iPxGgOz0RfE9/KJuDElP1ley4VxXc$QJxzfvdbHYBpydCbHoFg3GJEqMFULwskiuqiJctoYpII")))
		assert.Error(t, hash.Compare(ctx, []byte("test"), []byte("$pbkdf2-sha512$I=100000,l=32$bdHBpn7OWOivJMVJypy2UqR0UnaD5prQXRZevj/05YU$+wArTfv1a+bNGO1iZrmEdVjhA+lL11wF4/IxpgYfPwc")))
	})

	t.Run("scrypt", func(t *testing.T) {
		t.Parallel()
		assert.Nil(t, hash.Compare(ctx, []byte("test"), []byte("$scrypt$ln=16384,r=8,p=1$2npRo7P03Mt8keSoMbyD/tKFWyUzjiQf2svUaNDSrhA=$MiCzNcIplSMqSBrm4HckjYqYhaVPPjTARTzwB1cVNYE=")))
		assert.Nil(t, hash.CompareScrypt(ctx, []byte("test"), []byte("$scrypt$ln=16384,r=8,p=1$2npRo7P03Mt8keSoMbyD/tKFWyUzjiQf2svUaNDSrhA=$MiCzNcIplSMqSBrm4HckjYqYhaVPPjTARTzwB1cVNYE=")))
		assert.Error(t, hash.Compare(ctx, []byte("test"), []byte("$scrypt$ln=16384,r=8,p=1$2npRo7P03Mt8keSoMbyD/tKFWyUzjiQf2svUaNDSrhA=$MiCzNcIplSMqSBrm4HckjYqYhaVPPjTARTzwB1cVNYF=")))

		assert.Error(t, hash.Compare(ctx, []byte("test"), []byte("$scrypt$2npRo7P03Mt8keSoMbyD/tKFWyUzjiQf2svUaNDSrhA=$MiCzNcIplSMqSBrm4HckjYqYhaVPPjTARTzwB1cVNYE=")))
		assert.Error(t, hash.Compare(ctx, []byte("test"), []byte("$scrypt$ln=16384,r=8,p=1$(2npRo7P03Mt8keSoMbyD/tKFWyUzjiQf2svUaNDSrhA=$MiCzNcIplSMqSBrm4HckjYqYhaVPPjTARTzwB1cVNYE=")))
		assert.Error(t, hash.Compare(ctx, []byte("test"), []byte("$scrypt$ln=16384,r=8,p=1$(2npRo7P03Mt8keSoMbyD/tKFWyUzjiQf2svUaNDSrhA=$(MiCzNcIplSMqSBrm4HckjYqYhaVPPjTARTzwB1cVNYE=")))
		assert.Error(t, hash.Compare(ctx, []byte("test"), []byte("$scrypt$ln=16385,r=8,p=1$2npRo7P03Mt8keSoMbyD/tKFWyUzjiQf2svUaNDSrhA=$MiCzNcIplSMqSBrm4HckjYqYhaVPPjTARTzwB1cVNYE=")))
		assert.Error(t, hash.Compare(ctx, []byte("tesu"), []byte("$scrypt$ln=16384,r=8,p=1$2npRo7P03Mt8keSoMbyD/tKFWyUzjiQf2svUaNDSrhA=$MiCzNcIplSMqSBrm4HckjYqYhaVPPjTARTzwB1cVNYE=")))
		assert.Error(t, hash.Compare(ctx, []byte("tesu"), []byte("$scrypt$ln=abc,r=8,p=1$2npRo7P03Mt8keSoMbyD/tKFWyUzjiQf2svUaNDSrhA=$MiCzNcIplSMqSBrm4HckjYqYhaVPPjTARTzwB1cVNYE=")))
	})

	t.Run("firescrypt", func(t *testing.T) {
		t.Parallel()
		assert.Nil(t, hash.Compare(ctx, []byte("8x4WjoDbSxJZdR"), []byte("$firescrypt$ln=14,r=8,p=1$sPtDhWcd1MfdAw==$xbSou7FOl6mChCyzpCPIQ7tku7nsQMTFtyOZSXXd7tjBa4NtimOx7v42Gv2SfzPQu1oxM2/k4SsbOu73wlKe1A==$Bw==$YE0dO4bwD4JnJafh6lZZfkp1MtKzuKAXQcDCJNJNyeCHairWHKENOkbh3dzwaCdizzOspwr/FITUVlnOAwPKyw==")))
		assert.Nil(t, hash.CompareFirebaseScrypt(ctx, []byte("8x4WjoDbSxJZdR"), []byte("$firescrypt$ln=14,r=8,p=1$sPtDhWcd1MfdAw==$xbSou7FOl6mChCyzpCPIQ7tku7nsQMTFtyOZSXXd7tjBa4NtimOx7v42Gv2SfzPQu1oxM2/k4SsbOu73wlKe1A==$Bw==$YE0dO4bwD4JnJafh6lZZfkp1MtKzuKAXQcDCJNJNyeCHairWHKENOkbh3dzwaCdizzOspwr/FITUVlnOAwPKyw==")))
		assert.Error(t, hash.Compare(ctx, []byte("8x4WjoDbSxJZdR"), []byte("$firescrypt$ln=14,r=8,p=1$sPtDhWcd1MfdAw==$xbSou7FOl6mChCyzpCPIQ7tku7nsQMTFtyOZSXXd7tjBa4NtimOx7v42Gv2SfzPQu1oxM2/k4SsbOu73wlKe1A==$Bw==$YE0dO4bwD4JnJafh6lZZfkp1MtKzuKAXQcDCJNJNyeCHairWHKENOkbh3dzwaCdizzOspwr/FITUVlnOAwPKyc==")))
	})

	t.Run("SSHA", func(t *testing.T) {
		t.Parallel()
		assert.Nil(t, hash.Compare(ctx, []byte("test123"), []byte("{SSHA}JFZFs0oHzxbMwkSJmYVeI8MnTDy/276a")))
		assert.Nil(t, hash.CompareSSHA(ctx, []byte("test123"), []byte("{SSHA}JFZFs0oHzxbMwkSJmYVeI8MnTDy/276a")))
		assert.Error(t, hash.CompareSSHA(ctx, []byte("badtest"), []byte("{SSHA}JFZFs0oHzxbMwkSJmYVeI8MnTDy/276a")))
		assert.Error(t, hash.Compare(ctx, []byte(""), []byte("{SSHA}tooshort")))

		assert.Nil(t, hash.Compare(ctx, []byte("test123"), []byte("{SSHA256}czO44OTV17PcF1cRxWrLZLy9xHd7CWyVYplr1rOhuMlx/7IK")))
		assert.Nil(t, hash.CompareSSHA(ctx, []byte("test123"), []byte("{SSHA256}czO44OTV17PcF1cRxWrLZLy9xHd7CWyVYplr1rOhuMlx/7IK")))
		assert.Error(t, hash.CompareSSHA(ctx, []byte("badtest"), []byte("{SSHA256}czO44OTV17PcF1cRxWrLZLy9xHd7CWyVYplr1rOhuMlx/7IK")))

		assert.Nil(t, hash.Compare(ctx, []byte("test123"), []byte("{SSHA512}xPUl/px+1cG55rUH4rzcwxdOIPSB2TingLpiJJumN2xyDWN4Ix1WQG3ihnvHaWUE8MYNkvMi5rf0C9NYixHsE6Yh59M=")))
		assert.Nil(t, hash.CompareSSHA(ctx, []byte("test123"), []byte("{SSHA512}xPUl/px+1cG55rUH4rzcwxdOIPSB2TingLpiJJumN2xyDWN4Ix1WQG3ihnvHaWUE8MYNkvMi5rf0C9NYixHsE6Yh59M=")))
		assert.Error(t, hash.CompareSSHA(ctx, []byte("badtest"), []byte("{SSHA512}xPUl/px+1cG55rUH4rzcwxdOIPSB2TingLpiJJumN2xyDWN4Ix1WQG3ihnvHaWUE8MYNkvMi5rf0C9NYixHsE6Yh59M=")))
		assert.Error(t, hash.CompareSSHA(ctx, []byte("test123"), []byte("{SSHAnotExistent}xPUl/px+1cG55rUH4rzcwxdOIPSB2TingLpiJJumN2xyDWN4Ix1WQG3ihnvHaWUE8MYNkvMi5rf0C9NYixHsE6Yh59M=")))
	})

	t.Run("sha1", func(t *testing.T) {
		t.Parallel()
		//pf: {SALT}{PASSWORD}
		assert.Nil(t, hash.Compare(ctx, []byte("test"), []byte("$sha1$pf=e1NBTFR9e1BBU1NXT1JEfQ==$NW9wbWtnejAzcg==$2qU2SGWP8viTM1md3FiI3+rjWXQ=")))
		assert.Error(t, hash.Compare(ctx, []byte("wrongpass"), []byte("$sha1$pf=e1NBTFR9e1BBU1NXT1JEfQ==$NW9wbWtnejAzcg==$2qU2SGWP8viTM1md3FiI3+rjWXQ=")))
		assert.Error(t, hash.Compare(ctx, []byte("tset"), []byte("$sha1$pf=e1NBTFR9e1BBU1NXT1JEfQ==$NW9wbWtnejAzcg==$2qU2SGWP8viTM1md3FiI3+rjWXQ=")))
		// wrong salt
		assert.Error(t, hash.Compare(ctx, []byte("test"), []byte("$sha1$pf=e1NBTFR9e1BBU1NXT1JEfQ==$cDJvb3ZrZGJ6cQ==$2qU2SGWP8viTM1md3FiI3+rjWXQ=")))
		// salt not encoded
		assert.Error(t, hash.Compare(ctx, []byte("test"), []byte("$sha1$pf=e1NBTFR9e1BBU1NXT1JEfQ==$5opmkgz03r$2qU2SGWP8viTM1md3FiI3+rjWXQ=")))
		assert.Nil(t, hash.Compare(ctx, []byte("BwS^514g^cv@Z"), []byte("$sha1$pf=e1NBTFR9e1BBU1NXT1JEfQ==$NW9wbWtnejAzcg==$99h9net4BXl7qdTRaiGUobLROxM=")))
		// no format string
		assert.Error(t, hash.Compare(ctx, []byte("test"), []byte("$sha1$pf=$NW9wbWtnejAzcg==$2qU2SGWP8viTM1md3FiI3+rjWXQ=")))
		assert.Error(t, hash.Compare(ctx, []byte("test"), []byte("$sha1$$NW9wbWtnejAzcg==$2qU2SGWP8viTM1md3FiI3+rjWXQ=")))
		// wrong number of parameters
		assert.Error(t, hash.Compare(ctx, []byte("test"), []byte("$sha1$NW9wbWtnejAzcg==$2qU2SGWP8viTM1md3FiI3+rjWXQ=")))
		// pf: ??staticPrefix??{SALT}{PASSWORD}
		assert.Nil(t, hash.Compare(ctx, []byte("test"), []byte("$sha1$pf=Pz9zdGF0aWNQcmVmaXg/P3tTQUxUfXtQQVNTV09SRH0=$NW9wbWtnejAzcg==$SAAxMUn7jxckQXkBmsVF0nHwqso=")))
		// pf: {PASSWORD}%%{SALT}
		assert.Nil(t, hash.Compare(ctx, []byte("test"), []byte("$sha1$pf=e1BBU1NXT1JEfSUle1NBTFR9$NW9wbWtnejAzcg==$YX0AW8/MW5ojUlnzTaR43ucHCog=")))
		// pf: ${PASSWORD}${SALT}$
		assert.Nil(t, hash.Compare(ctx, []byte("test"), []byte("$sha1$pf=JHtQQVNTV09SRH0ke1NBTFR9JA==$NW9wbWtnejAzcg==$iE5n1yjX3oAdxRHwZ4u57I4LpQo=")))
	})

	t.Run("sha256", func(t *testing.T) {
		t.Parallel()
		//pf: {SALT}{PASSWORD}
		assert.Nil(t, hash.Compare(ctx, []byte("test"), []byte("$sha256$pf=e1NBTFR9e1BBU1NXT1JEfQ==$NW9wbWtnejAzcg==$0gfRVLCvtBCk20udLDEY5vNhujWx7RGjwRIS1ebMsLY=")))
		assert.Nil(t, hash.CompareSHA(ctx, []byte("test"), []byte("$sha256$pf=e1NBTFR9e1BBU1NXT1JEfQ==$NW9wbWtnejAzcg==$0gfRVLCvtBCk20udLDEY5vNhujWx7RGjwRIS1ebMsLY=")))
		assert.Error(t, hash.Compare(ctx, []byte("wrongpass"), []byte("$sha256$pf=e1NBTFR9e1BBU1NXT1JEfQ==$NW9wbWtnejAzcg==$0gfRVLCvtBCk20udLDEY5vNhujWx7RGjwRIS1ebMsLY=")))
		//pf: {SALT}$${PASSWORD}
		assert.Nil(t, hash.Compare(ctx, []byte("test"), []byte("$sha256$pf=e1NBTFR9JCR7UEFTU1dPUkR9$NW9wbWtnejAzcg==$HokCOi9OtiZaZRvnkgemV3B4UUHpI7kA8zq/EZWH2NY=")))
	})

	t.Run("sha512", func(t *testing.T) {
		t.Parallel()
		//pf: {SALT}{PASSWORD}
		assert.Nil(t, hash.Compare(ctx, []byte("test"), []byte("$sha512$pf=e1NBTFR9e1BBU1NXT1JEfQ==$NW9wbWtnejAzcg==$6ctpVuApMNp0CgBXcdHw/GC562eFEFGr4gpgANX8ZYsX+j5B19IkdmOY2Fytsz3QUwSWdGcUjbqwgJGTH0UYvw==")))
		assert.Nil(t, hash.CompareSHA(ctx, []byte("test"), []byte("$sha512$pf=e1NBTFR9e1BBU1NXT1JEfQ==$NW9wbWtnejAzcg==$6ctpVuApMNp0CgBXcdHw/GC562eFEFGr4gpgANX8ZYsX+j5B19IkdmOY2Fytsz3QUwSWdGcUjbqwgJGTH0UYvw==")))
		assert.Error(t, hash.Compare(ctx, []byte("wrongpass"), []byte("$sha512$pf=e1NBTFR9e1BBU1NXT1JEfQ==$NW9wbWtnejAzcg==$6ctpVuApMNp0CgBXcdHw/GC562eFEFGr4gpgANX8ZYsX+j5B19IkdmOY2Fytsz3QUwSWdGcUjbqwgJGTH0UYvw==")))
		//pf: {SALT}$${PASSWORD}
		assert.Nil(t, hash.Compare(ctx, []byte("test"), []byte("$sha512$pf=e1NBTFR9JCR7UEFTU1dPUkR9$NW9wbWtnejAzcg==$1F9BPW8UtdJkZ9Dhlf+D4X4dJ9xfuH8y04EfuCP2k4aGPPq/aWxU9/xe3LydHmYW1/K3zu3NFO9ETVrZettz3w==")))
	})

	t.Run("sha unknown", func(t *testing.T) {
		t.Parallel()
		assert.Error(t, hash.Compare(ctx, []byte("test"), []byte("$shaNotExistent$pf=e1NBTFR9e1BBU1NXT1JEfQ==$NW9wbWtnejAzcg==$6ctpVuApMNp0CgBXcdHw/GC562eFEFGr4gpgANX8ZYsX+j5B19IkdmOY2Fytsz3QUwSWdGcUjbqwgJGTH0UYvw==")))
	})

	t.Run("md5", func(t *testing.T) {
		t.Parallel()

		assert.Nil(t, hash.Compare(ctx, []byte("test"), []byte("$md5$CY9rzUYh03PK3k6DJie09g==")))
		assert.Nil(t, hash.CompareMD5(ctx, []byte("test"), []byte("$md5$CY9rzUYh03PK3k6DJie09g==")))
		assert.Error(t, hash.Compare(ctx, []byte("test"), []byte("$md5$WhBei51A4TKXgNYuoiZdig==")))
		assert.Error(t, hash.Compare(ctx, []byte("test"), []byte("$md5$Dk/E5LQLsx4yt8QbUbvpdg==")))

		assert.Nil(t, hash.Compare(ctx, []byte("ory"), []byte("$md5$ptoWyof5SobW+pbZu2QXoQ==")))
		assert.Nil(t, hash.CompareMD5(ctx, []byte("ory"), []byte("$md5$ptoWyof5SobW+pbZu2QXoQ==")))
		assert.Error(t, hash.Compare(ctx, []byte("ory"), []byte("$md5$4skj967KRHFsnPFoL5dMMw==")))

		assert.ErrorIs(t, hash.Compare(ctx, []byte("ory"), []byte("$md5$$")), hash.ErrInvalidHash)
		assert.Error(t, hash.Compare(ctx, []byte("ory"), []byte("$md5$$$")))
		assert.Error(t, hash.Compare(ctx, []byte("ory"), []byte("$md5$pf=$$")))
		assert.ErrorIs(t, hash.Compare(ctx, []byte("ory"), []byte("$md5$pf=MTIz$Z$")), base64.CorruptInputError(0))
		assert.ErrorIs(t, hash.Compare(ctx, []byte("ory"), []byte("$md5$pf=MTIz$Z$")), base64.CorruptInputError(0))
		assert.ErrorIs(t, hash.Compare(ctx, []byte("ory"), []byte("$md5$pf=MTIz$MTIz$Z")), base64.CorruptInputError(0))

		assert.Nil(t, hash.Compare(ctx, []byte("test"), []byte("$md5$pf=e1NBTFR9e1BBU1NXT1JEfQ==$MTIz$q+RdKCgc+ipCAcm5ChQwlQ=="))) // pf={SALT}{PASSWORD} salt=123
		assert.Error(t, hash.Compare(ctx, []byte("test"), []byte("$md5$pf=e1NBTFR9e1BBU1NXT1JEfQ==$MTIz$hh8ZTp1hGPPZQqcr4+UXSQ==")))

		assert.Nil(t, hash.CompareMD5(ctx, []byte("test"), []byte("$md5$pf=e1NBTFR9JCR7UEFTU1dPUkR9$MTIzNA==$ud392Z8rfZ+Ou7ZFXYLKbA=="))) // pf={SALT}$${PASSWORD} salt=1234
		assert.Error(t, hash.CompareMD5(ctx, []byte("test1"), []byte("$md5$pf=e1NBTFR9JCR7UEFTU1dPUkR9$MTIzNA==$ud392Z8rfZ+Ou7ZFXYLKbA==")))

		assert.Nil(t, hash.CompareMD5(ctx, []byte("ory"), []byte("$md5$pf=e1BBU1NXT1JEfXtTQUxUfSQ/$MTIzNDU2Nzg5$8PhwWanVRnpJAFK4NUjR0w=="))) // pf={PASSWORD}{SALT}$? salt=123456789
		assert.Error(t, hash.CompareMD5(ctx, []byte("ory1"), []byte("$md5$pf=e1BBU1NXT1JEfXtTQUxUfSQ/$MTIzNDU2Nzg5$8PhwWanVRnpJAFK4NUjR0w==")))
	})

	t.Run("md5-crypt", func(t *testing.T) {
		t.Parallel()

		assert.Nil(t, hash.Compare(ctx, []byte("test"), []byte("$md5-crypt$TVEiiKNb$SN6/pUaRQS/E8Jh46As2C/")))
		assert.Nil(t, hash.CompareMD5Crypt(ctx, []byte("test"), []byte("$md5-crypt$TVEiiKNb$SN6/pUaRQS/E8Jh46As2C/")))
		assert.Nil(t, hash.Compare(ctx, []byte("test"), []byte("$md5-crypt$$whuMjZj.HMFoaTaZRRtkO0")))
		assert.Error(t, hash.Compare(ctx, []byte("test"), []byte("$md5-crypt$xWMlm2eL$GGTOpgZu4p2k6ORprAu3b.")))

		assert.Nil(t, hash.Compare(ctx, []byte("ory"), []byte("$md5-crypt$xWMlm2eL$GGTOpgZu4p2k6ORprAu3b.")))
		assert.Nil(t, hash.CompareMD5Crypt(ctx, []byte("ory"), []byte("$md5-crypt$xWMlm2eL$GGTOpgZu4p2k6ORprAu3b.")))
		assert.Error(t, hash.Compare(ctx, []byte("ory"), []byte("$md5-crypt$E7zjruqF$RTglYR1CzBHwwiTk9nVzx1")))

		assert.ErrorIs(t, hash.Compare(ctx, []byte("ory"), []byte("$md5-crypt$$")), hash.ErrMismatchedHashAndPassword)
		assert.Error(t, hash.Compare(ctx, []byte("ory"), []byte("$md5-crypt$$$")))
		// per crypt(5), `md5crypt` can be run without a salt, but the salt section must still be present
		assert.Error(t, hash.Compare(ctx, []byte("test"), []byte("$md5-crypt$whuMjZj.HMFoaTaZRRtkO0")), "md5crypt decode error: provided encoded hash has an invalid format")
	})

	t.Run("sha256-crypt", func(t *testing.T) {
		t.Parallel()

		assert.Nil(t, hash.Compare(ctx, []byte("test"), []byte("$sha256-crypt$rounds=535000$05R.9KB6UC2kLI3w$Q/zslzx./JjkAVPTwp6th7nW5l7JU91Gte/UmIh.U78")))
		assert.Nil(t, hash.CompareSHA256Crypt(ctx, []byte("test"), []byte("$sha256-crypt$rounds=535000$05R.9KB6UC2kLI3w$Q/zslzx./JjkAVPTwp6th7nW5l7JU91Gte/UmIh.U78")))
		assert.Error(t, hash.Compare(ctx, []byte("test"), []byte("$sha256-crypt$rounds=535000$awpcR7lDlnK/S7WE$vHU7KkQwyjfGz6u4MUi7.lH9htK/l63HloTsX1ZMz.3")))

		assert.Nil(t, hash.Compare(ctx, []byte("ory"), []byte("$sha256-crypt$rounds=535000$awpcR7lDlnK/S7WE$vHU7KkQwyjfGz6u4MUi7.lH9htK/l63HloTsX1ZMz.3")))
		assert.Nil(t, hash.CompareSHA256Crypt(ctx, []byte("ory"), []byte("$sha256-crypt$rounds=535000$awpcR7lDlnK/S7WE$vHU7KkQwyjfGz6u4MUi7.lH9htK/l63HloTsX1ZMz.3")))
		assert.Error(t, hash.Compare(ctx, []byte("ory"), []byte("$sha256-crypt$rounds=535000$T95kH8e37IGVdxzJ$gLeaNa6qRog.bx4Bzqp63ceWItH6nSAal6c3WmT5GHB")))

		assert.Error(t, hash.Compare(ctx, []byte("ory"), []byte("$sha256-crypt$$")), "shacrypt decode error: provided encoded hash has an invalid format")
		assert.Error(t, hash.Compare(ctx, []byte("ory"), []byte("$sha256-crypt$$$")))
	})

	t.Run("sha512-crypt", func(t *testing.T) {
		t.Parallel()

		assert.Nil(t, hash.Compare(ctx, []byte("test"), []byte("$sha512-crypt$rounds=656000$3LVbIAVxR//cRajw$uuNasMW.RYxlGzIRFU1Was70BPSa933AjxhZIGJdJBOlqJAHlgqa0yuiuq5JHF/ryNGryJkj87G9i3G2HPSXg1")))
		assert.Nil(t, hash.CompareSHA512Crypt(ctx, []byte("test"), []byte("$sha512-crypt$rounds=656000$3LVbIAVxR//cRajw$uuNasMW.RYxlGzIRFU1Was70BPSa933AjxhZIGJdJBOlqJAHlgqa0yuiuq5JHF/ryNGryJkj87G9i3G2HPSXg1")))
		assert.Error(t, hash.Compare(ctx, []byte("test"), []byte("$5$rounds=535000$awpcR7lDlnK/S7WE$vHU7KkQwyjfGz6u4MUi7.lH9htK/l63HloTsX1ZMz.3")))

		assert.Nil(t, hash.Compare(ctx, []byte("ory"), []byte("$sha512-crypt$rounds=656000$0baQbxBrfpKqvizk$Q9cYk1MeNAlECPgpG3jjfNI2DumLqd0yHbxzLdxiX6nsSD5i9n0awcbiCf8R5DzpIYxeBPznPcb1wtzlgUKtH0")))
		assert.Nil(t, hash.CompareSHA512Crypt(ctx, []byte("ory"), []byte("$sha512-crypt$rounds=656000$0baQbxBrfpKqvizk$Q9cYk1MeNAlECPgpG3jjfNI2DumLqd0yHbxzLdxiX6nsSD5i9n0awcbiCf8R5DzpIYxeBPznPcb1wtzlgUKtH0")))
		assert.Error(t, hash.Compare(ctx, []byte("ory"), []byte("$sha512-crypt$rounds=656000$hNcDLFO63bkYVDZf$Mt9dhH0xqfxWZ6Pu8zXw.Ku5f15IRTweuaDcUc.ObXWGn7B1h8YIWLmArZd8psd2mrUVswCXLAVptmISr.8iI/")))

		assert.Error(t, hash.Compare(ctx, []byte("ory"), []byte("$sha512-crypt$$")), "shacrypt decode error: provided encoded hash has an invalid format")
		assert.Error(t, hash.Compare(ctx, []byte("ory"), []byte("$sha512-crypt$$$")))
	})

	t.Run("hmac errors", func(t *testing.T) {
		t.Parallel()

		//Missing Key
		assert.ErrorIs(t, hash.Compare(ctx, []byte("test"), []byte("$hmac-md5$ZmU4Njk3Zjc0MmQwODA0MDVkMTI3MGU2MTYzMzE2Zjk=")), hash.ErrInvalidHash)
		assert.Error(t, hash.CompareHMAC(ctx, []byte("test"), []byte("$hmac-md5$ZmU4Njk3Zjc0MmQwODA0MDVkMTI3MGU2MTYzMzE2Zjk=")))
		assert.ErrorIs(t, hash.Compare(ctx, []byte("test"), []byte("$hmac-md5$ZmU4Njk3Zjc0MmQwODA0MDVkMTI3MGU2MTYzMzE2Zjk=$")), hash.ErrMismatchedHashAndPassword)
		assert.Error(t, hash.CompareHMAC(ctx, []byte("test"), []byte("$hmac-md5$ZmU4Njk3Zjc0MmQwODA0MDVkMTI3MGU2MTYzMzE2Zjk=$")))
		//Missing Password Hash
		assert.ErrorIs(t, hash.Compare(ctx, []byte("test"), []byte("$hmac-md5$MTIzNDU=")), hash.ErrInvalidHash)
		assert.Error(t, hash.CompareHMAC(ctx, []byte("test"), []byte("$hmac-md5$MTIzNDU=")))
		assert.ErrorIs(t, hash.Compare(ctx, []byte("test"), []byte("$hmac-md5$$MTIzNDU=")), hash.ErrMismatchedHashAndPassword)
		assert.Error(t, hash.CompareHMAC(ctx, []byte("test"), []byte("$hmac-md5$$MTIzNDU=")))
		//Missing Password Hash and Key
		assert.ErrorIs(t, hash.Compare(ctx, []byte("test"), []byte("$hmac-md5$")), hash.ErrInvalidHash)
		assert.Error(t, hash.CompareHMAC(ctx, []byte("test"), []byte("$hmac-md5$")))
		//Missing Hash Algorithm
		assert.ErrorIs(t, hash.Compare(ctx, []byte("test"), []byte("$hmac$ZmU4Njk3Zjc0MmQwODA0MDVkMTI3MGU2MTYzMzE2Zjk=$MTIzNDU=")), hash.ErrUnknownHashAlgorithm)
		assert.Error(t, hash.CompareHMAC(ctx, []byte("test"), []byte("$hmac$ZmU4Njk3Zjc0MmQwODA0MDVkMTI3MGU2MTYzMzE2Zjk=$MTIzNDU=")))
		//Missing Invalid Hash Algorithm
		assert.ErrorIs(t, hash.Compare(ctx, []byte("test"), []byte("$hmac-invalid$ZmU4Njk3Zjc0MmQwODA0MDVkMTI3MGU2MTYzMzE2Zjk=$MTIzNDU=")), hash.ErrUnknownHashAlgorithm)
		assert.Error(t, hash.CompareHMAC(ctx, []byte("test"), []byte("$hmac-invalid$ZmU4Njk3Zjc0MmQwODA0MDVkMTI3MGU2MTYzMzE2Zjk=$MTIzNDU=")))

	})

	t.Run("hmac-md4", func(t *testing.T) {
		t.Parallel()

		//Valid
		assert.Nil(t, hash.Compare(ctx, []byte("test"), []byte("$hmac-md4$MWQ5ZTI4Nzc2Zjg4YmE2MTQ5YjQ0OTMyOGE4NWU4YjA=$MTIzNDU=")))
		assert.Nil(t, hash.CompareHMAC(ctx, []byte("test"), []byte("$hmac-md4$MWQ5ZTI4Nzc2Zjg4YmE2MTQ5YjQ0OTMyOGE4NWU4YjA=$MTIzNDU=")))
		//Wrong Key
		assert.Error(t, hash.Compare(ctx, []byte("test"), []byte("$hmac-md4$MWQ5ZTI4Nzc2Zjg4YmE2MTQ5YjQ0OTMyOGE4NWU4YjA=$MTIzNA==")), hash.ErrMismatchedHashAndPassword)
		//Different password
		assert.Error(t, hash.Compare(ctx, []byte("ory"), []byte("$hmac-md4$MWQ5ZTI4Nzc2Zjg4YmE2MTQ5YjQ0OTMyOGE4NWU4YjA=$MTIzNDU=")))
	})

	t.Run("hmac-md5", func(t *testing.T) {
		t.Parallel()

		//Valid
		assert.Nil(t, hash.Compare(ctx, []byte("test"), []byte("$hmac-md5$ZmU4Njk3Zjc0MmQwODA0MDVkMTI3MGU2MTYzMzE2Zjk=$MTIzNDU=")))
		assert.Nil(t, hash.CompareHMAC(ctx, []byte("test"), []byte("$hmac-md5$ZmU4Njk3Zjc0MmQwODA0MDVkMTI3MGU2MTYzMzE2Zjk=$MTIzNDU=")))

		//Wrong Key
		assert.Error(t, hash.Compare(ctx, []byte("test"), []byte("$hmac-md5$ZmU4Njk3Zjc0MmQwODA0MDVkMTI3MGU2MTYzMzE2Zjk=$MTIzNA==")))
		//Different password
		assert.Error(t, hash.Compare(ctx, []byte("ory"), []byte("$hmac-md5$ZmU4Njk3Zjc0MmQwODA0MDVkMTI3MGU2MTYzMzE2Zjk=$MTIzNDU=")))

	})

	t.Run("hmac-sha1", func(t *testing.T) {
		t.Parallel()

		//Valid
		assert.Nil(t, hash.Compare(ctx, []byte("test"), []byte("$hmac-sha1$NDMyNjcxZTUyY2Y2YTBmYjZjZDE2NjQxYjAwNjFiZjAwOGEzNWM5MA==$MTIzNDU=")))
		assert.Nil(t, hash.CompareHMAC(ctx, []byte("test"), []byte("$hmac-sha1$NDMyNjcxZTUyY2Y2YTBmYjZjZDE2NjQxYjAwNjFiZjAwOGEzNWM5MA==$MTIzNDU=")))

		//Wrong Key
		assert.Error(t, hash.Compare(ctx, []byte("test"), []byte("$hmac-sha1$NDMyNjcxZTUyY2Y2YTBmYjZjZDE2NjQxYjAwNjFiZjAwOGEzNWM5MA==$MTIzNA==")))
		//Different password
		assert.Error(t, hash.Compare(ctx, []byte("ory"), []byte("$hmac-sha1$NDMyNjcxZTUyY2Y2YTBmYjZjZDE2NjQxYjAwNjFiZjAwOGEzNWM5MA==$MTIzNDU=")))

	})

	t.Run("hmac-sha224", func(t *testing.T) {
		t.Parallel()

		//Valid
		assert.Nil(t, hash.Compare(ctx, []byte("test"), []byte("$hmac-sha224$YmUwYmYzM2EwNGRlNDE0YjQzNjBhNmIyOThmNmIyYzI4OWQyMzk3MDUwZDFjMzliYjVmMDMyOTQ=$MTIzNDU=")))
		assert.Nil(t, hash.CompareHMAC(ctx, []byte("test"), []byte("$hmac-sha224$YmUwYmYzM2EwNGRlNDE0YjQzNjBhNmIyOThmNmIyYzI4OWQyMzk3MDUwZDFjMzliYjVmMDMyOTQ=$MTIzNDU=")))

		//Wrong Key
		assert.Error(t, hash.Compare(ctx, []byte("test"), []byte("$hmac-sha224$YmUwYmYzM2EwNGRlNDE0YjQzNjBhNmIyOThmNmIyYzI4OWQyMzk3MDUwZDFjMzliYjVmMDMyOTQ=$MTIzNA==")))
		//Different password
		assert.Error(t, hash.Compare(ctx, []byte("ory"), []byte("$hmac-sha224$YmUwYmYzM2EwNGRlNDE0YjQzNjBhNmIyOThmNmIyYzI4OWQyMzk3MDUwZDFjMzliYjVmMDMyOTQ=$MTIzNDU=")))

	})

	t.Run("hmac-sha256", func(t *testing.T) {
		t.Parallel()

		//Valid
		assert.Nil(t, hash.Compare(ctx, []byte("test"), []byte("$hmac-sha256$ZTAzMWJhMWMyOTM4YjFkMjgzZjkxOWExZGY5YWM2NmMxOTJhN2RkNzQ0MzJkNWZkNGFkYTI5OTk0MWJhMTA5Zg==$MTIzNDU=")))
		assert.Nil(t, hash.CompareHMAC(ctx, []byte("test"), []byte("$hmac-sha256$ZTAzMWJhMWMyOTM4YjFkMjgzZjkxOWExZGY5YWM2NmMxOTJhN2RkNzQ0MzJkNWZkNGFkYTI5OTk0MWJhMTA5Zg==$MTIzNDU=")))

		//Wrong Key
		assert.Error(t, hash.Compare(ctx, []byte("test"), []byte("$hmac-sha256$ZTAzMWJhMWMyOTM4YjFkMjgzZjkxOWExZGY5YWM2NmMxOTJhN2RkNzQ0MzJkNWZkNGFkYTI5OTk0MWJhMTA5Zg==$MTIzNA==")))
		//Different password
		assert.Error(t, hash.Compare(ctx, []byte("ory"), []byte("$hmac-sha256$ZTAzMWJhMWMyOTM4YjFkMjgzZjkxOWExZGY5YWM2NmMxOTJhN2RkNzQ0MzJkNWZkNGFkYTI5OTk0MWJhMTA5Zg==$MTIzNDU=")))

	})

	t.Run("hmac-sha384", func(t *testing.T) {
		t.Parallel()

		//Valid
		assert.Nil(t, hash.Compare(ctx, []byte("test"), []byte("$hmac-sha384$ZWEyMGM3NGE4Y2UzMTljNTdjZTlhZGQyYTZjNDE0MGQ4YjMwYWIwOWM4OTRiNWQ4MmZjODlhMzBhMmQzNGE5NmQ0NDY1NWRhYjQ2ZjhiYjBkNTRmYjk5YWZkZTA1MGY1$MTIzNDU=")))
		assert.Nil(t, hash.CompareHMAC(ctx, []byte("test"), []byte("$hmac-sha384$ZWEyMGM3NGE4Y2UzMTljNTdjZTlhZGQyYTZjNDE0MGQ4YjMwYWIwOWM4OTRiNWQ4MmZjODlhMzBhMmQzNGE5NmQ0NDY1NWRhYjQ2ZjhiYjBkNTRmYjk5YWZkZTA1MGY1$MTIzNDU=")))

		//Wrong Key
		assert.Error(t, hash.Compare(ctx, []byte("test"), []byte("$hmac-sha384$ZWEyMGM3NGE4Y2UzMTljNTdjZTlhZGQyYTZjNDE0MGQ4YjMwYWIwOWM4OTRiNWQ4MmZjODlhMzBhMmQzNGE5NmQ0NDY1NWRhYjQ2ZjhiYjBkNTRmYjk5YWZkZTA1MGY1$MTIzNA==")))
		//Different password
		assert.Error(t, hash.Compare(ctx, []byte("ory"), []byte("$hmac-sha384$ZWEyMGM3NGE4Y2UzMTljNTdjZTlhZGQyYTZjNDE0MGQ4YjMwYWIwOWM4OTRiNWQ4MmZjODlhMzBhMmQzNGE5NmQ0NDY1NWRhYjQ2ZjhiYjBkNTRmYjk5YWZkZTA1MGY1$MTIzNDU=")))

	})

	t.Run("hmac-sha512", func(t *testing.T) {
		t.Parallel()

		//Valid
		assert.Nil(t, hash.Compare(ctx, []byte("test"), []byte("$hmac-sha512$OTFmODY0ZTI1NmU0ZjVhYjhiMDViZGFmNGVmNGZmMGVlNTY4ODYwNWJhYTk4MTk2OTgyMzc3NzI1YTc4MzcxMTMzNzZmY2YxYTk5MGMxM2RiZDk2MGFmMmQ1YzRmODdlMGMwYTNkYjcyNjY0NjM4NGE4YzQ2MjNhZDZkN2UxZTE=$MTIzNDU=")))
		assert.Nil(t, hash.CompareHMAC(ctx, []byte("test"), []byte("$hmac-sha512$OTFmODY0ZTI1NmU0ZjVhYjhiMDViZGFmNGVmNGZmMGVlNTY4ODYwNWJhYTk4MTk2OTgyMzc3NzI1YTc4MzcxMTMzNzZmY2YxYTk5MGMxM2RiZDk2MGFmMmQ1YzRmODdlMGMwYTNkYjcyNjY0NjM4NGE4YzQ2MjNhZDZkN2UxZTE=$MTIzNDU=")))

		//Wrong Key
		assert.Error(t, hash.Compare(ctx, []byte("test"), []byte("$hmac-sha512$OTFmODY0ZTI1NmU0ZjVhYjhiMDViZGFmNGVmNGZmMGVlNTY4ODYwNWJhYTk4MTk2OTgyMzc3NzI1YTc4MzcxMTMzNzZmY2YxYTk5MGMxM2RiZDk2MGFmMmQ1YzRmODdlMGMwYTNkYjcyNjY0NjM4NGE4YzQ2MjNhZDZkN2UxZTE=$MTIzNA==")))
		//Different password
		assert.Error(t, hash.Compare(ctx, []byte("ory"), []byte("$hmac-sha512$OTFmODY0ZTI1NmU0ZjVhYjhiMDViZGFmNGVmNGZmMGVlNTY4ODYwNWJhYTk4MTk2OTgyMzc3NzI1YTc4MzcxMTMzNzZmY2YxYTk5MGMxM2RiZDk2MGFmMmQ1YzRmODdlMGMwYTNkYjcyNjY0NjM4NGE4YzQ2MjNhZDZkN2UxZTE=$MTIzNDU=")))

	})
}
