// Copyright © 2022 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package oidc

import (
	"bytes"
	"context"
	"encoding/json"
	"net/url"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/pkg/errors"
	"golang.org/x/oauth2"

	"github.com/ory/herodot"
	"github.com/ory/x/httpx"
)

type ProviderLark struct {
	config *Configuration
	reg    dependencies
}

func NewProviderLark(
	config *Configuration,
	reg dependencies,
) *ProviderLark {
	return &ProviderLark{
		config: config,
		reg:    reg,
	}
}

func (g *ProviderLark) Config() *Configuration {
	return g.config
}

func (g *ProviderLark) OAuth2(ctx context.Context) (*oauth2.Config, error) {
	var endpoint = oauth2.Endpoint{
		AuthURL:   "https://passport.feishu.cn/suite/passport/oauth/authorize",
		TokenURL:  "https://passport.feishu.cn/suite/passport/oauth/token",
		AuthStyle: oauth2.AuthStyleInParams,
	}

	return &oauth2.Config{
		ClientID:     g.config.ClientID,
		ClientSecret: g.config.ClientSecret,
		Endpoint:     endpoint,
		// DingTalk only allow to set scopes: openid or openid corpid
		Scopes:      g.config.Scope,
		RedirectURL: g.config.Redir(g.reg.Config().OIDCRedirectURIBase(ctx)),
	}, nil

}

func (g *ProviderLark) Claims(ctx context.Context, exchange *oauth2.Token, query url.Values) (*Claims, error) {
	type larkClaim struct {
		Sub          string `json:"sub"`
		Name         string `json:"name"`
		Picture      string `json:"picture"`
		OpenID       string `json:"open_id"`
		UnionID      string `json:"union_id"`
		EnName       string `json:"en_name"`
		TenantKey    string `json:"tenant_key"`
		AvatarURL    string `json:"avatar_url"`
		AvatarThumb  string `json:"avatar_thumb"`
		AvatarMiddle string `json:"avatar_middle"`
		AvatarBig    string `json:"avatar_big"`
		Email        string `json:"email"`
		UserID       string `json:"user_id"`
		Mobile       string `json:"mobile"`
	}
	var (
		userEndpoint = "https://passport.feishu.cn/suite/passport/oauth/userinfo"
		accessToken  = exchange.AccessToken
		client       = g.reg.HTTPClient(ctx, httpx.ResilientClientDisallowInternalIPs())
		user         larkClaim
	)

	req, err := retryablehttp.NewRequest("GET", userEndpoint, nil)
	if err != nil {
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("%s", err))
	}

	req.Header.Add("Authorization", "Bearer "+accessToken)
	resp, err := client.Do(req)
	if err != nil {
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("%s", err))
	}
	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("%s", err))
	}

	return &Claims{
		Issuer:      userEndpoint,
		Subject:     user.OpenID,
		Name:        user.Name,
		Nickname:    user.Name,
		Picture:     user.AvatarURL,
		Email:       user.Email,
		PhoneNumber: user.Mobile,
	}, nil

}

func (pl *ProviderLark) AuthCodeURLOptions(r ider) []oauth2.AuthCodeOption {
	return []oauth2.AuthCodeOption{}
}

func (g *ProviderLark) Exchange(ctx context.Context, code string, opts ...oauth2.AuthCodeOption) (*oauth2.Token, error) {

	type (
		larkExchangeReq struct {
			ClientId     string `json:"client_id"`
			ClientSecret string `json:"client_secret"`
			Code         string `json:"code"`
			GrantType    string `json:"grant_type"`
		}
		larkTokenResp struct {
			AccessToken      string `json:"access_token"`
			TokenType        string `json:"token_type"`
			ExpiresIn        int64  `json:"expires_in"`
			RefreshToken     string `json:"refresh_token"`
			RefreshExpiresIn int64  `json:"refresh_expires_in"`
		}
	)

	conf, err := g.OAuth2(ctx)
	if err != nil {
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("%s", err))
	}

	pTokenParams := &larkExchangeReq{
		ClientId:     conf.ClientID,
		ClientSecret: conf.ClientSecret,
		Code:         code,
		GrantType:    "authorization_code",
	}

	bs, err := json.Marshal(pTokenParams)
	if err != nil {
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("%s", err))
	}

	client := g.reg.HTTPClient(ctx, httpx.ResilientClientDisallowInternalIPs())
	req, err := retryablehttp.NewRequest("POST", conf.Endpoint.TokenURL, bytes.NewBuffer(bs))
	if err != nil {
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("%s", err))
	}
	req.Header.Add("Content-Type", "application/json;charset=UTF-8")

	resp, err := client.Do(req)
	if err != nil {
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("%s", err))
	}
	defer resp.Body.Close()

	var dToken larkTokenResp
	if err := json.NewDecoder(resp.Body).Decode(&dToken); err != nil {
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("%s", err))
	}
	token := &oauth2.Token{
		AccessToken:  dToken.AccessToken,
		TokenType:    dToken.TokenType,
		RefreshToken: dToken.RefreshToken,
		Expiry:       time.Unix(time.Now().Unix()+int64(dToken.ExpiresIn), 0),
	}

	return token, nil
}
