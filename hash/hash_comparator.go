package hash

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/bcrypt"

	"github.com/ory/kratos/driver/config"
)

var ErrUnknownHashAlgorithm = errors.New("unknown hash algorithm")
var ErrUnknownHashFormat = errors.New("unknown hash format")

func Compare(ctx context.Context, cfg *config.Config, password []byte, hash []byte) error {
	algorithm, realHash, err := parsePasswordHash(hash)
	if err != nil {
		return errors.WithStack(err)
	}

	switch algorithm {
	case Argon2AlgorithmId:
		return CompareArgon2id(ctx, password, realHash)
	case BcryptAlgorithmId:
		return CompareBcrypt(ctx, password, realHash)
	case BcryptAESAlgorithmId:
		return CompareBcryptAes(ctx, cfg.HasherBcryptAES(), password, realHash)
	default:
		return ErrUnknownHashAlgorithm
	}
}

func CompareBcrypt(_ context.Context, password []byte, hash string) error {
	if err := validateBcryptPasswordLength(password); err != nil {
		return err
	}

	err := bcrypt.CompareHashAndPassword([]byte(hash), password)
	if err != nil {
		return err
	}

	return nil
}

func IsPasswordHash(hash []byte) bool {
	_, _, err := parsePasswordHash(hash)
	return err == nil
}

func parsePasswordHash(input []byte) (algorithm, hash string, err error) {
	hashParts := strings.SplitN(string(input), "$", 3)
	if len(hashParts) != 3 {
		err = ErrUnknownHashFormat
		return
	}

	switch hashParts[1] {
	case Argon2AlgorithmId:
		algorithm = Argon2AlgorithmId
		hash = hashParts[2]
		return
	case BcryptAlgorithmId:
		algorithm = BcryptAlgorithmId
		hash = hashParts[2]
		return
	case BcryptAESAlgorithmId:
		algorithm = BcryptAESAlgorithmId
		hash = hashParts[2]
		return
	default:
		err = ErrUnknownHashAlgorithm
		return
	}
}

func aes256Decrypt(data string, key []byte) ([]byte, error) {
	dataHex, err := base64.RawStdEncoding.Strict().DecodeString(data)
	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := dataHex[:gcm.NonceSize()]
	ciphertext := dataHex[gcm.NonceSize():]

	return gcm.Open(nil, nonce, ciphertext, nil)
}

func CompareBcryptAes(_ context.Context, hashConfig *config.BcryptAES, password []byte, hash string) error {
	decrypted, err := aes256Decrypt(hash, []byte(hashConfig.Key))
	if err != nil {
		return errors.WithStack(err)
	}

	err = bcrypt.CompareHashAndPassword(decrypted, password)
	if err != nil {
		return err
	}

	return nil
}

func CompareArgon2id(_ context.Context, password []byte, hashPart string) error {
	// Extract the parameters, salt and derived key from the encoded password
	// hash.
	p, salt, hash, err := decodeArgon2idHash(hashPart)
	if err != nil {
		return err
	}

	// Derive the key from the other password using the same parameters.
	otherHash := argon2.IDKey(password, salt, p.Iterations, uint32(p.Memory), p.Parallelism, p.KeyLength)

	// Check that the contents of the hashed passwords are identical. Note
	// that we are using the subtle.ConstantTimeCompare() function for this
	// to help prevent timing attacks.
	if subtle.ConstantTimeCompare(hash, otherHash) == 1 {
		return nil
	}
	return ErrMismatchedHashAndPassword
}

func decodeArgon2idHash(encodedHash string) (p *config.Argon2, salt, hash []byte, err error) {
	parts := strings.Split(encodedHash, "$")
	if len(parts) != 4 {
		return nil, nil, nil, ErrInvalidHash
	}

	var version int
	_, err = fmt.Sscanf(parts[0], "v=%d", &version)
	if err != nil {
		return nil, nil, nil, err
	}
	if version != argon2.Version {
		return nil, nil, nil, ErrIncompatibleVersion
	}

	p = new(config.Argon2)
	_, err = fmt.Sscanf(parts[1], "m=%d,t=%d,p=%d", &p.Memory, &p.Iterations, &p.Parallelism)
	if err != nil {
		return nil, nil, nil, err
	}

	salt, err = base64.RawStdEncoding.Strict().DecodeString(parts[2])
	if err != nil {
		return nil, nil, nil, err
	}
	p.SaltLength = uint32(len(salt))

	hash, err = base64.RawStdEncoding.Strict().DecodeString(parts[3])
	if err != nil {
		return nil, nil, nil, err
	}
	p.KeyLength = uint32(len(hash))

	return p, salt, hash, nil
}
