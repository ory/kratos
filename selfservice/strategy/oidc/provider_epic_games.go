// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package oidc

import (
	"context"
	"net/url"

	"golang.org/x/oauth2"
)

type ProviderEpicGames struct {
	*ProviderGenericOIDC
}

type EpicGamesIdentityResponse struct {
	Data struct {
		Sub string `json:"sub"`
	} `json:"data"`
}

func NewProviderEpicGames(
	config *Configuration,
	reg Dependencies,
) Provider {
	return &ProviderEpicGames{
		ProviderGenericOIDC: &ProviderGenericOIDC{
			config: config,
			reg:    reg,
		},
	}
}

func (e *ProviderEpicGames) Config() *Configuration {
	return e.config
}

func (e *ProviderEpicGames) oauth2(ctx context.Context) (*oauth2.Config, error) {
	return &oauth2.Config{
		ClientID:     e.config.ClientID,
		ClientSecret: e.config.ClientSecret,
		Scopes:       e.config.Scope,
		Endpoint: oauth2.Endpoint{
			AuthURL:   "https://www.epicgames.com/id/authorize",
			TokenURL:  "https://api.epicgames.dev/epic/oauth/v2/token",
			AuthStyle: oauth2.AuthStyleInHeader,
		},
		RedirectURL: e.config.Redir(e.reg.Config().OIDCRedirectURIBase(ctx)),
	}, nil
}

func (e *ProviderEpicGames) OAuth2(ctx context.Context) (*oauth2.Config, error) {
	e.reg.Logger().WithField("provider", "epic-games").Trace("ProviderCreating new oauth2 configuration in OAuth2 method.")
	return e.oauth2(ctx)
}

func (e *ProviderEpicGames) Claims(ctx context.Context, exchange *oauth2.Token, query url.Values) (*Claims, error) {
	rawClaims := make(map[string]interface{})
	rawClaims["access_token"] = exchange.AccessToken
	rawClaims["refresh_token"] = exchange.RefreshToken
	claims := &Claims{
		Issuer:    "https://api.epicgames.dev/epic/oauth/v2",
		Subject:   exchange.Extra("account_id").(string),
		RawClaims: rawClaims,
	}

	return claims, nil
}

func (e *ProviderEpicGames) AuthCodeURLOptions(r ider) []oauth2.AuthCodeOption {
	return []oauth2.AuthCodeOption{}
}
