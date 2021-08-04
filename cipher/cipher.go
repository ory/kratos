package cipher

import "context"

// Cipher provides methods for encrypt and decrypt string
type Cipher interface {
	// Encrypt return a encrypted string from the clearString or an error if the encryption method failed.
	Encrypt(ctx context.Context, clearString string) ([]byte, error)

	// Decrypt return a cleared string from the encrypted string or an error if the decryption method failed.
	Decrypt(ctx context.Context, encryptedString []byte) (string, error)
}

type Provider interface {
	Cipher() Cipher
}
