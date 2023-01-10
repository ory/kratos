// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package oidc

import (
	"context"
	"encoding/json"
	"net/url"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/pkg/errors"
	"golang.org/x/oauth2"

	"github.com/ory/x/httpx"

	"github.com/ory/herodot"
)

type ProviderYandex struct {
	config *Configuration
	reg    dependencies
}

func NewProviderYandex(
	config *Configuration,
	reg dependencies,
) *ProviderYandex {
	return &ProviderYandex{
		config: config,
		reg:    reg,
	}
}

func (g *ProviderYandex) Config() *Configuration {
	return g.config
}

func (g *ProviderYandex) oauth2(ctx context.Context) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     g.config.ClientID,
		ClientSecret: g.config.ClientSecret,
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://oauth.yandex.com/authorize",
			TokenURL: "https://oauth.yandex.com/token",
		},
		Scopes:      g.config.Scope,
		RedirectURL: g.config.Redir(g.reg.Config().OIDCRedirectURIBase(ctx)),
	}
}

func (g *ProviderYandex) AuthCodeURLOptions(r ider) []oauth2.AuthCodeOption {
	return []oauth2.AuthCodeOption{}
}

func (g *ProviderYandex) OAuth2(ctx context.Context) (*oauth2.Config, error) {
	return g.oauth2(ctx), nil
}

func (g *ProviderYandex) Claims(ctx context.Context, exchange *oauth2.Token, query url.Values) (*Claims, error) {
	o, err := g.OAuth2(ctx)
	if err != nil {
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("%s", err))
	}

	client := g.reg.HTTPClient(ctx, httpx.ResilientClientDisallowInternalIPs(), httpx.ResilientClientWithClient(o.Client(ctx, exchange)))
	req, err := retryablehttp.NewRequest("GET", "https://login.yandex.ru/info?format=json&oauth_token="+exchange.AccessToken, nil)
	if err != nil {
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("%s", err))
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("%s", err))
	}
	defer resp.Body.Close()

	if err := logUpstreamError(g.reg.Logger(), resp); err != nil {
		return nil, err
	}

	var user struct {
		Id           string `json:"id,omitempty"`
		FirstName    string `json:"first_name,omitempty"`
		LastName     string `json:"last_name,omitempty"`
		Email        string `json:"default_email,omitempty"`
		Picture      string `json:"default_avatar_id,omitempty"`
		PictureEmpty bool   `json:"is_avatar_empty,omitempty"`
		Gender       string `json:"sex,omitempty"`
		BirthDay     string `json:"birthday,omitempty"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("%s", err))
	}

	if !user.PictureEmpty {
		user.Picture = "https://avatars.yandex.net/get-yapic/" + user.Picture + "/islands-200"
	} else {
		user.Picture = ""
	}

	return &Claims{
		Issuer:     "https://login.yandex.ru/info",
		Subject:    user.Id,
		GivenName:  user.FirstName,
		FamilyName: user.LastName,
		Picture:    user.Picture,
		Email:      user.Email,
		Gender:     user.Gender,
		Birthdate:  user.BirthDay,
	}, nil
}
