package hash

import (
	"bytes"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"

	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/crypto/sha3"

	"github.com/ory/kratos/driver/config"
)

const BcryptAESAlgorithmId = "bcryptAes"

type BcryptAES struct {
	c BcryptAESConfiguration
}

type BcryptAESConfiguration interface {
	config.Provider
}

func NewHasherBcryptAES(c BcryptAESConfiguration) *BcryptAES {
	return &BcryptAES{c: c}
}

func (h *BcryptAES) aes256Encrypt(data, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, errors.WithStack(err)
	}

	hash := append(nonce, gcm.Seal(nil, nonce, data, nil)...)
	encoded := make([]byte, hex.EncodedLen(len(hash)))
	hex.Encode(encoded, hash)

	return encoded, nil
}

func (h *BcryptAES) Generate(ctx context.Context, password []byte) ([]byte, error) {
	cfg := h.c.Config(ctx).HasherBcryptAES()
	sh := sha3.New512()
	sh.Write(password)
	bcryptPassword, err := bcrypt.GenerateFromPassword(sh.Sum(nil), int(cfg.Cost))
	if err != nil {
		return nil, errors.WithStack(err)
	}

	hash, err := h.aes256Encrypt(bcryptPassword, []byte(cfg.Key))

	if err != nil {
		return nil, errors.WithStack(err)
	}

	var b bytes.Buffer
	if _, err := fmt.Fprintf(
		&b,
		"$%s$%s",
		BcryptAESAlgorithmId,
		hash,
	); err != nil {
		return nil, errors.WithStack(err)
	}

	return b.Bytes(), nil
}
