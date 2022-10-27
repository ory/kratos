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

	assert.Error(t, hash.Compare(context.Background(), []byte("test"), []byte("$scrypt$2npRo7P03Mt8keSoMbyD/tKFWyUzjiQf2svUaNDSrhA=$MiCzNcIplSMqSBrm4HckjYqYhaVPPjTARTzwB1cVNYE=")))
	assert.Error(t, hash.Compare(context.Background(), []byte("test"), []byte("$scrypt$ln=16384,r=8,p=1$(2npRo7P03Mt8keSoMbyD/tKFWyUzjiQf2svUaNDSrhA=$MiCzNcIplSMqSBrm4HckjYqYhaVPPjTARTzwB1cVNYE=")))
	assert.Error(t, hash.Compare(context.Background(), []byte("test"), []byte("$scrypt$ln=16384,r=8,p=1$(2npRo7P03Mt8keSoMbyD/tKFWyUzjiQf2svUaNDSrhA=$(MiCzNcIplSMqSBrm4HckjYqYhaVPPjTARTzwB1cVNYE=")))
	assert.Error(t, hash.Compare(context.Background(), []byte("test"), []byte("$scrypt$ln=16385,r=8,p=1$2npRo7P03Mt8keSoMbyD/tKFWyUzjiQf2svUaNDSrhA=$MiCzNcIplSMqSBrm4HckjYqYhaVPPjTARTzwB1cVNYE=")))
	assert.Error(t, hash.Compare(context.Background(), []byte("tesu"), []byte("$scrypt$ln=16384,r=8,p=1$2npRo7P03Mt8keSoMbyD/tKFWyUzjiQf2svUaNDSrhA=$MiCzNcIplSMqSBrm4HckjYqYhaVPPjTARTzwB1cVNYE=")))
	assert.Error(t, hash.Compare(context.Background(), []byte("tesu"), []byte("$scrypt$ln=abc,r=8,p=1$2npRo7P03Mt8keSoMbyD/tKFWyUzjiQf2svUaNDSrhA=$MiCzNcIplSMqSBrm4HckjYqYhaVPPjTARTzwB1cVNYE=")))

	assert.Nil(t, hash.Compare(context.Background(), []byte("test123"), []byte("{SSHA}JFZFs0oHzxbMwkSJmYVeI8MnTDy/276a")))
	assert.Nil(t, hash.CompareSSHA(context.Background(), []byte("test123"), []byte("{SSHA}JFZFs0oHzxbMwkSJmYVeI8MnTDy/276a")))
	assert.Error(t, hash.CompareSSHA(context.Background(), []byte("badtest"), []byte("{SSHA}JFZFs0oHzxbMwkSJmYVeI8MnTDy/276a")))

	assert.Nil(t, hash.Compare(context.Background(), []byte("test123"), []byte("{SSHA256}czO44OTV17PcF1cRxWrLZLy9xHd7CWyVYplr1rOhuMlx/7IK")))
	assert.Nil(t, hash.CompareSSHA256(context.Background(), []byte("test123"), []byte("{SSHA256}czO44OTV17PcF1cRxWrLZLy9xHd7CWyVYplr1rOhuMlx/7IK")))
	assert.Error(t, hash.CompareSSHA256(context.Background(), []byte("badtest"), []byte("{SSHA256}czO44OTV17PcF1cRxWrLZLy9xHd7CWyVYplr1rOhuMlx/7IK")))

	assert.Nil(t, hash.Compare(context.Background(), []byte("test123"), []byte("{SSHA512}xPUl/px+1cG55rUH4rzcwxdOIPSB2TingLpiJJumN2xyDWN4Ix1WQG3ihnvHaWUE8MYNkvMi5rf0C9NYixHsE6Yh59M=")))
	assert.Nil(t, hash.CompareSSHA512(context.Background(), []byte("test123"), []byte("{SSHA512}xPUl/px+1cG55rUH4rzcwxdOIPSB2TingLpiJJumN2xyDWN4Ix1WQG3ihnvHaWUE8MYNkvMi5rf0C9NYixHsE6Yh59M=")))
	assert.Error(t, hash.CompareSSHA512(context.Background(), []byte("badtest"), []byte("{SSHA512}xPUl/px+1cG55rUH4rzcwxdOIPSB2TingLpiJJumN2xyDWN4Ix1WQG3ihnvHaWUE8MYNkvMi5rf0C9NYixHsE6Yh59M=")))

	// pf   {PASSWORD}
	// pass test
	// hash a94a8fe5ccb19ba61c4c0873d391e987982fbbd3
	// salt 5opmkgz03r
	assert.Nil(t, hash.Compare(context.Background(), []byte("test"), []byte("$sha1$pf=e1BBU1NXT1JEfQ==$NW9wbWtnejAzcg==$YTk0YThmZTVjY2IxOWJhNjFjNGMwODczZDM5MWU5ODc5ODJmYmJkMw==")))
	assert.Nil(t, hash.CompareSHA1(context.Background(), []byte("test"), []byte("$sha1$pf=e1BBU1NXT1JEfQ==$NW9wbWtnejAzcg==$YTk0YThmZTVjY2IxOWJhNjFjNGMwODczZDM5MWU5ODc5ODJmYmJkMw==")))

	// pf   {PASSWORD}{SALT}
	// pass test
	// hash af487fbccf86f4b2acd813881d17de88312f55e4
	// salt 5opmkgz03r
	assert.Nil(t, hash.Compare(context.Background(), []byte("test"), []byte("$sha1$pf=e1BBU1NXT1JEfXtTQUxUfQ==$NW9wbWtnejAzcg==$YWY0ODdmYmNjZjg2ZjRiMmFjZDgxMzg4MWQxN2RlODgzMTJmNTVlNA==")))

	// $sha1$pf=<salting-format>$<salt>$<hash>
	// pf   ??staticPrefix??{SALT}{PASSWORD}
	// pass test
	// hash 4800313149fb8f17244179019ac545d271f0aaca
	// salt 5opmkgz03r
	assert.Nil(t, hash.Compare(context.Background(), []byte("test"), []byte("$sha1$pf=Pz9zdGF0aWNQcmVmaXg/P3tTQUxUfXtQQVNTV09SRH0=$NW9wbWtnejAzcg==$NDgwMDMxMzE0OWZiOGYxNzI0NDE3OTAxOWFjNTQ1ZDI3MWYwYWFjYQ==")))

	// wrong password
	assert.Error(t, hash.Compare(context.Background(), []byte("test2"), []byte("$sha1$pf=e1BBU1NXT1JEfXtTQUxUfQ==$NW9wbWtnejAzcg==$YWY0ODdmYmNjZjg2ZjRiMmFjZDgxMzg4MWQxN2RlODgzMTJmNTVlNA==")))

	// wrong salt
	assert.Error(t, hash.Compare(context.Background(), []byte("test2"), []byte("$sha1$pf=e1BBU1NXT1JEfXtTQUxUfQ==$cDJvb3ZrZGJ6cQ==$YWY0ODdmYmNjZjg2ZjRiMmFjZDgxMzg4MWQxN2RlODgzMTJmNTVlNA==")))

	// salt not b64 encoded
	assert.Error(t, hash.Compare(context.Background(), []byte("test2"), []byte("$sha1$pf=e1BBU1NXT1JEfXtTQUxUfQ==$NW9wbWtnejAzcg==$YWY0ODdmYmNjZjg2ZjRiMmFjZDgxMzg4MWQxN2RlODgzMTJmNTVlNA==")))

	// no format string
	assert.Error(t, hash.Compare(context.Background(), []byte("test2"), []byte("$sha1$pf=$NW9wbWtnejAzcg==$YWY0ODdmYmNjZjg2ZjRiMmFjZDgxMzg4MWQxN2RlODgzMTJmNTVlNA==")))
	assert.Error(t, hash.Compare(context.Background(), []byte("test2"), []byte("$sha1$$NW9wbWtnejAzcg==$YWY0ODdmYmNjZjg2ZjRiMmFjZDgxMzg4MWQxN2RlODgzMTJmNTVlNA==")))

	// wrong number of parameters
	assert.Error(t, hash.Compare(context.Background(), []byte("test2"), []byte("$sha1$NW9wbWtnejAzcg==$YWY0ODdmYmNjZjg2ZjRiMmFjZDgxMzg4MWQxN2RlODgzMTJmNTVlNA==")))

	// pf   {PASSWORD}%%{SALT}
	assert.Nil(t, hash.Compare(context.Background(), []byte("test"), []byte("$sha1$pf=e1BBU1NXT1JEfSUle1NBTFR9$NW9wbWtnejAzcg==$NjE3ZDAwNWJjZmNjNWI5YTIzNTI1OWYzNGRhNDc4ZGVlNzA3MGE4OA==")))

	// TODO: format strings with $ fail for some reason
	// pf   ${PASSWORD}{SALT}$
	assert.Nil(t, hash.Compare(context.Background(), []byte("test"), []byte("$sha1$pf=JHtQQVNTV09SRH17U0FMVH0k$NW9wbWtnejAzcg==$M2NkZjI5MzZkYTJmYzU1NmJmYTUzM2FiMWViNTljZTcxMGFjODBlNQ==")))

}
