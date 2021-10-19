package hash

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha1" // #nosec G505 - compatibility for imported passwords
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"fmt"
	"hash"

	"github.com/pkg/errors"
	"golang.org/x/crypto/pbkdf2"
	"golang.org/x/crypto/sha3"
)

type Pbkdf2 struct {
	Algorithm  string
	Iterations uint32
	SaltLength uint32
	KeyLength  uint32
}

func (h *Pbkdf2) Generate(_ context.Context, password []byte) ([]byte, error) {
	salt := make([]byte, h.SaltLength)
	if _, err := rand.Read(salt); err != nil {
		return nil, err
	}

	key := pbkdf2.Key(password, salt, int(h.Iterations), int(h.KeyLength), getPseudorandomFunctionForPbkdf2(h.Algorithm))

	var b bytes.Buffer
	if _, err := fmt.Fprintf(
		&b,
		"$pbkdf2-%s$i=%d,l=%d$%s$%s",
		h.Algorithm,
		h.Iterations,
		h.KeyLength,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(key),
	); err != nil {
		return nil, errors.WithStack(err)
	}

	return b.Bytes(), nil
}

func (h *Pbkdf2) Understands(hash []byte) bool {
	return IsPbkdf2Hash(hash)
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
