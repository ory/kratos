// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package oidc

import (
	"context"
	"encoding/json"
	"log"
	"net/url"
	"strconv"

	"github.com/hashicorp/go-retryablehttp"

	"github.com/ory/x/httpx"

	"github.com/pkg/errors"
	"golang.org/x/oauth2"

	"github.com/ory/herodot"
)

type ProviderEParaksts struct {
	config *Configuration
	reg    Dependencies
}

func NewProviderEParaksts(
	config *Configuration,
	reg Dependencies,
) Provider {
	return &ProviderEParaksts{
		config: config,
		reg:    reg,
	}
}

func (g *ProviderEParaksts) Config() *Configuration {
	return g.config
}

func (g *ProviderEParaksts) oauth2(ctx context.Context) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     g.config.ClientID,
		ClientSecret: g.config.ClientSecret,
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://eidas-demo.eparaksts.lv/trustedx-authserver/oauth/lvrtc-eipsign-as?approval_prompt=auto&redirect_uri=http://127.0.0.1:4433/self-service/methods/oidc/callback/eparaksts&acr_values=urn:eparaksts:authentication:flow:sc_plugin&ui_locales=lv&digests_summary_algorithm=SHA256",
			TokenURL: "https://eidas-demo.eparaksts.lv/trustedx-authserver/oauth/lvrtc-eipsign-as/token",
		},
		Scopes:      g.config.Scope,
		RedirectURL: g.config.Redir(g.reg.Config().OIDCRedirectURIBase(ctx)),
	}
}

func (g *ProviderEParaksts) AuthCodeURLOptions(r ider) []oauth2.AuthCodeOption {
	return []oauth2.AuthCodeOption{}
}

func (g *ProviderEParaksts) OAuth2(ctx context.Context) (*oauth2.Config, error) {
	return g.oauth2(ctx), nil
}

func (g *ProviderEParaksts) Claims(ctx context.Context, exchange *oauth2.Token, query url.Values) (*Claims, error) {
	o, err := g.OAuth2(ctx)
	if err != nil {
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("%s", err))
	}

	ctx, client := httpx.SetOAuth2(ctx, g.reg.HTTPClient(ctx), o, exchange)
	req, err := retryablehttp.NewRequestWithContext(ctx, "GET", "https://eidas-demo.eparaksts.lv/trustedx-resources/openid/v1/users/me", nil)
	if err != nil {
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("%s", err))
	}
	log.Printf("Token" + exchange.AccessToken)
	resp, err := client.Do(req)
	if err != nil {
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("%s", err))
	}
	defer resp.Body.Close()

	if err := logUpstreamError(g.reg.Logger(), resp); err != nil {
		return nil, err
	}

	type User struct {
		Id           int    `json:"id,omitempty"`
		Name         string `json:"name,omitempty"`
		SerialNumber string `json:"serial_number,omitempty"`
		GivenName    string `json:"given_name,omitempty"`
		FamilyName   string `json:"family_name,omitempty"`
	}
	var user User

	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("%s", err))
	}

	return &Claims{
		Issuer:       "https://eidas-demo.eparaksts.lv/trustedx-resources/openid/v1/user/me",
		Subject:      strconv.Itoa(user.Id),
		Name:         user.Name,
		SerialNumber: user.SerialNumber,
		GivenName:    user.GivenName,
		FamilyName:   user.FamilyName,
	}, nil
}
