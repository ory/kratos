// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package cipher_test

import (
	"context"
	"encoding/hex"
	"fmt"
	"testing"

	confighelpers "github.com/ory/kratos/driver/config/testhelpers"

	"github.com/ory/x/configx"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/herodot"
	"github.com/ory/kratos/cipher"
	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/internal"
)

var goodSecret = []string{"secret-thirty-two-character-long"}

func TestCipher(t *testing.T) {
	ctx := context.Background()
	_, reg := internal.NewFastRegistryWithMocks(t, configx.WithValue(config.ViperKeySecretsDefault, goodSecret))

	ciphers := []cipher.Cipher{
		cipher.NewCryptAES(reg),
		cipher.NewCryptChaCha20(reg),
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

				ctx := confighelpers.WithConfigValue(ctx, config.ViperKeySecretsCipher, []string{""})

				// secret have to be set
				_, err := c.Encrypt(ctx, []byte("not-empty"))
				require.Error(t, err)
				var hErr *herodot.DefaultError
				require.ErrorAs(t, err, &hErr)
				assert.Equal(t, "Unable to encrypt message because no cipher secrets were configured.", hErr.Reason())

				ctx = confighelpers.WithConfigValue(ctx, config.ViperKeySecretsCipher, []string{"bad-length"})

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

				_, err = c.Decrypt(confighelpers.WithConfigValue(ctx, config.ViperKeySecretsCipher, []string{""}), "not-empty")
				require.Error(t, err)
			})
		})
	}

	c := cipher.NewNoop(reg)
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
