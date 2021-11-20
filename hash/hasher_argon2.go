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
)

var (
	ErrInvalidHash               = errors.New("the encoded hash is not in the correct format")
	ErrIncompatibleVersion       = errors.New("incompatible version of argon2")
	ErrMismatchedHashAndPassword = errors.New("passwords do not match")
)

type Argon2 struct {
	c Argon2Configuration
}

type Argon2Configuration interface {
	ConfigProvider
}

func NewHasherArgon2(c Argon2Configuration) *Argon2 {
	return &Argon2{c: c}
}

func toKB(mem bytesize.ByteSize) uint32 {
	return uint32(mem / bytesize.KB)
}

func (h *Argon2) Generate(ctx context.Context, password []byte) ([]byte, error) {
	p := h.c.HashConfig(ctx).HasherArgon2()

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
		"$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version, toKB(p.Memory), p.Iterations, p.Parallelism,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(hash),
	); err != nil {
		return nil, errors.WithStack(err)
	}

	return b.Bytes(), nil
}

func (h *Argon2) Understands(hash []byte) bool {
	return IsArgon2idHash(hash)
}
