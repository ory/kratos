// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package oidc

import (
	"context"
	"encoding/json"
	"net/url"
	"strings"
	"time"

	"github.com/pkg/errors"
	"golang.org/x/oauth2"

	"github.com/ory/x/httpx"

	"github.com/hashicorp/go-retryablehttp"

	"github.com/ory/herodot"
)

type ProviderDingTalk struct {
	config *Configuration
	reg    dependencies
}

func NewProviderDingTalk(
	config *Configuration,
	reg dependencies,
) *ProviderDingTalk {
	return &ProviderDingTalk{
		config: config,
		reg:    reg,
	}
}

func (g *ProviderDingTalk) Config() *Configuration {
	return g.config
}

func (g *ProviderDingTalk) oauth2(ctx context.Context) *oauth2.Config {
	var endpoint = oauth2.Endpoint{
		AuthURL:  "https://login.dingtalk.com/oauth2/auth",
		TokenURL: "https://api.dingtalk.com/v1.0/oauth2/userAccessToken",
	}

	return &oauth2.Config{
		ClientID:     g.config.ClientID,
		ClientSecret: g.config.ClientSecret,
		Endpoint:     endpoint,
		// DingTalk only allow to set scopes: openid or openid corpid
		Scopes:      g.config.Scope,
		RedirectURL: g.config.Redir(g.reg.Config().OIDCRedirectURIBase(ctx)),
	}
}

func (g *ProviderDingTalk) AuthCodeURLOptions(r ider) []oauth2.AuthCodeOption {
	return []oauth2.AuthCodeOption{
		oauth2.SetAuthURLParam("prompt", "consent"),
	}
}

func (g *ProviderDingTalk) OAuth2(ctx context.Context) (*oauth2.Config, error) {
	return g.oauth2(ctx), nil
}

func (g *ProviderDingTalk) Exchange(ctx context.Context, code string, opts ...oauth2.AuthCodeOption) (*oauth2.Token, error) {
	conf, err := g.OAuth2(ctx)
	if err != nil {
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("%s", err))
	}

	pTokenParams := &struct {
		ClientId     string `json:"clientId"`
		ClientSecret string `json:"clientSecret"`
		Code         string `json:"code"`
		GrantType    string `json:"grantType"`
	}{conf.ClientID, conf.ClientSecret, code, "authorization_code"}
	bs, err := json.Marshal(pTokenParams)
	if err != nil {
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("%s", err))
	}

	r := strings.NewReader(string(bs))
	client := g.reg.HTTPClient(ctx, httpx.ResilientClientDisallowInternalIPs())
	req, err := retryablehttp.NewRequest("POST", conf.Endpoint.TokenURL, r)
	if err != nil {
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("%s", err))
	}

	req.Header.Add("Content-Type", "application/json;charset=UTF-8")
	resp, err := client.Do(req)
	if err != nil {
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("%s", err))
	}
	defer resp.Body.Close()

	if err := logUpstreamError(g.reg.Logger(), resp); err != nil {
		return nil, err
	}

	var dToken struct {
		ErrCode     int    `json:"code"`
		ErrMsg      string `json:"message"`
		AccessToken string `json:"accessToken"` // Interface call credentials
		ExpiresIn   int64  `json:"expireIn"`    // access_token interface call credential timeout time, unit (seconds)
	}

	if err := json.NewDecoder(resp.Body).Decode(&dToken); err != nil {
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("%s", err))
	}

	if dToken.ErrCode != 0 {
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("dToken.ErrCode = %d, dToken.ErrMsg = %s", dToken.ErrCode, dToken.ErrMsg))
	}

	token := &oauth2.Token{
		AccessToken: dToken.AccessToken,
		Expiry:      time.Unix(time.Now().Unix()+int64(dToken.ExpiresIn), 0),
	}
	return token, nil
}

func (g *ProviderDingTalk) Claims(ctx context.Context, exchange *oauth2.Token, _ url.Values) (*Claims, error) {
	userInfoURL := "https://api.dingtalk.com/v1.0/contact/users/me"
	accessToken := exchange.AccessToken

	client := g.reg.HTTPClient(ctx, httpx.ResilientClientDisallowInternalIPs())
	req, err := retryablehttp.NewRequest("GET", userInfoURL, nil)
	if err != nil {
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("%s", err))
	}

	req.Header.Add("x-acs-dingtalk-access-token", accessToken)
	resp, err := client.Do(req)
	if err != nil {
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("%s", err))
	}
	defer resp.Body.Close()

	if err := logUpstreamError(g.reg.Logger(), resp); err != nil {
		return nil, err
	}

	var user struct {
		Nick      string `json:"nick"`
		OpenId    string `json:"openId"`
		AvatarUrl string `json:"avatarUrl"`
		Email     string `json:"email"`
		ErrMsg    string `json:"message"`
		ErrCode   string `json:"code"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("%s", err))
	}

	if user.ErrMsg != "" {
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("userResp.ErrCode = %s, userResp.ErrMsg = %s", user.ErrCode, user.ErrMsg))
	}

	return &Claims{
		Issuer:   userInfoURL,
		Subject:  user.OpenId,
		Nickname: user.Nick,
		Name:     user.Nick,
		Picture:  user.AvatarUrl,
		Email:    user.Email,
	}, nil
}
