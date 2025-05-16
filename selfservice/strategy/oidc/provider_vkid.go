// Copyright © 2025 Ory Corp
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

var _ OAuth2Provider = (*ProviderVKID)(nil)

type ProviderVKID struct {
	config *Configuration
	reg    Dependencies
}

func NewProviderVKID(
	config *Configuration,
	reg Dependencies,
) Provider {
	// This is required for all apps when the authorization code is exchanged for tokens.
	// See: https://id.vk.com/about/business/go/docs/en/vkid/latest/vk-id/connection/api-integration/realization
	config.PKCE = "force"
	// A unique device ID. The client must save this ID and pass it in subsequent requests to the authorization server.
	config.PassCallbackParams = []string{"device_id"}

	return &ProviderVKID{
		config: config,
		reg:    reg,
	}
}

func (p *ProviderVKID) Config() *Configuration {
	return p.config
}

func (p *ProviderVKID) oauth2(ctx context.Context) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     p.config.ClientID,
		ClientSecret: p.config.ClientSecret,
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://id.vk.com/authorize",
			TokenURL: "https://id.vk.com/oauth2/auth",
		},
		Scopes:      p.config.Scope,
		RedirectURL: p.config.Redir(p.reg.Config().OIDCRedirectURIBase(ctx)),
	}
}

func (p *ProviderVKID) AuthCodeURLOptions(r ider) []oauth2.AuthCodeOption {
	var opts []oauth2.AuthCodeOption
	if p.config.VKIDProviderParam != "" {
		opts = append(opts, oauth2.SetAuthURLParam("provider", p.config.VKIDProviderParam))
	}
	return opts
}

func (p *ProviderVKID) OAuth2(ctx context.Context) (*oauth2.Config, error) {
	return p.oauth2(ctx), nil
}

func (p *ProviderVKID) Claims(ctx context.Context, exchange *oauth2.Token, query url.Values) (*Claims, error) {
	o, err := p.OAuth2(ctx)
	if err != nil {
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("%s", err))
	}

	ctx, client := httpx.SetOAuth2(ctx, p.reg.HTTPClient(ctx), o, exchange)
	req, err := retryablehttp.NewRequestWithContext(ctx, "POST", "https://id.vk.com/oauth2/user_info?client_id="+p.config.ClientID+"&access_token="+exchange.AccessToken, nil)
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
		UserId    string `json:"user_id,omitempty"`
		FirstName string `json:"first_name,omitempty"`
		LastName  string `json:"last_name,omitempty"`
		Phone     string `json:"phone,omitempty"`
		Avatar    string `json:"avatar,omitempty"`
		Email     string `json:"email,omitempty"`
		Gender    int    `json:"sex,omitempty"`
		BirthDay  string `json:"birthday,omitempty"`
	}

	var response struct {
		User User `json:"user,omitempty"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("%s", err))
	}

	if response.User.UserId == "" {
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("VK ID did not return a user id in the user_info request."))
	}

	gender := ""
	switch response.User.Gender {
	case 1:
		gender = "female"
	case 2:
		gender = "male"
	}

	return &Claims{
		Issuer:              "https://id.vk.com/oauth2/user_info",
		Subject:             response.User.UserId,
		GivenName:           response.User.FirstName,
		FamilyName:          response.User.LastName,
		Picture:             response.User.Avatar,
		Email:               response.User.Email,
		EmailVerified:       response.User.Email != "", // VK ID returns only verified email
		PhoneNumber:         response.User.Phone,
		PhoneNumberVerified: response.User.Phone != "", // VK ID returns only verified phone number
		Gender:              gender,
		Birthdate:           response.User.BirthDay,
	}, nil
}
