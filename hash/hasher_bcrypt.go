// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package hash

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"

	"github.com/ory/kratos/schema"

	"golang.org/x/crypto/bcrypt"

	"github.com/ory/kratos/driver/config"
)

type Bcrypt struct {
	c BcryptConfiguration
}

type BcryptConfiguration interface {
	config.Provider
}

func NewHasherBcrypt(c BcryptConfiguration) *Bcrypt {
	return &Bcrypt{c: c}
}

func (h *Bcrypt) Generate(ctx context.Context, password []byte) ([]byte, error) {
	ctx, span := otel.GetTracerProvider().Tracer(tracingComponent).Start(ctx, "hash.Bcrypt.Generate")
	defer span.End()

	if err := validateBcryptPasswordLength(password); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	cost := int(h.c.Config().HasherBcrypt(ctx).Cost)
	span.SetAttributes(attribute.Int("bcrypt.cost", cost))
	hash, err := bcrypt.GenerateFromPassword(password, cost)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	return hash, nil
}

func validateBcryptPasswordLength(password []byte) error {
	// Bcrypt truncates the password to the first 72 bytes, following the OpenBSD implementation,
	// so if password is longer than 72 bytes, function returns an error
	// See https://en.wikipedia.org/wiki/Bcrypt#User_input
	if len(password) > 72 {
		return schema.NewPasswordPolicyViolationError(
			"#/password",
			"passwords are limited to a maximum length of 72 characters",
		)
	}
	return nil
}

func (h *Bcrypt) Understands(hash []byte) bool {
	return IsBcryptHash(hash)
}
