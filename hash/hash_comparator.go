package hash

import (
	"context"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"regexp"
	"strings"

	"github.com/pkg/errors"
	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/bcrypt"

	"github.com/ory/kratos/driver/config"
)

var ErrUnknownHashAlgorithm = errors.New("unknown hash algorithm")

func Compare(ctx context.Context, password []byte, hash []byte) error {
	if IsBcryptHash(hash) {
		return CompareBcrypt(ctx, password, hash)
	} else if IsArgon2idHash(hash) {
		return CompareArgon2id(ctx, password, hash)
	} else {
		return ErrUnknownHashAlgorithm
	}
}

func CompareBcrypt(_ context.Context, password []byte, hash []byte) error {
	if err := validateBcryptPasswordLength(password); err != nil {
		return err
	}

	err := bcrypt.CompareHashAndPassword(hash, password)
	if err != nil {
		return err
	}

	return nil
}

func CompareArgon2id(_ context.Context, password []byte, hash []byte) error {
	// Extract the parameters, salt and derived key from the encoded password
	// hash.
	p, salt, hash, err := decodeArgon2idHash(string(hash))
	if err != nil {
		return err
	}

	// Derive the key from the other password using the same parameters.
	otherHash := argon2.IDKey([]byte(password), salt, p.Iterations, uint32(p.Memory), p.Parallelism, p.KeyLength)

	// Check that the contents of the hashed passwords are identical. Note
	// that we are using the subtle.ConstantTimeCompare() function for this
	// to help prevent timing attacks.
	if subtle.ConstantTimeCompare(hash, otherHash) == 1 {
		return nil
	}
	return ErrMismatchedHashAndPassword
}

func IsBcryptHash(hash []byte) bool {
	res, _ := regexp.Match("^\\$2[abzy]?\\$", hash)
	return res
}

func IsArgon2idHash(hash []byte) bool {
	res, _ := regexp.Match("^\\$argon2id\\$", hash)
	return res
}

func decodeArgon2idHash(encodedHash string) (p *config.Argon2, salt, hash []byte, err error) {
	parts := strings.Split(encodedHash, "$")
	if len(parts) != 6 {
		return nil, nil, nil, ErrInvalidHash
	}

	var version int
	_, err = fmt.Sscanf(parts[2], "v=%d", &version)
	if err != nil {
		return nil, nil, nil, err
	}
	if version != argon2.Version {
		return nil, nil, nil, ErrIncompatibleVersion
	}

	p = new(config.Argon2)
	_, err = fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &p.Memory, &p.Iterations, &p.Parallelism)
	if err != nil {
		return nil, nil, nil, err
	}

	salt, err = base64.RawStdEncoding.Strict().DecodeString(parts[4])
	if err != nil {
		return nil, nil, nil, err
	}
	p.SaltLength = uint32(len(salt))

	hash, err = base64.RawStdEncoding.Strict().DecodeString(parts[5])
	if err != nil {
		return nil, nil, nil, err
	}
	p.KeyLength = uint32(len(hash))

	return p, salt, hash, nil
}
