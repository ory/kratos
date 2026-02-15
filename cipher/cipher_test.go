// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package cipher_test

import (
	"context"
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/herodot"
	"github.com/ory/kratos/cipher"
	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/pkg"
	"github.com/ory/x/configx"
	"github.com/ory/x/contextx"
)

var goodSecret = []string{"secret-thirty-two-character-long"}

func TestCipher(t *testing.T) {
	ctx := context.Background()
	_, reg := pkg.NewFastRegistryWithMocks(t, configx.WithValue(config.ViperKeySecretsDefault, goodSecret))

	ciphers := []cipher.Cipher{
		cipher.NewCryptAES(reg.Config()),
		cipher.NewCryptChaCha20(reg.Config()),
	}

	for _, c := range ciphers {
		t.Run(fmt.Sprintf("cipher=%T", c), func(t *testing.T) {
			t.Parallel()

			t.Run("case=all_work", func(t *testing.T) {
				t.Parallel()

				testAllWork(ctx, t, c)
			})

			t.Run("case=encryption_failed", func(t *testing.T) {
				t.Parallel()

				ctx := contextx.WithConfigValue(ctx, config.ViperKeySecretsCipher, []string{""})

				// secret have to be set
				_, err := c.Encrypt(ctx, []byte("not-empty"))
				require.Error(t, err)
				var hErr *herodot.DefaultError
				require.ErrorAs(t, err, &hErr)
				assert.Equal(t, "Unable to encrypt message because no cipher secrets were configured.", hErr.Reason())

				ctx = contextx.WithConfigValue(ctx, config.ViperKeySecretsCipher, []string{"bad-length"})

				// bad secret length
				_, err = c.Encrypt(ctx, []byte("not-empty"))
				require.ErrorAs(t, err, &hErr)
				assert.Equal(t, "Unable to encrypt message because no cipher secrets were configured.", hErr.Reason())
			})

			t.Run("case=decryption_failed", func(t *testing.T) {
				t.Parallel()

				_, err := c.Decrypt(ctx, hex.EncodeToString([]byte("bad-data")))
				require.Error(t, err)

				_, err = c.Decrypt(ctx, "not-empty")
				require.Error(t, err)

				_, err = c.Decrypt(contextx.WithConfigValue(ctx, config.ViperKeySecretsCipher, []string{""}), "not-empty")
				require.Error(t, err)
			})

			t.Run("case=short_ciphertext", func(t *testing.T) {
				t.Parallel()

				// XChaCha20-Poly1305 has 24-byte nonce, hex encoded is 48 chars
				// A valid ciphertext needs at least 24 bytes (nonce) + 16 bytes (tag) = 40 bytes minimum
				// Hex encoded minimum is 80 chars
				// This tests that we don't get panic on short ciphertext

				// 24 hex chars is only 12 bytes - less than nonce size (24 bytes)
				shortCiphertext := "00112233445566778899aabbccddeeff"
				_, err := c.Decrypt(ctx, shortCiphertext)
				require.Error(t, err)

				// 64 hex chars is 32 bytes - still less than nonce(24)+tag(16)=40
				mediumCiphertext := "00112233445566778899aabbccddeeff00112233445566778899aabbccddeeff"
				_, err = c.Decrypt(ctx, mediumCiphertext)
				require.Error(t, err)
			})
		})
	}

	c := cipher.NewNoop()
	t.Run(fmt.Sprintf("cipher=%T", c), func(t *testing.T) {
		t.Parallel()
		testAllWork(ctx, t, c)
	})
}

func testAllWork(ctx context.Context, t *testing.T, c cipher.Cipher) {
	message := "my secret message!"

	encryptedSecret, err := c.Encrypt(ctx, []byte(message))
	require.NoError(t, err)

	decryptedSecret, err := c.Decrypt(ctx, encryptedSecret)
	require.NoError(t, err, "encrypted", encryptedSecret)
	assert.Equal(t, message, string(decryptedSecret))

	// data to encrypt return blank result
	_, err = c.Encrypt(ctx, []byte(""))
	require.NoError(t, err)

	// empty encrypted data return blank
	_, err = c.Decrypt(ctx, "")
	require.NoError(t, err)
}
