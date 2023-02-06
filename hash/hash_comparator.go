// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package hash

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"  // #nosec G501
	"crypto/sha1" // #nosec G505 - compatibility for imported passwords
	"crypto/sha256"
	"crypto/sha512"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"regexp"
	"strings"

	"github.com/pkg/errors"
	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/crypto/pbkdf2"
	"golang.org/x/crypto/scrypt"

	"github.com/ory/kratos/driver/config"
)

var ErrUnknownHashAlgorithm = errors.New("unknown hash algorithm")

func Compare(ctx context.Context, password []byte, hash []byte) error {
	switch {
	case IsBcryptHash(hash):
		return CompareBcrypt(ctx, password, hash)
	case IsArgon2idHash(hash):
		return CompareArgon2id(ctx, password, hash)
	case IsArgon2iHash(hash):
		return CompareArgon2i(ctx, password, hash)
	case IsPbkdf2Hash(hash):
		return ComparePbkdf2(ctx, password, hash)
	case IsScryptHash(hash):
		return CompareScrypt(ctx, password, hash)
	case IsSSHAHash(hash):
		return CompareSSHA(ctx, password, hash)
	case IsSHAHash(hash):
		return CompareSHA(ctx, password, hash)
	case IsFirebaseScryptHash(hash):
		return CompareFirebaseScrypt(ctx, password, hash)
	case IsMD5Hash(hash):
		return CompareMD5(ctx, password, hash)
	default:
		return errors.WithStack(ErrUnknownHashAlgorithm)
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
	return errors.WithStack(ErrMismatchedHashAndPassword)
}

func CompareArgon2i(_ context.Context, password []byte, hash []byte) error {
	// Extract the parameters, salt and derived key from the encoded password
	// hash.
	p, salt, hash, err := decodeArgon2idHash(string(hash))
	if err != nil {
		return err
	}

	// Derive the key from the other password using the same parameters.
	otherHash := argon2.Key([]byte(password), salt, p.Iterations, uint32(p.Memory), p.Parallelism, p.KeyLength)

	// Check that the contents of the hashed passwords are identical. Note
	// that we are using the subtle.ConstantTimeCompare() function for this
	// to help prevent timing attacks.
	if subtle.ConstantTimeCompare(hash, otherHash) == 1 {
		return nil
	}
	return errors.WithStack(ErrMismatchedHashAndPassword)
}

func ComparePbkdf2(_ context.Context, password []byte, hash []byte) error {
	// Extract the parameters, salt and derived key from the encoded password
	// hash.
	p, salt, hash, err := decodePbkdf2Hash(string(hash))
	if err != nil {
		return err
	}

	// Derive the key from the other password using the same parameters.
	otherHash := pbkdf2.Key(password, salt, int(p.Iterations), int(p.KeyLength), getPseudorandomFunctionForPbkdf2(p.Algorithm))

	// Check that the contents of the hashed passwords are identical. Note
	// that we are using the subtle.ConstantTimeCompare() function for this
	// to help prevent timing attacks.
	if subtle.ConstantTimeCompare(hash, otherHash) == 1 {
		return nil
	}
	return errors.WithStack(ErrMismatchedHashAndPassword)
}

func CompareScrypt(_ context.Context, password []byte, hash []byte) error {
	// Extract the parameters, salt and derived key from the encoded password
	// hash.
	p, salt, hash, err := decodeScryptHash(string(hash))
	if err != nil {
		return err
	}

	// Derive the key from the other password using the same parameters.
	otherHash, err := scrypt.Key(password, salt, int(p.Cost), int(p.Block), int(p.Parrellization), int(p.KeyLength))
	if err != nil {
		return errors.WithStack(err)
	}

	// Check that the contents of the hashed passwords are identical. Note
	// that we are using the subtle.ConstantTimeCompare() function for this
	// to help prevent timing attacks.
	if subtle.ConstantTimeCompare(hash, otherHash) == 1 {
		return nil
	}
	return errors.WithStack(ErrMismatchedHashAndPassword)
}

func CompareSSHA(_ context.Context, password []byte, hash []byte) error {
	hasher, salt, hash, err := decodeSSHAHash(string(hash))

	if err != nil {
		return err
	}

	raw := append(password[:], salt[:]...)

	return compareSHAHelper(hasher, raw, hash)
}

func CompareSHA(_ context.Context, password []byte, hash []byte) error {

	hasher, pf, salt, hash, err := decodeSHAHash(string(hash))
	if err != nil {
		return err
	}

	r := strings.NewReplacer("{SALT}", string(salt), "{PASSWORD}", string(password))
	raw := []byte(r.Replace(string(pf)))

	return compareSHAHelper(hasher, raw, hash)
}

func CompareFirebaseScrypt(_ context.Context, password []byte, hash []byte) error {
	// Extract the parameters, salt and derived key from the encoded password
	// hash.
	p, salt, saltSeparator, hash, signerKey, err := decodeFirebaseScryptHash(string(hash))
	if err != nil {
		return err
	}

	// Derive the key from the other password using the same parameters.
	// FirebaseScript algorithm implementation from https://github.com/Aoang/firebase-scrypt
	ck, err := scrypt.Key(password, append(salt, saltSeparator...), int(p.Cost), int(p.Block), int(p.Parrellization), 32)
	if err != nil {
		return errors.WithStack(err)
	}

	var block cipher.Block
	if block, err = aes.NewCipher(ck); err != nil {
		return errors.WithStack(err)
	}

	cipherText := make([]byte, aes.BlockSize+len(signerKey))
	stream := cipher.NewCTR(block, cipherText[:aes.BlockSize])
	stream.XORKeyStream(cipherText[aes.BlockSize:], signerKey)
	otherHash := cipherText[aes.BlockSize:]

	// Check that the contents of the hashed passwords are identical. Note
	// that we are using the subtle.ConstantTimeCompare() function for this
	// to help prevent timing attacks.
	if subtle.ConstantTimeCompare(hash, otherHash) == 1 {
		return nil
	}
	return errors.WithStack(ErrMismatchedHashAndPassword)
}

func CompareMD5(_ context.Context, password []byte, hash []byte) error {
	// Extract the hash from the encoded password
	pf, salt, hash, err := decodeMD5Hash(string(hash))
	if err != nil {
		return err
	}

	arg := password
	if salt != nil {
		r := strings.NewReplacer("{SALT}", string(salt), "{PASSWORD}", string(password))
		arg = []byte(r.Replace(string(pf)))
	}
	// #nosec G401
	otherHash := md5.Sum(arg)

	// Check that the contents of the hashed passwords are identical. Note
	// that we are using the subtle.ConstantTimeCompare() function for this
	// to help prevent timing attacks.
	if subtle.ConstantTimeCompare(hash, otherHash[:]) == 1 {
		return nil
	}
	return errors.WithStack(ErrMismatchedHashAndPassword)
}

var (
	isBcryptHash         = regexp.MustCompile(`^\$2[abzy]?\$`)
	isArgon2idHash       = regexp.MustCompile(`^\$argon2id\$`)
	isArgon2iHash        = regexp.MustCompile(`^\$argon2i\$`)
	isPbkdf2Hash         = regexp.MustCompile(`^\$pbkdf2-sha[0-9]{1,3}\$`)
	isScryptHash         = regexp.MustCompile(`^\$scrypt\$`)
	isSSHAHash           = regexp.MustCompile(`^{SSHA(256|512)?}.*`)
	isSHAHash            = regexp.MustCompile(`^\$sha(1|256|512)\$`)
	isFirebaseScryptHash = regexp.MustCompile(`^\$firescrypt\$`)
	isMD5Hash            = regexp.MustCompile(`^\$md5\$`)
)

func IsBcryptHash(hash []byte) bool         { return isBcryptHash.Match(hash) }
func IsArgon2idHash(hash []byte) bool       { return isArgon2idHash.Match(hash) }
func IsArgon2iHash(hash []byte) bool        { return isArgon2iHash.Match(hash) }
func IsPbkdf2Hash(hash []byte) bool         { return isPbkdf2Hash.Match(hash) }
func IsScryptHash(hash []byte) bool         { return isScryptHash.Match(hash) }
func IsSSHAHash(hash []byte) bool           { return isSSHAHash.Match(hash) }
func IsSHAHash(hash []byte) bool            { return isSHAHash.Match(hash) }
func IsFirebaseScryptHash(hash []byte) bool { return isFirebaseScryptHash.Match(hash) }
func IsMD5Hash(hash []byte) bool            { return isMD5Hash.Match(hash) }

func IsValidHashFormat(hash []byte) bool {
	if IsBcryptHash(hash) ||
		IsArgon2idHash(hash) ||
		IsArgon2iHash(hash) ||
		IsPbkdf2Hash(hash) ||
		IsScryptHash(hash) ||
		IsSSHAHash(hash) ||
		IsSHAHash(hash) ||
		IsFirebaseScryptHash(hash) ||
		IsMD5Hash(hash) {
		return true
	} else {
		return false
	}
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

// decodePbkdf2Hash decodes PBKDF2 encoded password hash.
// format: $pbkdf2-<digest>$i=<iterations>,l=<length>$<salt>$<hash>
func decodePbkdf2Hash(encodedHash string) (p *Pbkdf2, salt, hash []byte, err error) {
	parts := strings.Split(encodedHash, "$")
	if len(parts) != 5 {
		return nil, nil, nil, ErrInvalidHash
	}

	p = new(Pbkdf2)
	digestParts := strings.SplitN(parts[1], "-", 2)
	if len(digestParts) != 2 {
		return nil, nil, nil, ErrInvalidHash
	}
	p.Algorithm = digestParts[1]

	_, err = fmt.Sscanf(parts[2], "i=%d,l=%d", &p.Iterations, &p.KeyLength)
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

// decodeScryptHash decodes Scrypt encoded password hash.
// format: $scrypt$ln=<cost>,r=<block>,p=<parrrelization>$<salt>$<hash>
func decodeScryptHash(encodedHash string) (p *Scrypt, salt, hash []byte, err error) {
	parts := strings.Split(encodedHash, "$")
	if len(parts) != 5 {
		return nil, nil, nil, ErrInvalidHash
	}

	p = new(Scrypt)

	_, err = fmt.Sscanf(parts[2], "ln=%d,r=%d,p=%d", &p.Cost, &p.Block, &p.Parrellization)
	if err != nil {
		return nil, nil, nil, err
	}

	salt, err = base64.StdEncoding.Strict().DecodeString(parts[3])
	if err != nil {
		return nil, nil, nil, err
	}
	p.SaltLength = uint32(len(salt))

	hash, err = base64.StdEncoding.Strict().DecodeString(parts[4])
	if err != nil {
		return nil, nil, nil, err
	}
	p.KeyLength = uint32(len(hash))

	return p, salt, hash, nil
}

// decodeSHAHash decodes SHA[1|256|512] encoded password hash in custom PHC format.
// format: $sha1$pf=<salting-format>$<salt>$<hash>
func decodeSHAHash(encodedHash string) (hasher string, pf, salt, hash []byte, err error) {
	parts := strings.Split(encodedHash, "$")

	if len(parts) != 5 {
		return "", nil, nil, nil, ErrInvalidHash
	}

	hasher = parts[1]

	_, err = fmt.Sscanf(parts[2], "pf=%s", &pf)
	if err != nil {
		return "", nil, nil, nil, err
	}

	pf, err = base64.StdEncoding.Strict().DecodeString(string(pf))
	if err != nil {
		return "", nil, nil, nil, err
	}

	salt, err = base64.StdEncoding.Strict().DecodeString(parts[3])
	if err != nil {
		return "", nil, nil, nil, err
	}

	hash, err = base64.StdEncoding.Strict().DecodeString(parts[4])
	if err != nil {
		return "", nil, nil, nil, err
	}

	return hasher, pf, salt, hash, nil
}

// used for CompareSHA and CompareSSHA
func compareSHAHelper(hasher string, raw []byte, hash []byte) error {

	var sha []byte

	switch hasher {
	case "sha1":
		sum := sha1.Sum(raw) // #nosec G401 - compatibility for imported passwords
		sha = sum[:]
	case "sha256":
		sum := sha256.Sum256(raw)
		sha = sum[:]
	case "sha512":
		sum := sha512.Sum512(raw)
		sha = sum[:]
	default:
		return errors.WithStack(ErrMismatchedHashAndPassword)
	}

	encodedHash := []byte(base64.StdEncoding.EncodeToString(hash))
	newEncodedHash := []byte(base64.StdEncoding.EncodeToString(sha))

	// Check that the contents of the hashed passwords are identical.
	// subtle.ConstantTimeCompare() is used to help prevent timing attacks.
	if subtle.ConstantTimeCompare(encodedHash, newEncodedHash) == 1 {
		return nil
	}
	return errors.WithStack(ErrMismatchedHashAndPassword)
}

// decodeSSHAHash decodes SSHA[1|256|512] encoded password hash in usual {SSHA...} format.
func decodeSSHAHash(encodedHash string) (hasher string, salt, hash []byte, err error) {
	re := regexp.MustCompile(`\{([^}]*)\}`)
	match := re.FindStringSubmatch(string(encodedHash))

	var index_of_salt_begin int
	var index_of_hash_begin int

	switch match[1] {
	case "SSHA":
		hasher = "sha1"
		index_of_hash_begin = 6
		index_of_salt_begin = 20

	case "SSHA256":
		hasher = "sha256"
		index_of_hash_begin = 9
		index_of_salt_begin = 32

	case "SSHA512":
		hasher = "sha512"
		index_of_hash_begin = 9
		index_of_salt_begin = 64

	default:
		return "", nil, nil, ErrInvalidHash
	}

	decoded, err := base64.StdEncoding.DecodeString(string(encodedHash[index_of_hash_begin:]))
	if err != nil {
		return "", nil, nil, ErrInvalidHash
	}

	if len(decoded) < index_of_salt_begin+1 {
		return "", nil, nil, ErrInvalidHash
	}

	salt = decoded[index_of_salt_begin:]
	hash = decoded[:index_of_salt_begin]

	return hasher, salt, hash, nil
}

// decodeFirebaseScryptHash decodes Firebase Scrypt encoded password hash.
// format: $firescrypt$ln=<mem_cost>,r=<rounds>,p=<parallelization>$<salt>$<hash>$<salt_separator>$<signer_key>
func decodeFirebaseScryptHash(encodedHash string) (p *Scrypt, salt, saltSeparator, hash, signerKey []byte, err error) {
	parts := strings.Split(encodedHash, "$")
	if len(parts) != 7 {
		return nil, nil, nil, nil, nil, ErrInvalidHash
	}

	p = new(Scrypt)

	_, err = fmt.Sscanf(parts[2], "ln=%d,r=%d,p=%d", &p.Cost, &p.Block, &p.Parrellization)
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}
	// convert from firebase config "mem_cost" to
	// scrypt CPU/memory cost parameter, which must be a power of two greater than 1.
	p.Cost = 1 << p.Cost

	salt, err = base64.StdEncoding.Strict().DecodeString(parts[3])
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}
	p.SaltLength = uint32(len(salt))

	hash, err = base64.StdEncoding.Strict().DecodeString(parts[4])
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}
	// Are all firebase script hashes of length 32?
	p.KeyLength = 32

	saltSeparator, err = base64.StdEncoding.Strict().DecodeString(parts[5])
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}

	signerKey, err = base64.StdEncoding.Strict().DecodeString(parts[6])
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}

	return p, salt, saltSeparator, hash, signerKey, nil
}

// decodeMD5Hash decodes MD5 encoded password hash.
// format without salt: $md5$<hash>
// format with salt $md5$pf=<salting-format>$<salt>$<hash>
func decodeMD5Hash(encodedHash string) (pf, salt, hash []byte, err error) {
	parts := strings.Split(encodedHash, "$")

	switch len(parts) {
	case 3:
		hash, err := base64.StdEncoding.Strict().DecodeString(parts[2])
		return nil, nil, hash, err
	case 5:
		_, err = fmt.Sscanf(parts[2], "pf=%s", &pf)
		if err != nil {
			return nil, nil, nil, err
		}

		pf, err := base64.StdEncoding.Strict().DecodeString(string(pf))
		if err != nil {
			return nil, nil, nil, err
		}

		salt, err = base64.StdEncoding.Strict().DecodeString(parts[3])
		if err != nil {
			return nil, nil, nil, err
		}

		hash, err = base64.StdEncoding.Strict().DecodeString(parts[4])
		if err != nil {
			return nil, nil, nil, err
		}

		return pf, salt, hash, nil
	default:
		return nil, nil, nil, ErrInvalidHash
	}
}
