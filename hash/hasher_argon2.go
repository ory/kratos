// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package hash

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"

	"github.com/inhies/go-bytesize"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"github.com/pkg/errors"
	"golang.org/x/crypto/argon2"

	"github.com/ory/kratos/driver/config"
)

var (
	ErrInvalidHash               = errors.New("the encoded hash is not in the correct format")
	ErrIncompatibleVersion       = errors.New("incompatible version of argon2")
	ErrMismatchedHashAndPassword = errors.New("passwords do not match")
)

type Argon2 struct {
	c Argon2Configuration
}

type Argon2Configuration interface {
	config.Provider
}

func NewHasherArgon2(c Argon2Configuration) *Argon2 {
	return &Argon2{c: c}
}

func toKB(mem bytesize.ByteSize) uint32 {
	//nolint:gosec // disable G115
	return uint32(mem / bytesize.KB)
}

func (h *Argon2) Generate(ctx context.Context, password []byte) ([]byte, error) {
	conf := h.c.Config().HasherArgon2(ctx)

	_, span := otel.GetTracerProvider().Tracer(tracingComponent).Start(ctx, "hash.Generate", trace.WithAttributes(
		attribute.String("hash.type", "argon2id"),
		attribute.String("hash.config", fmt.Sprintf("%#v", conf)),
	))
	defer span.End()

	salt := make([]byte, conf.SaltLength)
	if _, err := rand.Read(salt); err != nil {
		return nil, err
	}

	// ensure that the context is not canceled before doing the heavy lifting
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	// Pass the plaintext password, salt and parameters to the argon2.IDKey
	// function. This will generate a hash of the password using the Argon2id
	// variant.
	hash := argon2.IDKey(password, salt, conf.Iterations, toKB(conf.Memory), conf.Parallelism, conf.KeyLength)

	var b bytes.Buffer
	if _, err := fmt.Fprintf(
		&b,
		"$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version, toKB(conf.Memory), conf.Iterations, conf.Parallelism,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(hash),
	); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, errors.WithStack(err)
	}

	return b.Bytes(), nil
}

func (h *Argon2) Understands(hash []byte) bool {
	return IsArgon2idHash(hash)
}
