package cipher

import (
	"context"

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
func (a *AES) Encrypt(ctx context.Context, clearString string) ([]byte, error) {
	if len(clearString) == 0 {
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithReason("Can not encrypt empty string."))
	}

	if len(a.c.Config(ctx).SecretsAES()) == 0 {
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithReason("Unable to encrypt message because no AES secrets were configured."))
	}

	ciphertext, err := cryptopasta.Encrypt([]byte(clearString), &a.c.Config(ctx).SecretsAES()[0])
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return ciphertext, nil
}

// Decrypt returns the decrypted aes data
func (a *AES) Decrypt(ctx context.Context, encryptedString []byte) (string, error) {
	if len(encryptedString) == 0 {
		return "", errors.WithStack(herodot.ErrInternalServerError.WithReason("Can not decrypt empty message."))
	}

	secrets := a.c.Config(ctx).SecretsAES()
	if len(secrets) == 0 {
		return "", errors.WithStack(herodot.ErrInternalServerError.WithReason("Unable to decipher the encrypted message because no AES secrets were configured."))
	}

	for i := range secrets {
		plaintext, err := cryptopasta.Decrypt(encryptedString, &secrets[i])
		if err != nil {
			return string(plaintext), nil
		}
	}

	return "", errors.WithStack(herodot.ErrForbidden.WithReason("Unable to decipher the encrypted message."))
}
