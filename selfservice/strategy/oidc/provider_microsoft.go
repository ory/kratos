package oidc

import (
	"context"
	"net/url"

	"github.com/dgrijalva/jwt-go"

	gooidc "github.com/coreos/go-oidc"
	"github.com/ory/herodot"
	"github.com/ory/x/stringslice"
	"github.com/pkg/errors"
	"golang.org/x/oauth2"
)

type ProviderMicrosoft struct {
	config *Configuration
	public *url.URL
}

func NewProviderMicrosoft(
	config *Configuration,
	public *url.URL,
) *ProviderMicrosoft {
	return &ProviderMicrosoft{
		config: config,
		public: public,
	}
}

func (m *ProviderMicrosoft) Config() *Configuration {
	return m.config
}

func (m *ProviderMicrosoft) OAuth2(ctx context.Context) (*oauth2.Config, error) {
	endpointPrefix := "https://login.microsoftonline.com/" + m.config.Tenant
	endpoint := oauth2.Endpoint{
		AuthURL:  endpointPrefix + "/oauth2/v2.0/authorize",
		TokenURL: endpointPrefix + "/oauth2/v2.0/token",
	}

	scope := m.config.Scope
	if !stringslice.Has(scope, gooidc.ScopeOpenID) {
		scope = append(scope, gooidc.ScopeOpenID)
	}

	return &oauth2.Config{
		ClientID:     m.config.ClientID,
		ClientSecret: m.config.ClientSecret,
		Endpoint:     endpoint,
		Scopes:       scope,
		RedirectURL:  m.config.Redir(m.public),
	}, nil
}

func (m *ProviderMicrosoft) AuthCodeURLOptions(r request) []oauth2.AuthCodeOption {
	return []oauth2.AuthCodeOption{}
}

func (m *ProviderMicrosoft) Claims(ctx context.Context, exchange *oauth2.Token) (*Claims, error) {
	raw, ok := exchange.Extra("id_token").(string)
	if !ok || len(raw) == 0 {
		return nil, errors.WithStack(ErrIDTokenMissing)
	}

	parser := new(jwt.Parser)
	unverifiedClaims := microsoftUnverifiedClaims{}
	if _, _, err := parser.ParseUnverified(raw, &unverifiedClaims); err != nil {
		return nil, err
	}

	p, err := gooidc.NewProvider(context.Background(), unverifiedClaims.Issuer)
	if err != nil {
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("Unable to initialize OpenID Connect Provider: %s", err))
	}

	token, err := p.
		Verifier(&gooidc.Config{
			ClientID: m.config.ClientID,
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

type microsoftUnverifiedClaims struct {
	Issuer string `json:"iss,omitempty"`
}

func (c *microsoftUnverifiedClaims) Valid() error {
	return nil
}
