// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package oidc

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/hashicorp/go-retryablehttp"

	"github.com/ory/x/httpx"

	"github.com/ory/kratos/x"

	"github.com/pkg/errors"
	"golang.org/x/oauth2"

	"github.com/ory/herodot"
)

var _ OAuth2Provider = (*ProviderFacebook)(nil)

type ProviderFacebook struct {
	*ProviderGenericOIDC
}

func NewProviderFacebook(
	config *Configuration,
	reg Dependencies,
) Provider {
	config.IssuerURL = "https://www.facebook.com"
	return &ProviderFacebook{
		ProviderGenericOIDC: &ProviderGenericOIDC{
			config: config,
			reg:    reg,
		},
	}
}

func (g *ProviderFacebook) generateAppSecretProof(token *oauth2.Token) string {
	secret := g.config.ClientSecret
	data := token.AccessToken

	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}

func (g *ProviderFacebook) OAuth2(ctx context.Context) (*oauth2.Config, error) {
	p, err := g.provider(ctx)
	if err != nil {
		return nil, err
	}

	endpoint := p.Endpoint()
	endpoint.AuthURL = "https://facebook.com/dialog/oauth"
	endpoint.TokenURL = "https://graph.facebook.com/oauth/access_token"
	return g.oauth2ConfigFromEndpoint(ctx, endpoint), nil
}

func (g *ProviderFacebook) Claims(ctx context.Context, token *oauth2.Token, query url.Values) (*Claims, error) {
	o, err := g.OAuth2(ctx)
	if err != nil {
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("%s", err))
	}

	appSecretProof := g.generateAppSecretProof(token)
	// Do not use the versioned Graph API here. If you do, it will break once the version is deprecated. See also:
	//
	// When you use https://graph.facebook.com/me without specifying a version, Facebook defaults to the oldest
	// available version your app supports. This behavior ensures backward compatibility but can lead to unintended
	// issues if that version becomes deprecated.
	u, err := url.Parse(fmt.Sprintf("https://graph.facebook.com/me?fields=id,name,first_name,last_name,middle_name,email,picture,birthday,gender&appsecret_proof=%s", appSecretProof))
	if err != nil {
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("%s", err))
	}

	ctx, client := httpx.SetOAuth2(ctx, g.reg.HTTPClient(ctx), o, token)
	req, err := retryablehttp.NewRequestWithContext(ctx, "GET", u.String(), nil)
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

	var user struct {
		Id            string `json:"id,omitempty"`
		Name          string `json:"name,omitempty"`
		FirstName     string `json:"first_name,omitempty"`
		LastName      string `json:"last_name,omitempty"`
		MiddleName    string `json:"middle_name,omitempty"`
		Email         string `json:"email,omitempty"`
		EmailVerified bool
		Picture       struct {
			Data struct {
				Url string `json:"url,omitempty"`
			} `json:"data,omitempty"`
		} `json:"picture,omitempty"`
		BirthDay string `json:"birthday,omitempty"`
		Gender   string `json:"gender,omitempty"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("%s", err))
	}

	if user.Email != "" {
		user.EmailVerified = true
	}

	return &Claims{
		Issuer:            u.String(),
		Subject:           user.Id,
		Name:              user.Name,
		GivenName:         user.FirstName,
		FamilyName:        user.LastName,
		MiddleName:        user.MiddleName,
		Nickname:          user.Name,
		PreferredUsername: user.Name,
		Picture:           user.Picture.Data.Url,
		Email:             user.Email,
		EmailVerified:     x.ConvertibleBoolean(user.EmailVerified),
		Gender:            user.Gender,
		Birthdate:         user.BirthDay,
	}, nil
}
