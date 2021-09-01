package cipher

import (
	"context"
	"encoding/hex"
	"github.com/ory/kratos/driver/config"
)

// Noop is default cipher implementation witch does not do encryption

type NoopConfiguration interface {
	config.Provider
}

type Noop struct {
	c NoopConfiguration
}

func NewNoop(c NoopConfiguration) *Noop {
	return &Noop{c: c}
}

// Encrypt returns a ChaCha encryption of plaintext
func (c *Noop) Encrypt(ctx context.Context, message []byte) (string, error) {
	return hex.EncodeToString(message), nil
}

// Decrypt decrypts data using 256 bit key
func (c *Noop) Decrypt(ctx context.Context, ciphertext string) ([]byte, error) {
	return hex.DecodeString(ciphertext)
}
