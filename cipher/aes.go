// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package cipher

import (
	"context"
	"crypto/aes"
	stdcipher "crypto/cipher"
	"encoding/hex"

	"github.com/pkg/errors"

	"github.com/ory/herodot"
)

type AES struct {
	c SecretsProvider
}

func NewCryptAES(c SecretsProvider) *AES {
	return &AES{c: c}
}

// aesGCM returns an AES-256-GCM AEAD that generates a random nonce internally
// and prepends it to the ciphertext (nonce | ciphertext | tag).
func aesGCM(key *[32]byte) (stdcipher.AEAD, error) {
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, errors.WithStack(err)
	}
	gcm, err := stdcipher.NewGCMWithRandomNonce(block)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return gcm, nil
}

// Encrypt returns the AES-256-GCM encryption of plaintext.
func (a *AES) Encrypt(ctx context.Context, message []byte) (string, error) {
	if len(message) == 0 {
		return "", nil
	}

	if len(a.c.SecretsCipher(ctx)) == 0 {
		return "", errors.WithStack(herodot.ErrMisconfiguration().WithReason("Unable to encrypt message because no cipher secrets were configured."))
	}

	gcm, err := aesGCM(&a.c.SecretsCipher(ctx)[0])
	if err != nil {
		return "", errors.WithStack(herodot.ErrForbidden().WithWrap(err))
	}

	return hex.EncodeToString(gcm.Seal(nil, nil, message, nil)), nil
}

// Decrypt returns the decrypted AES data.
func (a *AES) Decrypt(ctx context.Context, ciphertext string) ([]byte, error) {
	if len(ciphertext) == 0 {
		return nil, nil
	}

	secrets := a.c.SecretsCipher(ctx)
	if len(secrets) == 0 {
		return nil, errors.WithStack(herodot.ErrMisconfiguration().WithReason("Unable to decipher the encrypted message because no AES secrets were configured."))
	}

	decode, err := hex.DecodeString(ciphertext)
	if err != nil {
		return nil, errors.WithStack(herodot.ErrBadRequest().WithWrap(err))
	}

	for i := range secrets {
		gcm, err := aesGCM(&secrets[i])
		if err != nil {
			continue
		}
		if plaintext, err := gcm.Open(nil, nil, decode, nil); err == nil {
			return plaintext, nil
		}
	}

	return nil, errors.WithStack(herodot.ErrForbidden().WithReason("Unable to decipher the encrypted message."))
}
