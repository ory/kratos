package hash

import (
	"crypto/sha1" // #nosec G505 - compatibility for imported passwords
	"crypto/sha256"
	"crypto/sha512"
	"hash"

	"golang.org/x/crypto/sha3"
)

type Pbkdf2 struct {
	Algorithm  string
	Iterations uint32
	SaltLength uint32
	KeyLength  uint32
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
