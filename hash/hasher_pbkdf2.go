package hash

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"fmt"
	"hash"

	"github.com/pkg/errors"
	"golang.org/x/crypto/pbkdf2"
	"golang.org/x/crypto/sha3"

	"github.com/ory/kratos/driver/config"
)

type Pbkdf2 struct {
	c Pbkdf2Configuration
}

type Pbkdf2Configuration interface {
	config.Provider
}

func NewHasherPbkdf2(c Pbkdf2Configuration) *Pbkdf2 {
	return &Pbkdf2{c: c}
}

func (h *Pbkdf2) Generate(ctx context.Context, password []byte) ([]byte, error) {
	p := h.c.Config(ctx).HasherPbkdf2()

	salt := make([]byte, p.SaltLength)
	if _, err := rand.Read(salt); err != nil {
		return nil, err
	}

	key := pbkdf2.Key(password, salt, int(p.Iterations), int(p.KeyLength), getPseudorandomFunctionForPbkdf2(p.Algorithm))

	var b bytes.Buffer
	if _, err := fmt.Fprintf(
		&b,
		"$pbkdf2_%s$c=%d$%s$%s",
		p.Algorithm,
		p.Iterations,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(key),
	); err != nil {
		return nil, errors.WithStack(err)
	}

	return b.Bytes(), nil
}

func getPseudorandomFunctionForPbkdf2(alg string) func() hash.Hash {
	switch alg {
	case "sha1":
		return sha1.New
	case "sha224":
		return sha3.New224
	case "sha256":
		return sha256.New
	case "sha384":
		return sha3.New384
	case "sha512":
		return sha512.New
	default:
		return sha256.New
	}
}
