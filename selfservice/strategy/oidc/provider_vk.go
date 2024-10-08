// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package oidc

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/go-retryablehttp"

	"github.com/ory/x/httpx"

	"github.com/pkg/errors"
	"golang.org/x/oauth2"

	"github.com/ory/herodot"
)

var _ OAuth2Provider = (*ProviderVK)(nil)

type ProviderVK struct {
	config *Configuration
	reg    Dependencies
}

func NewProviderVK(
	config *Configuration,
	reg Dependencies,
) Provider {
	return &ProviderVK{
		config: config,
		reg:    reg,
	}
}

func (p *ProviderVK) Config() *Configuration {
	return p.config
}

func (p *ProviderVK) oauth2(ctx context.Context) *oauth2.Config {
	var endpoint oauth2.Endpoint
	if p.config.PKCE == "force" {
		endpoint = oauth2.Endpoint{
			AuthURL:  "https://id.vk.com/authorize",
			TokenURL: "https://id.vk.com/oauth2/auth",
		}
	} else {
		endpoint = oauth2.Endpoint{
			AuthURL:  "https://oauth.vk.com/authorize",
			TokenURL: "https://oauth.vk.com/access_token",
		}
	}
	return &oauth2.Config{
		ClientID:     p.config.ClientID,
		ClientSecret: p.config.ClientSecret,
		Endpoint:     endpoint,
		Scopes:       p.config.Scope,
		RedirectURL:  p.config.Redir(p.reg.Config().OIDCRedirectURIBase(ctx)),
	}
}

func (p *ProviderVK) AuthCodeURLOptions(r ider) []oauth2.AuthCodeOption {
	return []oauth2.AuthCodeOption{}
}

func (p *ProviderVK) AccessTokenURLOptions(r *http.Request) []oauth2.AuthCodeOption {
	if p.config.PKCE == "force" {
		if deviceID := r.URL.Query().Get("device_id"); deviceID != "" {
			return []oauth2.AuthCodeOption{oauth2.SetAuthURLParam("device_id", deviceID)}
		}
	}
	return []oauth2.AuthCodeOption{}
}

func (p *ProviderVK) OAuth2(ctx context.Context) (*oauth2.Config, error) {
	return p.oauth2(ctx), nil
}

func (p *ProviderVK) Claims(ctx context.Context, exchange *oauth2.Token, query url.Values) (*Claims, error) {
	o, err := p.OAuth2(ctx)
	if err != nil {
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("%s", err))
	}

	ctx, client := httpx.SetOAuth2(ctx, p.reg.HTTPClient(ctx), o, exchange)
	req, err := retryablehttp.NewRequestWithContext(ctx, "GET", "https://api.vk.com/method/users.get?fields=photo_200,nickname,bdate,sex&access_token="+exchange.AccessToken+"&v=5.103", nil)
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

	type User struct {
		Id        int    `json:"id,omitempty"`
		FirstName string `json:"first_name,omitempty"`
		LastName  string `json:"last_name,omitempty"`
		Nickname  string `json:"nickname,omitempty"`
		Picture   string `json:"photo_200,omitempty"`
		Email     string `json:"-"`
		Phone     string `json:"-"`
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

	if p.config.PKCE == "force" {
		reqData := strings.NewReader("client_id=" + p.config.ClientID + "&access_token=" + exchange.AccessToken)
		req, err = retryablehttp.NewRequestWithContext(ctx, "POST", "https://id.vk.com/oauth2/user_info", reqData)
		if err != nil {
			return nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("%s", err))
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		resp, err = client.Do(req)
		if err != nil {
			return nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("%s", err))
		}
		defer resp.Body.Close()

		if err := logUpstreamError(p.reg.Logger(), resp); err != nil {
			return nil, err
		}

		var userInfo struct {
			User struct {
				Email string `json:"email,omitempty"`
				Phone string `json:"phone,omitempty"`
			} `json:"user,omitempty"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
			return nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("%s", err))
		}

		if len(userInfo.User.Phone) > 0 && userInfo.User.Phone[0] != '+' {
			userInfo.User.Phone = "+" + userInfo.User.Phone
		}

		user.Email = userInfo.User.Email
		user.Phone = userInfo.User.Phone
	} else {
		if email, ok := exchange.Extra("email").(string); ok {
			user.Email = email
		}
	}

	gender := ""
	switch user.Gender {
	case 1:
		gender = "female"
	case 2:
		gender = "male"
	}

	t, err := time.Parse("2.1.2006", user.BirthDay)
	if err != nil {
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("%s", err))
	}
	birthDay := t.Format("2006-01-02")

	return &Claims{
		Issuer:      "https://api.vk.com/method/users.get",
		Subject:     strconv.Itoa(user.Id),
		GivenName:   user.FirstName,
		FamilyName:  user.LastName,
		Nickname:    user.Nickname,
		Picture:     user.Picture,
		Email:       user.Email,
		PhoneNumber: user.Phone,
		Gender:      gender,
		Birthdate:   birthDay,
	}, nil
}
