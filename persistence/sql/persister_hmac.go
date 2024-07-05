// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package sql

import (
	"context"
	"crypto/hmac"
	"crypto/sha512"
	"fmt"

	"go.opentelemetry.io/otel/trace"
)

func (p *Persister) hmacValue(ctx context.Context, value string) string {
	return hmacValueWithSecret(ctx, value, p.r.Config().SecretsSession(ctx)[0])
}

func hmacValueWithSecret(ctx context.Context, value string, secret []byte) string {
	_, span := trace.SpanFromContext(ctx).TracerProvider().Tracer("").Start(ctx, "persistence.sql.hmacValueWithSecret")
	defer span.End()
	h := hmac.New(sha512.New512_256, secret)
	_, _ = h.Write([]byte(value))
	return fmt.Sprintf("%x", h.Sum(nil))
}
