package oidc

import (
	"context"

	"github.com/dghubble/oauth1"

	"github.com/ory/kratos/cipher"
)

type oAuth1Provider interface {
	Claims(ctx context.Context, token *oauth1.Token) (*Claims, error)
	Config() *Configuration
}

type oAuth1Token struct {
	op     oAuth1Provider
	token  *oauth1.Token
	claims *Claims
}

func (o *oAuth1Token) Claims(ctx context.Context) (*Claims, error) {
	if o.claims == nil {
		claims, err := o.op.Claims(ctx, o.token)
		if err != nil {
			return nil, err
		}

		o.claims = claims
	}

	return o.claims, nil
}

func (t *oAuth1Token) CredentialsConfig(ctx context.Context, c cipher.Cipher) (*ProviderCredentialsConfig, error) {
	claims, err := t.Claims(ctx)
	if err != nil {
		return nil, err
	}
	return &ProviderCredentialsConfig{
		Subject:  claims.Subject,
		Provider: t.op.Config().ID,
	}, nil
}
