package hash

import "context"

// Hasher provides methods for generating and comparing password hashes.
type Hasher interface {
	// Generate returns a hash derived from the password or an error if the hash method failed.
	Generate(ctx context.Context, password []byte) ([]byte, error)
}

type HashProvider interface {
	Hasher() Hasher
}
