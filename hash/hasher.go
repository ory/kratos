package hash

import (
	"context"
	"time"

	"github.com/inhies/go-bytesize"
)

// Hasher provides methods for generating and comparing password hashes.
type Hasher interface {
	// Generate returns a hash derived from the password or an error if the hash method failed.
	Generate(ctx context.Context, password []byte) ([]byte, error)

	// Understands returns whether the given hash can be understood by this hasher.
	Understands(hash []byte) bool
}

// Generator is the interface that objects that can construct hashers must implement
type Generator interface {
	Hasher() Hasher
}

// HashConfigProvider is the interface that objects that can generate configuration for the implemented hashers must
// implement
type HashConfigProvider interface {
	HasherBcrypt() *BcryptConfig
	HasherArgon2() *Argon2Config
}

type BcryptConfig struct {
	Cost uint32 `json:"cost"`
}

type Argon2Config struct {
	Memory            bytesize.ByteSize `json:"memory"`
	Iterations        uint32            `json:"iterations"`
	Parallelism       uint8             `json:"parallelism"`
	SaltLength        uint32            `json:"salt_length"`
	KeyLength         uint32            `json:"key_length"`
	ExpectedDuration  time.Duration     `json:"expected_duration"`
	ExpectedDeviation time.Duration     `json:"expected_deviation"`
	DedicatedMemory   bytesize.ByteSize `json:"dedicated_memory"`
}

type ConfigProvider interface {
	HashConfig(ctx context.Context) HashConfigProvider
}

func NewHasher(algorithm string, provider ConfigProvider) Hasher {
	if algorithm == "bcrypt" {
		return NewHasherBcrypt(provider)
	} else {
		return NewHasherArgon2(provider)
	}
}
