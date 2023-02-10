// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package oidc

import (
	"context"

	gooidc "github.com/coreos/go-oidc"
	"golang.org/x/oauth2"

	"github.com/ory/x/stringslice"
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
