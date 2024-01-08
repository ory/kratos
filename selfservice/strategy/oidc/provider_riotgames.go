// Copyright © 2024 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package oidc

import (
	"context"
	"encoding/json"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/pkg/errors"
	"golang.org/x/oauth2"
	"net/url"

	"github.com/ory/herodot"
	"github.com/ory/x/httpx"
)

type ProviderRiotGames struct {
	*ProviderGenericOIDC
}

var (
	rsoAuthEndpoint = oauth2.Endpoint{
		AuthURL:   "https://auth.riotgames.com/authorize",
		TokenURL:  "https://auth.riotgames.com/token",
		AuthStyle: oauth2.AuthStyleInHeader,
	}
	rsoUserEndpoint = "https://auth.riotgames.com/userinfo"
)

func NewProviderRiotGames(
	config *Configuration,
	reg Dependencies,
) Provider {
	return &ProviderRiotGames{
		&ProviderGenericOIDC{
			config: config,
			reg:    reg,
		},
	}
}

func (rs *ProviderRiotGames) Config() *Configuration {
	return rs.config
}

func (rs *ProviderRiotGames) OAuth2(ctx context.Context) (*oauth2.Config, error) {

	return &oauth2.Config{
		ClientID:     rs.config.ClientID,
		ClientSecret: rs.config.ClientSecret,
		Endpoint:     rsoAuthEndpoint,
		// Riot Games uses fixed scope that can not be configured in runtime
		Scopes:      rs.config.Scope,
		RedirectURL: rs.config.Redir(rs.reg.Config().OIDCRedirectURIBase(ctx)),
	}, nil

}

func (rs *ProviderRiotGames) Claims(ctx context.Context, exchange *oauth2.Token, query url.Values) (*Claims, error) {
	// riotGamesClaim is defined in the https://beta.developer.riotgames.com/sign-on
	type riotGamesClaim struct {
		Sub  string `json:"sub"`
		Cpid string `json:"cpid"`
		Jti  string `json:"jti"`
	}
	var (
		client = rs.reg.HTTPClient(ctx, httpx.ResilientClientDisallowInternalIPs())
		user   riotGamesClaim
	)

	req, err := retryablehttp.NewRequest("GET", rsoUserEndpoint, nil)
	if err != nil {
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("%s", err))
	}

	exchange.SetAuthHeader(req.Request)
	res, err := client.Do(req)
	if err != nil {
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("%s", err))
	}
	defer res.Body.Close()

	if err := logUpstreamError(rs.reg.Logger(), res); err != nil {
		return nil, err
	}

	if err := json.NewDecoder(res.Body).Decode(&user); err != nil {
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("%s", err))
	}

	return &Claims{
		Issuer:  rsoUserEndpoint,
		Subject: user.Sub,
	}, nil
}

func (rs *ProviderRiotGames) AuthCodeURLOptions(r ider) []oauth2.AuthCodeOption {
	return []oauth2.AuthCodeOption{}
}
