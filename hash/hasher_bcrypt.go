// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package hash

import (
	"context"
	"fmt"

	"github.com/ory/kratos/text"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

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
	conf := h.c.Config().HasherBcrypt(ctx)

	_, span := otel.GetTracerProvider().Tracer(tracingComponent).Start(ctx, "hash.Generate", trace.WithAttributes(
		attribute.String("hash.type", "bcrypt"),
		attribute.String("hash.config", fmt.Sprintf("%#v", conf)),
	))
	defer span.End()

	if err := validateBcryptPasswordLength(password); err != nil {
		return nil, err
	}

	// ensure that the context is not canceled before doing the heavy lifting
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	hash, err := bcrypt.GenerateFromPassword(password, int(conf.Cost))
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
		return schema.NewPasswordPolicyViolationError(
			"#/password",
			text.NewErrorValidationPasswordMaxLength(72, len(password)),
		)
	}
	return nil
}

func (h *Bcrypt) Understands(hash []byte) bool {
	return IsBcryptHash(hash)
}
