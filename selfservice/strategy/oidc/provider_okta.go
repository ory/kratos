// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package oidc

import (
	"context"

	gooidc "github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"

	"github.com/ory/x/stringslice"
)

type ProviderOkta struct {
	*ProviderGenericOIDC
	JWKSUrl string
}

func NewProviderOkta(
	config *Configuration,
	reg Dependencies,
) Provider {
	return &ProviderOkta{
		ProviderGenericOIDC: &ProviderGenericOIDC{
			config: config,
			reg:    reg,
		},
		JWKSUrl: config.IssuerURL + "/oauth2/v1/keys",
	}
}

func (o *ProviderOkta) oauth2ConfigFromEndpoint(ctx context.Context, endpoint oauth2.Endpoint) *oauth2.Config {
	scope := o.config.Scope
	if !stringslice.Has(scope, gooidc.ScopeOpenID) {
		scope = append(scope, gooidc.ScopeOpenID)
	}

	return &oauth2.Config{
		ClientID:     o.config.ClientID,
		ClientSecret: o.config.ClientSecret,
		Endpoint:     endpoint,
		Scopes:       scope,
		RedirectURL:  o.config.Redir(o.reg.Config().OIDCRedirectURIBase(ctx)),
	}
}

func (o *ProviderOkta) OAuth2(ctx context.Context) (*oauth2.Config, error) {
	p, err := o.provider(ctx)
	if err != nil {
		return nil, err
	}

	endpoint := p.Endpoint()
	return o.oauth2ConfigFromEndpoint(ctx, endpoint), nil
}

func (o *ProviderOkta) AuthCodeURLOptions(r ider) []oauth2.AuthCodeOption {
	options := o.ProviderGenericOIDC.AuthCodeURLOptions(r)
	return options
}

var _ IDTokenVerifier = new(ProviderOkta)

func (p *ProviderOkta) Verify(ctx context.Context, rawIDToken string) (*Claims, error) {
	keySet := gooidc.NewRemoteKeySet(ctx, p.JWKSUrl)
	ctx = gooidc.ClientContext(ctx, p.reg.HTTPClient(ctx).HTTPClient)
	return verifyToken(ctx, keySet, p.config, rawIDToken, p.config.IssuerURL)
}

var _ NonceValidationSkipper = new(ProviderOkta)

func (a *ProviderOkta) CanSkipNonce(c *Claims) bool {
	// Not all SDKs support nonce validation, so we skip it if no nonce is present in the claims of the ID Token.
	return c.Nonce == ""
}
