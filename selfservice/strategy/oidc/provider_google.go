package oidc

import (
	"context"

	"golang.org/x/oauth2"
)

type ProviderGoogle struct {
	*ProviderGenericOIDC
}

func NewProviderGoogle(
	config *Configuration,
	reg dependencies,
) *ProviderGoogle {
	config.IssuerURL = "https://accounts.google.com"
	return &ProviderGoogle{
		ProviderGenericOIDC: &ProviderGenericOIDC{
			config: config,
			reg:    reg,
		},
	}
}

func (g *ProviderGoogle) Exchange(ctx context.Context, code string) (*oauth2.Token, error) {
	conf, err := g.OAuth2(ctx)
	if err != nil {
		return nil, err
	}

	return conf.Exchange(ctx, code)
}
