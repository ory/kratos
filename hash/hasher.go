package hash

import "context"

// Hasher provides methods for generating and comparing password hashes.
type Hasher interface {
	// Generate returns a hash derived from the password or an error if the hash method failed.
	Generate(ctx context.Context, password []byte) ([]byte, error)

	// IsSameAlgorithm returns whether given password hashed algorithm is same or not.
	IsSameAlgorithm(hash []byte) bool
}

type HashProvider interface {
	Hasher() Hasher
}
