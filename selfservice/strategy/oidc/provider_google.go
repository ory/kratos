// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package oidc

import (
	"context"

	gooidc "github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"

	"github.com/ory/x/stringslice"
)

var _ OAuth2Provider = (*ProviderGoogle)(nil)

type ProviderGoogle struct {
	*ProviderGenericOIDC
	JWKSUrl string
}

func NewProviderGoogle(
	config *Configuration,
	reg Dependencies,
) Provider {
	config.IssuerURL = "https://accounts.google.com"
	return &ProviderGoogle{
		ProviderGenericOIDC: &ProviderGenericOIDC{
			config: config,
			reg:    reg,
		},
		JWKSUrl: "https://www.googleapis.com/oauth2/v3/certs",
	}
}

func (g *ProviderGoogle) oauth2ConfigFromEndpoint(ctx context.Context, endpoint oauth2.Endpoint) *oauth2.Config {
	scope := g.config.Scope
	if !stringslice.Has(scope, gooidc.ScopeOpenID) {
		scope = append(scope, gooidc.ScopeOpenID)
	}

	scope = stringslice.Filter(scope, func(s string) bool { return s == gooidc.ScopeOfflineAccess })

	return &oauth2.Config{
		ClientID:     g.config.ClientID,
		ClientSecret: g.config.ClientSecret,
		Endpoint:     endpoint,
		Scopes:       scope,
		RedirectURL:  g.config.Redir(g.reg.Config().OIDCRedirectURIBase(ctx)),
	}
}

func (g *ProviderGoogle) OAuth2(ctx context.Context) (*oauth2.Config, error) {
	p, err := g.provider(ctx)
	if err != nil {
		return nil, err
	}

	endpoint := p.Endpoint()
	return g.oauth2ConfigFromEndpoint(ctx, endpoint), nil
}

func (g *ProviderGoogle) AuthCodeURLOptions(r ider) []oauth2.AuthCodeOption {
	scope := g.config.Scope
	options := g.ProviderGenericOIDC.AuthCodeURLOptions(r)

	if stringslice.Has(scope, gooidc.ScopeOfflineAccess) {
		options = append(options, oauth2.AccessTypeOffline)
	}

	return options
}

var _ IDTokenVerifier = new(ProviderGoogle)

const issuerUrlGoogle = "https://accounts.google.com"

func (p *ProviderGoogle) Verify(ctx context.Context, rawIDToken string) (*Claims, error) {
	keySet := gooidc.NewRemoteKeySet(ctx, p.JWKSUrl)
	ctx = gooidc.ClientContext(ctx, p.reg.HTTPClient(ctx).HTTPClient)

	return verifyToken(ctx, keySet, p.config, rawIDToken, issuerUrlGoogle)
}

var _ NonceValidationSkipper = new(ProviderGoogle)

func (a *ProviderGoogle) CanSkipNonce(c *Claims) bool {
	// Not all SDKs support nonce validation, so we skip it if no nonce is present in the claims of the ID Token.
	return c.Nonce == ""
}
