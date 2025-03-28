// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package hash

import (
	"bytes"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/md5"  //nolint:all // System compatibility for imported passwords
	"crypto/sha1" //nolint:all // System compatibility for imported passwords
	"crypto/sha256"
	"crypto/sha512"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"hash"
	"regexp"
	"strings"

	"github.com/go-crypt/crypt"
	"github.com/go-crypt/crypt/algorithm/md5crypt"
	"github.com/go-crypt/crypt/algorithm/shacrypt"
	"github.com/pkg/errors"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/bcrypt"

	//nolint:staticcheck
	//lint:ignore SA1019 compatibility for imported passwords
	"golang.org/x/crypto/md4" //nolint:gosec // disable G115 G501 -- compatibility for imported passwords
	"golang.org/x/crypto/pbkdf2"
	"golang.org/x/crypto/scrypt"

	"github.com/ory/kratos/driver/config"
)

var ErrUnknownHashAlgorithm = errors.New("unknown hash algorithm")

func NewCryptDecoder() *crypt.Decoder {
	decoder := crypt.NewDecoder()

	// The register function only returns an error if the decoder is nil or the algorithm is already registered.
	// This is never the case here, if it is, we did something horribly wrong.
	if err := md5crypt.RegisterDecoderCommon(decoder); err != nil {
		panic(err)
	}

	if err := shacrypt.RegisterDecoder(decoder); err != nil {
		panic(err)
	}

	return decoder
}

var CryptDecoder = NewCryptDecoder()

type SupportedHasher struct {
	Comparator func(ctx context.Context, password, hash []byte) error
	Name       string
	Is         func(hash []byte) bool
}

func AddSupportedHasher(s SupportedHasher) {
	supportedHashers = append(supportedHashers, s)
}

var supportedHashers = []SupportedHasher{
	{
		Comparator: CompareMD5Crypt,
		Name:       "md5crypt",
		Is:         IsMD5CryptHash,
	},
	{
		Comparator: CompareBcrypt,
		Name:       "bcrypt",
		Is:         IsBcryptHash,
	},
	{
		Comparator: CompareSHA256Crypt,
		Name:       "sha256crypt",
		Is:         IsSHA256CryptHash,
	},
	{
		Comparator: CompareSHA512Crypt,
		Name:       "sha512crypt",
		Is:         IsSHA512CryptHash,
	},
	{
		Comparator: CompareArgon2id,
		Name:       "argon2id",
		Is:         IsArgon2idHash,
	},
	{
		Comparator: CompareArgon2i,
		Name:       "argon2i",
		Is:         IsArgon2iHash,
	},
	{
		Comparator: ComparePbkdf2,
		Name:       "pbkdf2",
		Is:         IsPbkdf2Hash,
	},
	{
		Comparator: CompareScrypt,
		Name:       "scrypt",
		Is:         IsScryptHash,
	},
	{
		Comparator: CompareSSHA,
		Name:       "ssha",
		Is:         IsSSHAHash,
	},
	{
		Comparator: CompareSHA,
		Name:       "sha",
		Is:         IsSHAHash,
	},
	{
		Comparator: CompareFirebaseScrypt,
		Name:       "firebasescrypt",
		Is:         IsFirebaseScryptHash,
	},
	{
		Comparator: CompareMD5,
		Name:       "md5",
		Is:         IsMD5Hash,
	},
	{
		Comparator: CompareHMAC,
		Name:       "hmac",
		Is:         IsHMACHash,
	},
}

func Compare(ctx context.Context, password, hash []byte) error {
	ctx, span := otel.GetTracerProvider().Tracer(tracingComponent).Start(ctx, "hash.Compare")
	defer span.End()

	for _, h := range supportedHashers {
		if h.Is(hash) {
			span.SetAttributes(attribute.String("hash.type", h.Name))
			return h.Comparator(ctx, password, hash)
		}
	}

	span.SetAttributes(attribute.String("hash.type", "unknown"))
	return errors.WithStack(ErrUnknownHashAlgorithm)
}

