package hash

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/crypto/sha3"

	"github.com/ory/kratos/driver/config"
)

var ErrUnknownHashAlgorithm = fmt.Errorf("unknown hash algorithm")
var ErrUnknownHashFormat = fmt.Errorf("unknown hash format")

func Compare(ctx context.Context, cfg *config.Config, password []byte, hash []byte) error {
	algorithm, realHash, err := ParsePasswordHash(hash)
	if err != nil {
		return errors.WithStack(err)
	}

	switch algorithm {
	case Argon2AlgorithmId:
		return CompareArgon2id(ctx, cfg, password, realHash)
	case BcryptAlgorithmId:
		return CompareBcrypt(ctx, cfg, password, realHash)
	case BcryptAESAlgorithmId:
		return CompareBcryptAes(ctx, cfg, password, realHash)
	default:
		return ErrUnknownHashAlgorithm
	}
}

func CompareBcrypt(_ context.Context, _ *config.Config, password []byte, hash string) error {
	if err := validateBcryptPasswordLength(password); err != nil {
		return errors.WithStack(err)
	}

	err := bcrypt.CompareHashAndPassword([]byte(hash), password)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func IsPasswordHash(hash []byte) bool {
	_, _, err := ParsePasswordHash(hash)
	return err == nil
}

func ParsePasswordHash(input []byte) (algorithm, hash string, err error) {
	hashParts := strings.SplitN(string(input), "$", 3)
	if len(hashParts) != 3 {
		err = errors.WithStack(ErrUnknownHashFormat)
		return
	}

	hash = "$" + hashParts[2]
	switch hashParts[1] {
	case Argon2AlgorithmId:
		algorithm = Argon2AlgorithmId
		return
	case BcryptAlgorithmId:
		algorithm = BcryptAlgorithmId
		return
	case BcryptAESAlgorithmId:
		algorithm = BcryptAESAlgorithmId
		return
	default:
		err = ErrUnknownHashAlgorithm
		return
	}
}

func aes256Decrypt(data string, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	decoded, err := hex.DecodeString(data)
	if err != nil {
		return nil, err
	}

	nonce := decoded[:gcm.NonceSize()]
	ciphertext := decoded[gcm.NonceSize():]

	return gcm.Open(nil, nonce, ciphertext, nil)
}

func CompareBcryptAes(_ context.Context, cfg *config.Config, password []byte, hash string) error {
	aesDecrypted, err := aes256Decrypt(hash[1:], []byte(cfg.HasherBcryptAES().Key))
	if err != nil {
		return errors.WithStack(err)
	}

	sh := sha3.New512()
	sh.Write(password)
	err = bcrypt.CompareHashAndPassword(aesDecrypted, sh.Sum(nil))
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func CompareArgon2id(_ context.Context, _ *config.Config, password []byte, hashPart string) error {
	// Extract the parameters, salt and derived key from the encoded password
	// hash.
	p, salt, hash, err := decodeArgon2idHash(hashPart)
	if err != nil {
		return errors.WithStack(err)
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
	if len(parts) != 5 {
		return nil, nil, nil, ErrInvalidHash
	}

	var version int
	_, err = fmt.Sscanf(parts[1], "v=%d", &version)
	if err != nil {
		return nil, nil, nil, err
	}
	if version != argon2.Version {
		return nil, nil, nil, ErrIncompatibleVersion
	}

	p = new(config.Argon2)
	_, err = fmt.Sscanf(parts[2], "m=%d,t=%d,p=%d", &p.Memory, &p.Iterations, &p.Parallelism)
	if err != nil {
		return nil, nil, nil, err
	}

	salt, err = base64.RawStdEncoding.Strict().DecodeString(parts[3])
	if err != nil {
		return nil, nil, nil, err
	}
	p.SaltLength = uint32(len(salt))

	hash, err = base64.RawStdEncoding.Strict().DecodeString(parts[4])
	if err != nil {
		return nil, nil, nil, err
	}
	p.KeyLength = uint32(len(hash))

	return p, salt, hash, nil
}
