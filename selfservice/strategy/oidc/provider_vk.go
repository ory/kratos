// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package oidc

import (
	"context"
	"encoding/json"
	"net/url"
	"strconv"

	"github.com/hashicorp/go-retryablehttp"

	"github.com/ory/x/httpx"

	"github.com/pkg/errors"
	"golang.org/x/oauth2"

	"github.com/ory/herodot"
)

type ProviderVK struct {
	config *Configuration
	reg    dependencies
}

func NewProviderVK(
	config *Configuration,
	reg dependencies,
) *ProviderVK {
	return &ProviderVK{
		config: config,
		reg:    reg,
	}
}

func (g *ProviderVK) Config() *Configuration {
	return g.config
}

func (g *ProviderVK) oauth2(ctx context.Context) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     g.config.ClientID,
		ClientSecret: g.config.ClientSecret,
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://oauth.vk.com/authorize",
			TokenURL: "https://oauth.vk.com/access_token",
		},
		Scopes:      g.config.Scope,
		RedirectURL: g.config.Redir(g.reg.Config().OIDCRedirectURIBase(ctx)),
	}
}

func (g *ProviderVK) AuthCodeURLOptions(r ider) []oauth2.AuthCodeOption {
	return []oauth2.AuthCodeOption{}
}

func (g *ProviderVK) OAuth2(ctx context.Context) (*oauth2.Config, error) {
	return g.oauth2(ctx), nil
}

func (g *ProviderVK) Claims(ctx context.Context, exchange *oauth2.Token, query url.Values) (*Claims, error) {
	o, err := g.OAuth2(ctx)
	if err != nil {
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("%s", err))
	}

	client := g.reg.HTTPClient(ctx, httpx.ResilientClientWithClient(o.Client(ctx, exchange)))
	req, err := retryablehttp.NewRequest("GET", "https://api.vk.com/method/users.get?fields=photo_200,nickname,bdate,sex&access_token="+exchange.AccessToken+"&v=5.103", nil)
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

	type User struct {
		Id        int    `json:"id,omitempty"`
		FirstName string `json:"first_name,omitempty"`
		LastName  string `json:"last_name,omitempty"`
		Nickname  string `json:"nickname,omitempty"`
		Picture   string `json:"photo_200,omitempty"`
		Email     string `json:"-"`
		Gender    int    `json:"sex,omitempty"`
		BirthDay  string `json:"bdate,omitempty"`
	}

	var response struct {
		Result []User `json:"response,omitempty"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("%s", err))
	}

	if len(response.Result) == 0 {
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("VK did not return a user in the userinfo request."))
	}

	user := response.Result[0]

	if email, ok := exchange.Extra("email").(string); ok {
		user.Email = email
	}

	gender := ""
	switch user.Gender {
	case 1:
		gender = "female"
	case 2:
		gender = "male"
	}

	return &Claims{
		Issuer:     "https://api.vk.com/method/users.get",
		Subject:    strconv.Itoa(user.Id),
		GivenName:  user.FirstName,
		FamilyName: user.LastName,
		Nickname:   user.Nickname,
		Picture:    user.Picture,
		Email:      user.Email,
		Gender:     gender,
		Birthdate:  user.BirthDay,
	}, nil
}
