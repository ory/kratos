// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package cipher

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"io"

	"github.com/pkg/errors"
	"golang.org/x/crypto/chacha20poly1305"

	"github.com/ory/herodot"
	"github.com/ory/kratos/driver/config"
)

type ChaCha20Configuration interface {
	config.Provider
}

type XChaCha20Poly1305 struct {
	c ChaCha20Configuration
}

func NewCryptChaCha20(c ChaCha20Configuration) *XChaCha20Poly1305 {
	return &XChaCha20Poly1305{c: c}
}

// Encrypt returns a ChaCha encryption of plaintext
func (c *XChaCha20Poly1305) Encrypt(ctx context.Context, message []byte) (string, error) {
	if len(message) == 0 {
		return "", nil
	}

	if len(c.c.Config().SecretsCipher(ctx)) == 0 {
		return "", errors.WithStack(herodot.ErrInternalServerError.WithReason("Unable to encrypt message because no cipher secrets were configured."))
	}

	aead, err := chacha20poly1305.NewX(c.c.Config().SecretsCipher(ctx)[0][:])
	if err != nil {
		return "", herodot.ErrInternalServerError.WithWrap(err).WithReason("Unable to generate key")
	}

	nonce := make([]byte, aead.NonceSize(), aead.NonceSize()+len(message)+aead.Overhead())
	_, err = io.ReadFull(rand.Reader, nonce)
	if err != nil {
		return "", errors.WithStack(herodot.ErrInternalServerError.WithWrap(err).WithReason("Unable to generate nonce"))
	}

	encryptedMsg := aead.Seal(nonce, nonce, message, nil)
	return hex.EncodeToString(encryptedMsg), nil
}

// Decrypt decrypts data using 256 bit key
func (c *XChaCha20Poly1305) Decrypt(ctx context.Context, ciphertext string) ([]byte, error) {
	if len(ciphertext) == 0 {
		return nil, nil
	}

	secrets := c.c.Config().SecretsCipher(ctx)
	if len(secrets) == 0 {
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithReason("Unable to decipher the encrypted message because no cipher secrets were configured."))
	}

	rawCiphertext, err := hex.DecodeString(ciphertext)
	if err != nil {
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithWrap(err).WithReason("Unable to decode hex encrypted string"))
	}

	for i := range secrets {
		aead, err := chacha20poly1305.NewX(secrets[i][:])
		if err != nil {
			return nil, errors.WithStack(herodot.ErrInternalServerError.WithWrap(err).WithReason("Unable to instanciate chacha20"))
		}

		if len(ciphertext) < aead.NonceSize() {
			return nil, errors.WithStack(herodot.ErrInternalServerError.WithReason("cipher text too short"))
		}

		nonce, ciphertext := rawCiphertext[:aead.NonceSize()], rawCiphertext[aead.NonceSize():]
		plaintext, err := aead.Open(nil, nonce, ciphertext, nil)
		if err == nil {
			return plaintext, nil
		}
	}

	return nil, errors.WithStack(herodot.ErrInternalServerError.WithReason("Unable to decrypt string"))
}
