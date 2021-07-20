package crypt

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
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithReason("Expected plaintext string does not empty"))
	}
	ciphertext, err := cryptopasta.Encrypt([]byte(clearString), &a.c.Config(ctx).SecretsAES()[0])
	return ciphertext, err

}

// Decrypt returns the decrypted aes data
func (a *AES) Decrypt(ctx context.Context, encryptedString []byte) (string, error) {
	if len(encryptedString) == 0 {
		return "", errors.WithStack(herodot.ErrInternalServerError.WithReason("Expected encrypted string does not empty"))
	}
	var err error
	var plaintext []byte
	secrets := a.c.Config(ctx).SecretsAES()
	for i := range secrets {
		plaintext, err = cryptopasta.Decrypt(encryptedString, &secrets[i])
		if err == nil {
			break
		}
	}
	return string(plaintext), err

	//return "", errors.WithStack(herodot.ErrNotFound.WithReason("secret key not found for decryption"))

}
