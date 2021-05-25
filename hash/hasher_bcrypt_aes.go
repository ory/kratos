package hash

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"

	"github.com/gtank/cryptopasta"
	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/crypto/sha3"

	"github.com/ory/kratos/driver/config"
)

var BcryptAESAlgorithmId = []byte("bcryptaes")

type BcryptAES struct {
	c BcryptAESConfiguration
}

type BcryptAESConfiguration interface {
	config.Provider
}

func NewHasherBcryptAES(c BcryptAESConfiguration) *BcryptAES {
	return &BcryptAES{c: c}
}

func (h *BcryptAES) Generate(ctx context.Context, password []byte) ([]byte, error) {
	cfg, err := h.c.Config(ctx).HasherBcryptAES()
	if err != nil {
		return nil, errors.WithStack(err)
	}

	sh := sha3.New512()
	sh.Write(password)
	bcryptPassword, err := bcrypt.GenerateFromPassword(sh.Sum(nil), int(cfg.Cost))
	if err != nil {
		return nil, errors.WithStack(err)
	}

	hash, err := cryptopasta.Encrypt(bcryptPassword, &cfg.Key[0])
	if err != nil {
		return nil, errors.WithStack(err)
	}

	encoded := make([]byte, hex.EncodedLen(len(hash)))
	hex.Encode(encoded, hash[:])

	var b bytes.Buffer
	if _, err := fmt.Fprintf(
		&b,
		"$%s$%s",
		BcryptAESAlgorithmId,
		encoded,
	); err != nil {
		return nil, errors.WithStack(err)
	}

	return b.Bytes(), nil
}
