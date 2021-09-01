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
	"github.com/ory/kratos/internal"
)

func TestCipher(t *testing.T) {
	cfg, reg := internal.NewFastRegistryWithMocks(t)
	goodSecret := []string{"secret-thirty-two-character-long"}

	ciphers := []cipher.Cipher{
		cipher.NewCryptAES(reg),
		cipher.NewCryptChaCha20(reg),
	}

	for _, c := range ciphers {
		t.Run(fmt.Sprintf("cipher=%T", c), func(t *testing.T) {

			t.Run("case=all_work", func(t *testing.T) {
				cfg.MustSet(config.ViperKeySecretsCipher, goodSecret)
				testAllWork(t, c, cfg)
			})

			t.Run("case=encryption_failed", func(t *testing.T) {
				// unset secret
				err := cfg.Set(config.ViperKeySecretsCipher, []string{})
				require.NoError(t, err)

				// secret have to be set
				_, err = c.Encrypt(context.Background(), []byte("not-empty"))
				require.Error(t, err)

				// unset secret
				err = cfg.Set(config.ViperKeySecretsCipher, []string{"bad-length"})
				require.NoError(t, err)

				// bad secret length
				_, err = c.Encrypt(context.Background(), []byte("not-empty"))
				if e, ok := err.(*herodot.DefaultError); ok {
					t.Logf("reason contains: %s", e.Reason())
				}
				t.Logf("err type %T contains: %s", err, err.Error())
				require.Error(t, err)
			})

			t.Run("case=decryption_failed", func(t *testing.T) {
				// set secret
				err := cfg.Set(config.ViperKeySecretsCipher, goodSecret)
				require.NoError(t, err)

				//
				_, err = c.Decrypt(context.Background(), hex.EncodeToString([]byte("bad-data")))
				require.Error(t, err)

				_, err = c.Decrypt(context.Background(), "not-empty")
				require.Error(t, err)

				// unset secret
				err = cfg.Set(config.ViperKeySecretsCipher, []string{})
				require.NoError(t, err)

				_, err = c.Decrypt(context.Background(), "not-empty")
				require.Error(t, err)
			})
		})
	}
	c := cipher.NewNoop(reg)
	t.Run(fmt.Sprintf("cipher=%T", c), func(t *testing.T) {
		cfg.MustSet(config.ViperKeySecretsCipher, goodSecret)
		testAllWork(t, c, cfg)
	})
}

func testAllWork(t *testing.T, c cipher.Cipher, cfg *config.Config) {
	goodSecret := []string{"secret-thirty-two-character-long"}
	cfg.MustSet(config.ViperKeySecretsCipher, goodSecret)

	message := "my secret message!"

	encryptedSecret, err := c.Encrypt(context.Background(), []byte(message))
	require.NoError(t, err)

	decryptedSecret, err := c.Decrypt(context.Background(), encryptedSecret)
	require.NoError(t, err, "encrypted", encryptedSecret)
	assert.Equal(t, message, string(decryptedSecret))

	// data to encrypt return blank result
	_, err = c.Encrypt(context.Background(), []byte(""))
	require.NoError(t, err)

	// empty encrypted data return blank
	_, err = c.Decrypt(context.Background(), "")
	require.NoError(t, err)
}
