package crypt

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"log"

	"github.com/ory/herodot"

	"github.com/pkg/errors"

	"github.com/ory/kratos/driver/config"
)

type AES struct {
	c AESConfiguration
}

type AESConfiguration interface {
	config.Provider
}

func NewCryptAES(c AESConfiguration) *AES {
	return &AES{c: c}
}

// Encrypt return a AES encrypt of plaintext
func (a *AES) Encrypt(ctx context.Context, clearString string) (string, error) {
	if len(clearString) == 0 {
		return "", errors.WithStack(herodot.ErrInternalServerError.WithReason("Expected plaintext string does not empty"))
	}

	p := a.c.Config(ctx).CryptAES()

	//Since the key is in string, we need to convert decode it to bytes
	key := []byte(p.Secret)

	plaintext := []byte(clearString)

	//Create a new Cipher Block from the key
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	//Create a new GCM - https://en.wikipedia.org/wiki/Galois/Counter_Mode
	//https://golang.org/pkg/crypto/cipher/#NewGCM
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	//Create a nonce. Nonce should be from GCM
	nonce := make([]byte, aesGCM.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	//Encrypt the data using aesGCM.Seal
	//Since we don't want to save the nonce somewhere else in this case, we add it as a prefix to the encrypted data. The first nonce argument in Seal is the prefix.
	ciphertext := aesGCM.Seal(nonce, nonce, plaintext, nil)
	return fmt.Sprintf("%x", ciphertext), nil
}

// Decrypt returns the decrypted aes data
func (a *AES) Decrypt(ctx context.Context, encryptedString string) (string, error) {
	if len(encryptedString) == 0 {
		return "", errors.WithStack(herodot.ErrInternalServerError.WithReason("Expected encrypted string does not empty"))
	}

	p := a.c.Config(ctx).CryptAES()
	log.Printf("configAES: %+v", p)
	key := []byte(p.Secret)
	enc, _ := hex.DecodeString(encryptedString)

	//Create a new Cipher Block from the key
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	//Create a new GCM
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	//Get the nonce size
	nonceSize := aesGCM.NonceSize()

	//Extract the nonce from the encrypted data
	nonce, ciphertext := enc[:nonceSize], enc[nonceSize:]

	//Decrypt the data
	plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s", plaintext), nil
}
