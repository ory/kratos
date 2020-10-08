package oidc

import (
	"context"
	"net/url"

	"github.com/pkg/errors"
	"golang.org/x/oauth2"

	"github.com/ory/herodot"
	"github.com/ory/x/stringslice"

	gooidc "github.com/coreos/go-oidc"
)

var _ Provider = new(ProviderGenericOIDC)

type ProviderGenericOIDC struct {
	p      *gooidc.Provider
	config *Configuration
	public *url.URL
}

func NewProviderGenericOIDC(
	config *Configuration,
	public *url.URL,
) *ProviderGenericOIDC {
	return &ProviderGenericOIDC{
		config: config,
		public: public,
	}
}

func (g *ProviderGenericOIDC) Config() *Configuration {
	return g.config
}

func (g *ProviderGenericOIDC) provider(ctx context.Context) (*gooidc.Provider, error) {
	if g.p == nil {
		p, err := gooidc.NewProvider(context.Background(), g.config.IssuerURL)
		if err != nil {
			return nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("Unable to initialize OpenID Connect Provider: %s", err))
		}
		g.p = p
	}
	return g.p, nil
}

func (g *ProviderGenericOIDC) oauth2ConfigFromEndpoint(endpoint oauth2.Endpoint) *oauth2.Config {
	scope := g.config.Scope
	if !stringslice.Has(scope, gooidc.ScopeOpenID) {
		scope = append(scope, gooidc.ScopeOpenID)
	}

	return &oauth2.Config{
		ClientID:     g.config.ClientID,
		ClientSecret: g.config.ClientSecret,
		Endpoint:     endpoint,
		Scopes:       scope,
		RedirectURL:  g.config.Redir(g.public),
	}
}

func (g *ProviderGenericOIDC) OAuth2(ctx context.Context) (*oauth2.Config, error) {
	p, err := g.provider(ctx)
	if err != nil {
		return nil, err
	}

	endpoint := p.Endpoint()

	return g.oauth2ConfigFromEndpoint(endpoint), nil
}

func (g *ProviderGenericOIDC) AuthCodeURLOptions(r ider) []oauth2.AuthCodeOption {
	var options []oauth2.AuthCodeOption

	if isForced(r) {
		options = append(options, oauth2.SetAuthURLParam("prompt", "login"))
	}
	if len(g.config.RequestedClaims) != 0 {
		options = append(options, oauth2.SetAuthURLParam("claims", g.config.RequestedClaims))
	}

	return options
}

func (g *ProviderGenericOIDC) verifyAndDecodeClaimsWithProvider(ctx context.Context, provider *gooidc.Provider, raw string) (*Claims, error) {
	token, err := provider.
		Verifier(&gooidc.Config{
			ClientID: g.config.ClientID,
		}).
		Verify(ctx, raw)
	if err != nil {
		return nil, errors.WithStack(herodot.ErrBadRequest.WithReasonf("%s", err))
	}

	var claims Claims
	if err := token.Claims(&claims); err != nil {
		return nil, errors.WithStack(herodot.ErrBadRequest.WithReasonf("%s", err))
	}

	return &claims, nil
}

func (g *ProviderGenericOIDC) Claims(ctx context.Context, exchange *oauth2.Token) (*Claims, error) {
	raw, ok := exchange.Extra("id_token").(string)
	if !ok || len(raw) == 0 {
		return nil, errors.WithStack(ErrIDTokenMissing)
	}

	p, err := g.provider(ctx)
	if err != nil {
		return nil, err
	}

	return g.verifyAndDecodeClaimsWithProvider(ctx, p, raw)
}
