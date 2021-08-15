/*
Copyright 2020 WILDCARD SA.

Licensed under the WILDCARD SA License, Version 1.0 (the "License");
WILDCARD SA is register in french corporation.
You may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.w6d.io/licenses/LICENSE-1.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is prohibited.
Created on 13/08/2021
*/

package cipher_test

import (
	"context"
	"testing"

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
	})
	t.Run("case=encrypt_failed", func(t *testing.T) {
		// set secret
		err := cfg.Set(config.ViperKeySecretsAES, goodSecret)
		require.NoError(t, err)

		// data to encrypt can not be empty
		_, err = aes.Encrypt(context.Background(), "")
		require.Error(t, err)

		// unset secret
		err = cfg.Set(config.ViperKeySecretsAES, []string{})
		require.NoError(t, err)

		// secret have to be set
		_, err = aes.Encrypt(context.Background(), "not-empty")
		require.Error(t, err)

	})
	t.Run("case=decrypt_failed", func(t *testing.T) {
		// set secret
		err := cfg.Set(config.ViperKeySecretsAES, goodSecret)
		require.NoError(t, err)

		// encrypted data could not be empty
		_, err = aes.Decrypt(context.Background(), []byte{})
		require.Error(t, err)

		//
		_, err = aes.Decrypt(context.Background(), []byte("bad-data"))
		require.Error(t, err)

		// unset secret
		err = cfg.Set(config.ViperKeySecretsAES, []string{})
		require.NoError(t, err)

		_, err = aes.Decrypt(context.Background(), []byte("not-empty"))
		require.Error(t, err)
	})
}
