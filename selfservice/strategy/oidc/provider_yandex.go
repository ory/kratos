// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package oidc

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/pkg/errors"
	"golang.org/x/oauth2"

	"github.com/ory/x/httpx"

	"github.com/ory/herodot"
)

var _ OAuth2Provider = (*ProviderYandex)(nil)

type ProviderYandex struct {
	config *Configuration
	reg    Dependencies
}

func NewProviderYandex(
	config *Configuration,
	reg Dependencies,
) Provider {
	return &ProviderYandex{
		config: config,
		reg:    reg,
	}
}

func (p *ProviderYandex) Config() *Configuration {
	return p.config
}

func (p *ProviderYandex) oauth2(ctx context.Context) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     p.config.ClientID,
		ClientSecret: p.config.ClientSecret,
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://oauth.yandex.com/authorize",
			TokenURL: "https://oauth.yandex.com/token",
		},
		Scopes:      p.config.Scope,
		RedirectURL: p.config.Redir(p.reg.Config().OIDCRedirectURIBase(ctx)),
	}
}

func (p *ProviderYandex) AuthCodeURLOptions(r ider) []oauth2.AuthCodeOption {
	return []oauth2.AuthCodeOption{}
}

func (p *ProviderYandex) AccessTokenURLOptions(r *http.Request) []oauth2.AuthCodeOption {
	return []oauth2.AuthCodeOption{}
}

func (p *ProviderYandex) OAuth2(ctx context.Context) (*oauth2.Config, error) {
	return p.oauth2(ctx), nil
}

func (p *ProviderYandex) Claims(ctx context.Context, exchange *oauth2.Token, query url.Values) (*Claims, error) {
	o, err := p.OAuth2(ctx)
	if err != nil {
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("%s", err))
	}

	ctx, client := httpx.SetOAuth2(ctx, p.reg.HTTPClient(ctx), o, exchange)
	req, err := retryablehttp.NewRequestWithContext(ctx, "GET", "https://login.yandex.ru/info?format=json&oauth_token="+exchange.AccessToken, nil)
	if err != nil {
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("%s", err))
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("%s", err))
	}
	defer resp.Body.Close()

	if err := logUpstreamError(p.reg.Logger(), resp); err != nil {
		return nil, err
	}

	var user struct {
		Id        string `json:"id,omitempty"`
		FirstName string `json:"first_name,omitempty"`
		LastName  string `json:"last_name,omitempty"`
		Email     string `json:"default_email,omitempty"`
		Phone     struct {
			Number string `json:"number,omitempty"`
		} `json:"default_phone,omitempty"`
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
		Issuer:      "https://login.yandex.ru/info",
		Subject:     user.Id,
		GivenName:   user.FirstName,
		FamilyName:  user.LastName,
		Picture:     user.Picture,
		Email:       user.Email,
		PhoneNumber: user.Phone.Number,
		Gender:      user.Gender,
		Birthdate:   user.BirthDay,
	}, nil
}
