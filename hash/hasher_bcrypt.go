package hash

import (
	"context"

	"golang.org/x/crypto/bcrypt"

	"github.com/ory/kratos/schema/errors"
)

type Bcrypt struct {
	c BcryptConfiguration
}

type BcryptConfiguration interface {
	ConfigProvider
}

func NewHasherBcrypt(c BcryptConfiguration) *Bcrypt {
	return &Bcrypt{c: c}
}

func (h *Bcrypt) Generate(ctx context.Context, password []byte) ([]byte, error) {
	if err := validateBcryptPasswordLength(password); err != nil {
		return nil, err
	}

	hash, err := bcrypt.GenerateFromPassword(password, int(h.c.HashConfig(ctx).HasherBcrypt().Cost))
	if err != nil {
		return nil, err
	}

	return hash, nil
}

func validateBcryptPasswordLength(password []byte) error {
	// Bcrypt truncates the password to the first 72 bytes, following the OpenBSD implementation,
	// so if password is longer than 72 bytes, function returns an error
	// See https://en.wikipedia.org/wiki/Bcrypt#User_input
	if len(password) > 72 {
		return errors.NewPasswordPolicyViolationError(
			"#/password",
			"passwords are limited to a maximum length of 72 characters",
		)
	}
	return nil
}

func (h *Bcrypt) Understands(hash []byte) bool {
	return IsBcryptHash(hash)
}
