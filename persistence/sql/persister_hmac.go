package sql

import (
	"context"
	"crypto/hmac"
	"crypto/sha512"
	"crypto/subtle"
	"fmt"
)

func (p *Persister) hmacValue(ctx context.Context, value string) string {
	return p.hmacValueWithSecret(value, p.r.Config(ctx).SecretsSession()[0])
}

func (p *Persister) hmacValueWithSecret(value string, secret []byte) string {
	h := hmac.New(sha512.New512_256, secret)
	_, _ = h.Write([]byte(value))
	return fmt.Sprintf("%x", h.Sum(nil))
}

func (p *Persister) hmacConstantCompare(ctx context.Context, value, hash string) bool {
	for _, secret := range p.r.Config(ctx).SecretsSession() {
		if subtle.ConstantTimeCompare([]byte(p.hmacValueWithSecret(value, secret)), []byte(hash)) == 1 {
			return true
		}
	}
	return false
}
