package oidc

import (
	"context"
	"net/url"

	"github.com/pkg/errors"
	"golang.org/x/oauth2"

	"github.com/ory/herodot"
	"github.com/ory/x/stringslice"

	gooidc "github.com/coreos/go-oidc"

	"github.com/ory/hive/selfservice"
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

func (g *ProviderGenericOIDC) OAuth2(ctx context.Context) (*oauth2.Config, error) {
	p, err := g.provider(ctx)
	if err != nil {
		return nil, err
	}

	endpoint := p.Endpoint()

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
	}, nil
}

func (g *ProviderGenericOIDC) Claims(ctx context.Context, exchange *oauth2.Token) (*Claims, error) {
	raw, ok := exchange.Extra("id_token").(string)
	if !ok || len(raw) == 0 {
		return nil, errors.WithStack(selfservice.ErrIDTokenMissing)
	}

	p, err := g.provider(ctx)
	if err != nil {
		return nil, err
	}

	token, err := p.
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
