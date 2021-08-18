package cipher_test

import (
	"context"
	"encoding/hex"
	"testing"

	"github.com/ory/herodot"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/cipher"
	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/internal"
)

func TestAES_Cipher(t *testing.T) {
	cfg, reg := internal.NewFastRegistryWithMocks(t)
	aes := cipher.NewCryptAES(reg)
	goodSecret := []string{"secret-thirty-two-character-long"}

	t.Run("case=all_work", func(t *testing.T) {
		secret := "sc8cIZp7RlNDEc2Qg8isN"

		encryptedSecret, err := aes.Encrypt(context.Background(), secret)
		require.NoError(t, err)

		decryptedSecret, err := aes.Decrypt(context.Background(), encryptedSecret)
		require.NoError(t, err, "encrypted", encryptedSecret)
		assert.Equal(t, secret, decryptedSecret)

		// data to encrypt return blank result
		_, err = aes.Encrypt(context.Background(), "")
		require.NoError(t, err)

		// empty encrypted data return blank
		_, err = aes.Decrypt(context.Background(), "")
		require.NoError(t, err)

	})
	t.Run("case=encryption_failed", func(t *testing.T) {
		// unset secret
		err := cfg.Set(config.ViperKeySecretsCipher, []string{})
		require.NoError(t, err)

		// secret have to be set
		_, err = aes.Encrypt(context.Background(), "not-empty")
		require.Error(t, err)

		// unset secret
		err = cfg.Set(config.ViperKeySecretsCipher, []string{"bad-length"})
		require.NoError(t, err)

		// bad secret length
		_, err = aes.Encrypt(context.Background(), "not-empty")
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
		_, err = aes.Decrypt(context.Background(), hex.EncodeToString([]byte("bad-data")))
		require.Error(t, err)

		_, err = aes.Decrypt(context.Background(), "not-empty")
		require.Error(t, err)

		// unset secret
		err = cfg.Set(config.ViperKeySecretsCipher, []string{})
		require.NoError(t, err)

		_, err = aes.Decrypt(context.Background(), "not-empty")
		require.Error(t, err)
	})
}
