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

type ChaCha20 struct {
	c ChaCha20Configuration
}

func NewCryptChaCha20(c ChaCha20Configuration) *ChaCha20 {
	return &ChaCha20{c: c}
}

// Encrypt returns a ChaCha encryption of plaintext
func (c *ChaCha20) Encrypt(ctx context.Context, clearString string) (string, error) {
	if len(clearString) == 0 {
		return "", nil
	}
	plaintext := []byte(clearString)

	if len(c.c.Config(ctx).SecretsCipher()) == 0 {
		return "", herodot.ErrInternalServerError.WithReason("Unable to encrypt message because no cipher secrets were configured.")
	}

	aead, err := chacha20poly1305.NewX(c.c.Config(ctx).SecretsCipher()[0][:])
	if err != nil {
		return "", herodot.ErrInternalServerError.WithWrap(err).WithReason("Unable to generate key")
	}

	nonce := make([]byte, aead.NonceSize(), aead.NonceSize()+len(plaintext)+aead.Overhead())
	_, err = io.ReadFull(rand.Reader, nonce)
	if err != nil {
		return "", herodot.ErrInternalServerError.WithWrap(err).WithReason("Unable to generate nonce")
	}

	encryptedMsg := aead.Seal(nonce, nonce, plaintext, nil)
	return hex.EncodeToString(encryptedMsg), nil
}

// Decrypt decrypt data using 256 bit key
func (c *ChaCha20) Decrypt(ctx context.Context, encryptedString string) (string, error) {
	if len(encryptedString) == 0 {
		return "", nil
	}

	secrets := c.c.Config(ctx).SecretsCipher()
	if len(secrets) == 0 {
		return "", errors.WithStack(herodot.ErrInternalServerError.WithReason("Unable to decipher the encrypted message because no cipher secrets were configured."))
	}

	ciphertext, err := hex.DecodeString(encryptedString)
	if err != nil {
		return "", errors.WithStack(herodot.ErrInternalServerError.WithWrap(err).WithReason("Unable to decode hex encrypted string"))
	}

	for i := range secrets {
		aead, err := chacha20poly1305.NewX(secrets[i][:])
		if err != nil {
			return "", herodot.ErrInternalServerError.WithWrap(err).WithReason("Unable to instanciate chacha20")
		}
		if len(ciphertext) < aead.NonceSize() {
			return "", herodot.ErrInternalServerError.WithReason("cipher text too short")
		}
		nonce, ciphertext := ciphertext[:aead.NonceSize()], ciphertext[aead.NonceSize():]
		plaintext, err := aead.Open(nil, nonce, ciphertext, nil)
		if err == nil {
			return string(plaintext), nil
		}
	}

	return "", herodot.ErrInternalServerError.WithReason("Unable to decrypt string")
}
