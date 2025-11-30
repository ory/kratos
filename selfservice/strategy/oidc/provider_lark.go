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
	"github.com/ory/kratos/selfservice/strategy/oidc/claims"
	"github.com/ory/x/httpx"
)

var _ OAuth2Provider = (*ProviderLark)(nil)

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
	reg Dependencies,
) Provider {
	return &ProviderLark{
		&ProviderGenericOIDC{
			config: config,
			reg:    reg,
		},
	}
}

func (pl *ProviderLark) Config() *Configuration {
	return pl.config
}

func (pl *ProviderLark) OAuth2(ctx context.Context) (*oauth2.Config, error) {
	return &oauth2.Config{
		ClientID:     pl.config.ClientID,
		ClientSecret: pl.config.ClientSecret,
		Endpoint:     larkAuthEndpoint,
		// Lark uses fixed scope that can not be configured in runtime
		Scopes:      pl.config.Scope,
		RedirectURL: pl.config.Redir(pl.reg.Config().OIDCRedirectURIBase(ctx)),
	}, nil
}

func (pl *ProviderLark) Claims(ctx context.Context, exchange *oauth2.Token, query url.Values) (*claims.Claims, error) {
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
		client = pl.reg.HTTPClient(ctx, httpx.ResilientClientDisallowInternalIPs())
		user   larkClaim
	)

	req, err := retryablehttp.NewRequest("GET", larkUserEndpoint, nil)
	if err != nil {
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithWrap(err).WithReasonf("%s", err))
	}

	exchange.SetAuthHeader(req.Request)
	res, err := client.Do(req)
	if err != nil {
		return nil, errors.WithStack(herodot.ErrUpstreamError.WithWrap(err).WithReasonf("%s", err))
	}
	defer func() { _ = res.Body.Close() }()

	if err := logUpstreamError(pl.reg.Logger(), res); err != nil {
		return nil, err
	}

	if err := json.NewDecoder(res.Body).Decode(&user); err != nil {
		return nil, errors.WithStack(herodot.ErrUpstreamError.WithWrap(err).WithReasonf("%s", err))
	}

	return &claims.Claims{
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
