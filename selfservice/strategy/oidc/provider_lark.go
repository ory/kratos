// Copyright © 2022 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package oidc

import (
	"context"
	"encoding/json"
	"net/url"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/pkg/errors"
	"golang.org/x/oauth2"

	"github.com/ory/herodot"
	"github.com/ory/x/httpx"
)

type ProviderLark struct {
	*ProviderGenericOIDC
}

var (
	larkAuthEndpoint = oauth2.Endpoint{
		AuthURL:   "https://passport.feishu.cn/suite/passport/oauth/authorize",
		TokenURL:  "https://passport.feishu.cn/suite/passport/oauth/token",
		AuthStyle: oauth2.AuthStyleInParams,
	}
	larkUserEndpoint = "https://passport.feishu.cn/suite/passport/oauth/userinfo"
)

func NewProviderLark(
	config *Configuration,
	reg dependencies,
) *ProviderLark {
	return &ProviderLark{
		&ProviderGenericOIDC{
			config: config,
			reg:    reg,
		},
	}
}

func (g *ProviderLark) Config() *Configuration {
	return g.config
}

func (g *ProviderLark) OAuth2(ctx context.Context) (*oauth2.Config, error) {

	return &oauth2.Config{
		ClientID:     g.config.ClientID,
		ClientSecret: g.config.ClientSecret,
		Endpoint:     larkAuthEndpoint,
		// Lark uses fixed scope that can not be configured in runtime
		Scopes:      g.config.Scope,
		RedirectURL: g.config.Redir(g.reg.Config().OIDCRedirectURIBase(ctx)),
	}, nil

}

func (g *ProviderLark) Claims(ctx context.Context, exchange *oauth2.Token, query url.Values) (*Claims, error) {
	// larkClaim is defined in the https://open.feishu.cn/document/common-capabilities/sso/api/get-user-info
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
		client = g.reg.HTTPClient(ctx, httpx.ResilientClientDisallowInternalIPs())
		user   larkClaim
	)

	req, err := retryablehttp.NewRequest("GET", larkUserEndpoint, nil)
	if err != nil {
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("%s", err))
	}

	exchange.SetAuthHeader(req.Request)
	res, err := client.Do(req)
	if err != nil {
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("%s", err))
	}
	defer res.Body.Close()

	if err := logUpstreamError(g.reg.Logger(), res); err != nil {
		return nil, err
	}

	if err := json.NewDecoder(res.Body).Decode(&user); err != nil {
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("%s", err))
	}

	return &Claims{
		Issuer:      larkUserEndpoint,
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
