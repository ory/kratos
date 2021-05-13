package hash

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"

	"github.com/inhies/go-bytesize"

	"github.com/pkg/errors"
	"golang.org/x/crypto/argon2"

	"github.com/ory/kratos/driver/config"
)

var (
	ErrInvalidHash               = errors.New("the encoded hash is not in the correct format")
	ErrIncompatibleVersion       = errors.New("incompatible version of argon2")
	ErrMismatchedHashAndPassword = errors.New("passwords do not match")
)

var Argon2AlgorithmId = []byte("argon2id")

type Argon2 struct {
	c Argon2Configuration
}

type Argon2Configuration interface {
	config.Provider
}

func NewHasherArgon2(c Argon2Configuration) *Argon2 {
	return &Argon2{c: c}
}

func toKB(mem bytesize.ByteSize) uint32 {
	return uint32(mem / bytesize.KB)
}

func (h *Argon2) Generate(ctx context.Context, password []byte) ([]byte, error) {
	p := h.c.Config(ctx).HasherArgon2()

	salt := make([]byte, p.SaltLength)
	if _, err := rand.Read(salt); err != nil {
		return nil, err
	}

	// Pass the plaintext password, salt and parameters to the argon2.IDKey
	// function. This will generate a hash of the password using the Argon2id
	// variant.
	hash := argon2.IDKey([]byte(password), salt, p.Iterations, toKB(p.Memory), p.Parallelism, p.KeyLength)

	var b bytes.Buffer
	if _, err := fmt.Fprintf(
		&b,
		"$%s$v=%d$m=%d,t=%d,p=%d$%s$%s",
		Argon2AlgorithmId,
		argon2.Version, toKB(p.Memory), p.Iterations, p.Parallelism,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(hash),
	); err != nil {
		return nil, errors.WithStack(err)
	}

	return b.Bytes(), nil
}
