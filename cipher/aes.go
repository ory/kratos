package cipher

import (
	"context"
	"encoding/hex"

	"github.com/gtank/cryptopasta"

	"github.com/ory/herodot"

	"github.com/pkg/errors"

	"github.com/ory/kratos/driver/config"
)

type AES struct {
	c AESConfiguration
}

type AESConfiguration interface {
	config.Provider
}

func NewCryptAES(c AESConfiguration) *AES {
	return &AES{c: c}
}

// Encrypt return a AES encrypt of plaintext
func (a *AES) Encrypt(ctx context.Context, message []byte) (string, error) {
	if len(message) == 0 {
		// do nothing if empty instead of return an error
		// return nil, errors.WithStack(herodot.ErrInternalServerError.WithReason("Can not encrypt empty string."))
		return "", nil
	}

	if len(a.c.Config(ctx).SecretsCipher()) == 0 {
		return "", errors.WithStack(herodot.ErrInternalServerError.WithReason("Unable to encrypt message because no cipher secrets were configured."))
	}

	ciphertext, err := cryptopasta.Encrypt(message, &a.c.Config(ctx).SecretsCipher()[0])
	return hex.EncodeToString(ciphertext), errors.WithStack(err)
}

// Decrypt returns the decrypted aes data
func (a *AES) Decrypt(ctx context.Context, ciphertext string) ([]byte, error) {
	if len(ciphertext) == 0 {
		// do nothing if empty instead of return an error
		// return "", errors.WithStack(herodot.ErrInternalServerError.WithReason("Can not decrypt empty message."))
		return nil, nil
	}

	secrets := a.c.Config(ctx).SecretsCipher()
	if len(secrets) == 0 {
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithReason("Unable to decipher the encrypted message because no AES secrets were configured."))
	}

	decode, err := hex.DecodeString(ciphertext)
	if err != nil {
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithWrap(err))
	}

	for i := range secrets {
		plaintext, err := cryptopasta.Decrypt(decode, &secrets[i])
		if err == nil {
			return plaintext, nil
		}
	}

	return nil, errors.WithStack(herodot.ErrInternalServerError.WithReason("Unable to decipher the encrypted message."))
}
