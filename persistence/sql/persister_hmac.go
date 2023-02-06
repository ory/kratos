// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package sql

import (
	"context"
	"crypto/hmac"
	"crypto/sha512"
	"crypto/subtle"
	"fmt"
)

func (p *Persister) hmacValue(ctx context.Context, value string) string {
	return p.hmacValueWithSecret(ctx, value, p.r.Config().SecretsSession(ctx)[0])
}

func (p *Persister) hmacValueWithSecret(ctx context.Context, value string, secret []byte) string {
	_, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.hmacValueWithSecret")
	defer span.End()
	h := hmac.New(sha512.New512_256, secret)
	_, _ = h.Write([]byte(value))
	return fmt.Sprintf("%x", h.Sum(nil))
}

func (p *Persister) hmacConstantCompare(ctx context.Context, value, hash string) bool {
	for _, secret := range p.r.Config().SecretsSession(ctx) {
		if subtle.ConstantTimeCompare([]byte(p.hmacValueWithSecret(ctx, value, secret)), []byte(hash)) == 1 {
			return true
		}
	}
	return false
}
