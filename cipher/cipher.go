package cipher

import "context"

// Cipher provides methods for encrypt and decrypt string
type Cipher interface {
	// Encrypt encrypts the (binary) message and returns a hex-encoded binary ciphertext
	// or an error if the encryption failed.
	//
	// If the message is empty, the ciphertext is also empty and no error is returned.
	Encrypt(ctx context.Context, message []byte) (string, error)

	// Decrypt takes a hex-encoded binary ciphertext and decrypts it or returns an error if the decryption
	// failed.
	//
	// If the ciphertext is empty a nil byte slice is returned.
	Decrypt(ctx context.Context, encrypted string) ([]byte, error)
}

type Provider interface {
	Cipher() Cipher
}