func CompareMD5Crypt(_ context.Context, password, hash []byte) error {
	// the password has successfully been validated (has prefix `$md5-crypt`),
	// the decoder expect the module crypt identifier instead (`$1`), which means we need to replace the prefix
	// before decoding
	hash = bytes.TrimPrefix(hash, []byte("$md5-crypt"))
	hash = append([]byte("$1"), hash...)

	return compareCryptHelper(password, string(hash))
}

func CompareBcrypt(ctx context.Context, password, hash []byte) error {
	if err := validateBcryptPasswordLength(password); err != nil {
		return err
	}

	// ensure that the context is not canceled before doing the heavy lifting
	if ctx.Err() != nil {
		return ctx.Err()
	}

	err := bcrypt.CompareHashAndPassword(hash, password)
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return errors.WithStack(ErrMismatchedHashAndPassword)
		}
		return err
	}

	return nil
}

func CompareSHA256Crypt(_ context.Context, password, hash []byte) error {
	hash = bytes.TrimPrefix(hash, []byte("$sha256-crypt"))
	hash = append([]byte("$5"), hash...)

	return compareCryptHelper(password, string(hash))
}

func CompareSHA512Crypt(_ context.Context, password, hash []byte) error {
	hash = bytes.TrimPrefix(hash, []byte("$sha512-crypt"))
	hash = append([]byte("$6"), hash...)

	return compareCryptHelper(password, string(hash))
}

func CompareArgon2id(ctx context.Context, password, hash []byte) error {
	// Extract the parameters, salt and derived key from the encoded password
	// hash.
	p, salt, hash, err := decodeArgon2idHash(string(hash))
	if err != nil {
		return err
	}

	// ensure that the context is not canceled before doing the heavy lifting
	if ctx.Err() != nil {
		return ctx.Err()
	}

	// Derive the key from the other password using the same parameters.
	//nolint:gosec // disable G115
	otherHash := argon2.IDKey(password, salt, p.Iterations, uint32(p.Memory), p.Parallelism, p.KeyLength)

	return comparePasswordHashConstantTime(hash, otherHash)
}

func CompareArgon2i(ctx context.Context, password, hash []byte) error {
	// Extract the parameters, salt and derived key from the encoded password
	// hash.
	p, salt, hash, err := decodeArgon2idHash(string(hash))
	if err != nil {
		return err
	}

	// ensure that the context is not canceled before doing the heavy lifting
	if ctx.Err() != nil {
		return ctx.Err()
	}

	// Derive the key from the other password using the same parameters.
	otherHash := argon2.Key(password, salt, p.Iterations, uint32(p.Memory), p.Parallelism, p.KeyLength)

	return comparePasswordHashConstantTime(hash, otherHash)
}

func ComparePbkdf2(ctx context.Context, password, hash []byte) error {
	// Extract the parameters, salt and derived key from the encoded password
	// hash.
	p, salt, hash, err := decodePbkdf2Hash(string(hash))
	if err != nil {
		return err
	}

	// ensure that the context is not canceled before doing the heavy lifting
	if ctx.Err() != nil {
		return ctx.Err()
	}

	// Derive the key from the other password using the same parameters.
	otherHash := pbkdf2.Key(password, salt, int(p.Iterations), int(p.KeyLength), getPseudorandomFunctionForPbkdf2(p.Algorithm))

	return comparePasswordHashConstantTime(hash, otherHash)
}

func CompareScrypt(ctx context.Context, password, hash []byte) error {
	// Extract the parameters, salt and derived key from the encoded password
	// hash.
	p, salt, hash, err := decodeScryptHash(string(hash))
	if err != nil {
		return err
	}

	// ensure that the context is not canceled before doing the heavy lifting
	if ctx.Err() != nil {
		return ctx.Err()
	}

	// Derive the key from the other password using the same parameters.
	otherHash, err := scrypt.Key(password, salt, int(p.Cost), int(p.Block), int(p.Parrellization), int(p.KeyLength))
	if err != nil {
		return errors.WithStack(err)
	}

	return comparePasswordHashConstantTime(hash, otherHash)
}

