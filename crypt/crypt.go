package crypt

import "context"

// Crypt provides methods for encrypt and decrypt string
type Crypt interface {
	// Encrypt return a encrypted string from the clearString or an error if the encryption method failed.
	Encrypt(ctx context.Context, clearString string) (string, error)

	// Decrypt return a cleared string from the encrypted string or an error if the decryption method failed.
	Decrypt(ctx context.Context, encryptedString string) (string, error)
}

type CryptProvider interface {
	Crypt() Crypt
}
