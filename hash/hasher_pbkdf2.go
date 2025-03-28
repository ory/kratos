// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package hash

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha1" //#nosec G505 -- compatibility for imported passwords
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"fmt"
	"hash"

	"github.com/pkg/errors"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/crypto/pbkdf2"
	"golang.org/x/crypto/sha3"
)

type Pbkdf2 struct {
	Algorithm  string
	Iterations uint32
	SaltLength uint32
	KeyLength  uint32
}

func (h *Pbkdf2) Generate(ctx context.Context, password []byte) ([]byte, error) {
	_, span := otel.GetTracerProvider().Tracer(tracingComponent).Start(ctx, "hash.Generate", trace.WithAttributes(
		attribute.String("hash.type", "pbkdf2"),
		attribute.String("hash.config", fmt.Sprintf("%#v", h)),
	))
	defer span.End()

	salt := make([]byte, h.SaltLength)
	if _, err := rand.Read(salt); err != nil {
		return nil, err
	}

	// ensure that the context is not canceled before doing the heavy lifting
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	key := pbkdf2.Key(password, salt, int(h.Iterations), int(h.KeyLength), getPseudorandomFunctionForPbkdf2(h.Algorithm))

	var b bytes.Buffer
	if _, err := fmt.Fprintf(
		&b,
		"$pbkdf2-%s$i=%d,l=%d$%s$%s",
		h.Algorithm,
		h.Iterations,
		h.KeyLength,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(key),
	); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, errors.WithStack(err)
	}

	return b.Bytes(), nil
}

func (h *Pbkdf2) Understands(hash []byte) bool {
	return IsPbkdf2Hash(hash)
}

func getPseudorandomFunctionForPbkdf2(alg string) func() hash.Hash {
	switch alg {
	case "sha1":
		return sha1.New
	case "sha224":
		return sha3.New224
	case "sha256":
		return sha256.New
	case "sha384":
		return sha3.New384
	case "sha512":
		return sha512.New
	default:
		return sha256.New
	}
}