func CompareSSHA(ctx context.Context, password, hash []byte) error {
	hasher, salt, hash, err := decodeSSHAHash(string(hash))
	if err != nil {
		return err
	}

	raw := append(password[:], salt[:]...)

	// ensure that the context is not canceled before doing the heavy lifting
	if ctx.Err() != nil {
		return ctx.Err()
	}

	return CompareSHAHelper(hasher, raw, hash)
}

func CompareSHA(ctx context.Context, password, hash []byte) error {
	hasher, pf, salt, hash, err := decodeSHAHash(string(hash))
	if err != nil {
		return err
	}

	r := strings.NewReplacer("{SALT}", string(salt), "{PASSWORD}", string(password))
	raw := []byte(r.Replace(string(pf)))

	// ensure that the context is not canceled before doing the heavy lifting
	if ctx.Err() != nil {
		return ctx.Err()
	}

	return CompareSHAHelper(hasher, raw, hash)
}

func CompareFirebaseScrypt(ctx context.Context, password, hash []byte) error {
	// Extract the parameters, salt and derived key from the encoded password
	// hash.
	p, salt, saltSeparator, hash, signerKey, err := decodeFirebaseScryptHash(string(hash))
	if err != nil {
		return err
	}

	// ensure that the context is not canceled before doing the heavy lifting
	if ctx.Err() != nil {
		return ctx.Err()
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

	return comparePasswordHashConstantTime(hash, otherHash)
}

func CompareMD5(ctx context.Context, password, hash []byte) error {
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

	// ensure that the context is not canceled before doing the heavy lifting
	if ctx.Err() != nil {
		return ctx.Err()
	}

	//#nosec G401 -- compatibility for imported passwords
	otherHash := md5.Sum(arg)

	return comparePasswordHashConstantTime(hash, otherHash[:])
}

func CompareHMAC(ctx context.Context, password, hash []byte) error {
	// Extract the hash from the encoded password
	hasher, hash, key, err := decodeHMACHash(string(hash))
	if err != nil {
		return err
	}

	// ensure that the context is not canceled before doing the heavy lifting
	if ctx.Err() != nil {
		return ctx.Err()
	}

	mac := hmac.New(hasher, key)
	_, err = mac.Write(password)
	if err != nil {
		return err
	}

	otherHash := []byte(hex.EncodeToString(mac.Sum(nil)))

	return comparePasswordHashConstantTime(hash, otherHash)
}

var (
	isMD5CryptHash       = regexp.MustCompile(`^\$md5-crypt\$`)
	isBcryptHash         = regexp.MustCompile(`^\$2[abzy]?\$`)
	isSHA256CryptHash    = regexp.MustCompile(`^\$sha256-crypt\$`)
	isSHA512CryptHash    = regexp.MustCompile(`^\$sha512-crypt\$`)
	isArgon2idHash       = regexp.MustCompile(`^\$argon2id\$`)
	isArgon2iHash        = regexp.MustCompile(`^\$argon2i\$`)
	isPbkdf2Hash         = regexp.MustCompile(`^\$pbkdf2-sha[0-9]{1,3}\$`)
	isScryptHash         = regexp.MustCompile(`^\$scrypt\$`)
	isSSHAHash           = regexp.MustCompile(`^{SSHA(256|512)?}.*`)
	isSHAHash            = regexp.MustCompile(`^\$sha(1|256|512)\$`)
	isFirebaseScryptHash = regexp.MustCompile(`^\$firescrypt\$`)
	isMD5Hash            = regexp.MustCompile(`^\$md5\$`)
	isHMACHash           = regexp.MustCompile(`^\$hmac-(md4|md5|sha1|sha224|sha256|sha384|sha512)\$`)
)

func IsMD5CryptHash(hash []byte) bool       { return isMD5CryptHash.Match(hash) }
func IsBcryptHash(hash []byte) bool         { return isBcryptHash.Match(hash) }
func IsSHA256CryptHash(hash []byte) bool    { return isSHA256CryptHash.Match(hash) }
func IsSHA512CryptHash(hash []byte) bool    { return isSHA512CryptHash.Match(hash) }
func IsArgon2idHash(hash []byte) bool       { return isArgon2idHash.Match(hash) }
func IsArgon2iHash(hash []byte) bool        { return isArgon2iHash.Match(hash) }
func IsPbkdf2Hash(hash []byte) bool         { return isPbkdf2Hash.Match(hash) }
func IsScryptHash(hash []byte) bool         { return isScryptHash.Match(hash) }
func IsSSHAHash(hash []byte) bool           { return isSSHAHash.Match(hash) }
func IsSHAHash(hash []byte) bool            { return isSHAHash.Match(hash) }
func IsFirebaseScryptHash(hash []byte) bool { return isFirebaseScryptHash.Match(hash) }
func IsMD5Hash(hash []byte) bool            { return isMD5Hash.Match(hash) }
func IsHMACHash(hash []byte) bool           { return isHMACHash.Match(hash) }

func IsValidHashFormat(hash []byte) bool {
	for _, h := range supportedHashers {
		if h.Is(hash) {
			return true
		}
	}

	return false
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

// CompareSHAHelper compares the raw password with the hash using the given hasher.
func CompareSHAHelper(hasher string, raw []byte, hash []byte) error {
	var sha []byte

	switch hasher {
	case "sha1":
		sum := sha1.Sum(raw) //#nosec G401 -- compatibility for imported passwords
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

	return comparePasswordHashConstantTime(encodedHash, newEncodedHash)
}

func compareCryptHelper(password []byte, hash string) error {
	digest, err := CryptDecoder.Decode(hash)
	if err != nil {
		return err
	}

	if digest.MatchBytes(password) {
		return nil
	}

	return errors.WithStack(ErrMismatchedHashAndPassword)
}

var regexSSHA = regexp.MustCompile(`\{([^}]*)\}`)

// decodeSSHAHash decodes SSHA[1|256|512] encoded password hash in usual {SSHA...} format.
func decodeSSHAHash(encodedHash string) (hasher string, salt, hash []byte, err error) {
	match := regexSSHA.FindStringSubmatch(string(encodedHash))

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

// decodeHMACHash decodes HMAC encoded password hash.
// format : $hmac-<hash function>$<hash>$<key>
func decodeHMACHash(encodedHash string) (hasher func() hash.Hash, hash, key []byte, err error) {
	parts := strings.Split(encodedHash, "$")

	if len(parts) != 4 {
		return nil, nil, nil, ErrInvalidHash
	}

	hashMatch := isHMACHash.FindStringSubmatch(encodedHash)

	if len(hashMatch) != 2 {
		return nil, nil, nil, errors.WithStack(ErrUnknownHashAlgorithm)
	}

	switch hashMatch[1] {
	case "md4":
		hasher = md4.New //#nosec G401 -- compatibility for imported passwords
	case "md5":
		hasher = md5.New //#nosec G401 -- compatibility for imported passwords
	case "sha1":
		hasher = sha1.New //#nosec G401 -- compatibility for imported passwords
	case "sha224":
		hasher = sha256.New224
	case "sha256":
		hasher = sha256.New
	case "sha384":
		hasher = sha512.New384
	case "sha512":
		hasher = sha512.New
	default:
		return nil, nil, nil, errors.WithStack(ErrUnknownHashAlgorithm)
	}

	hash, err = base64.StdEncoding.Strict().DecodeString(parts[2])
	if err != nil {
		return nil, nil, nil, err
	}

	key, err = base64.StdEncoding.Strict().DecodeString(parts[3])
	if err != nil {
		return nil, nil, nil, err
	}

	return hasher, hash, key, nil
}

func comparePasswordHashConstantTime(hash, otherHash []byte) error {
	// use subtle.ConstantTimeCompare() to prevent timing attacks.
	if subtle.ConstantTimeCompare(hash, otherHash) == 1 {
		return nil
	}
	return errors.WithStack(ErrMismatchedHashAndPassword)
}
