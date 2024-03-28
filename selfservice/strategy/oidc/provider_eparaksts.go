// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package oidc

import (
	"context"
	"encoding/json"
	"net/url"
	"path"
	"strings"

	"github.com/hashicorp/go-retryablehttp"

	"github.com/ory/x/httpx"

	"github.com/pkg/errors"
	"golang.org/x/oauth2"

	"github.com/ory/herodot"
)

type ProviderEParaksts struct {
	config   *Configuration
	reg      Dependencies
	acrValue string
}

func NewProviderEParaksts(
	config *Configuration,
	reg Dependencies,
) Provider {
	return &ProviderEParaksts{
		config:   config,
		reg:      reg,
		acrValue: "urn:eparaksts:authentication:flow:sc_plugin",
	}
}

func NewProviderEParakstsMobile(
	config *Configuration,
	reg Dependencies,
) Provider {
	return &ProviderEParaksts{
		config:   config,
		reg:      reg,
		acrValue: "urn:eparaksts:authentication:flow:mobileid",
	}
}

func (g *ProviderEParaksts) Config() *Configuration {
	return g.config
}

func (g *ProviderEParaksts) oauth2(ctx context.Context) (*oauth2.Config, error) {
	endpoint, err := url.Parse(g.config.IssuerURL)
	if err != nil {
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("%s", err))
	}

	authUrl := *endpoint
	authUrl.Path = path.Join(authUrl.Path, "/trustedx-authserver/oauth/lvrtc-eipsign-as")

	values := url.Values{}
	values.Add("prompt", "login")
	values.Add("acr_values", g.acrValue)
	values.Add("ui_locales", "lv")
	authUrl.RawQuery = values.Encode()

	tokenUrl := *endpoint
	tokenUrl.Path = path.Join(tokenUrl.Path, "/trustedx-authserver/oauth/lvrtc-eipsign-as/token")

	c := &oauth2.Config{
		ClientID:     g.config.ClientID,
		ClientSecret: g.config.ClientSecret,
		Endpoint: oauth2.Endpoint{
			AuthURL:  authUrl.String(),
			TokenURL: tokenUrl.String(),
		},
		Scopes:      g.config.Scope,
		RedirectURL: g.config.Redir(g.reg.Config().OIDCRedirectURIBase(ctx)),
	}
	return c, nil
}

func (g *ProviderEParaksts) AuthCodeURLOptions(r ider) []oauth2.AuthCodeOption {
	return []oauth2.AuthCodeOption{}
}

func (g *ProviderEParaksts) OAuth2(ctx context.Context) (*oauth2.Config, error) {
	return g.oauth2(ctx)
}

func (g *ProviderEParaksts) Claims(ctx context.Context, exchange *oauth2.Token, query url.Values) (*Claims, error) {
	o, err := g.OAuth2(ctx)
	if err != nil {
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("%s", err))
	}

	u, err := url.Parse(g.config.IssuerURL)
	if err != nil {
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("%s", err))
	}
	u.Path = path.Join(u.Path, "/trustedx-resources/openid/v1/users/me")

	ctx, client := httpx.SetOAuth2(ctx, g.reg.HTTPClient(ctx), o, exchange)
	req, err := retryablehttp.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if err != nil {
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("%s", err))
	}

	req.Header.Add("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("%s", err))
	}
	defer resp.Body.Close()

	if err := logUpstreamError(g.reg.Logger(), resp); err != nil {
		return nil, err
	}

	type User struct {
		Subject      string `json:"sub,omitempty"`
		GivenName    string `json:"given_name,omitempty"`
		FamilyName   string `json:"family_name,omitempty"`
		SerialNumber string `json:"serial_number,omitempty"`
	}
	var user User

	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("%s", err))
	}

	serialNumberDigits := g.ParseSerialNumber(user.SerialNumber)

	return &Claims{
		Issuer:  g.config.IssuerURL,
		Subject: user.Subject,
		RawClaims: map[string]interface{}{
			"serial_number": serialNumberDigits,
		},
		GivenName:  user.GivenName,
		FamilyName: user.FamilyName,
	}, nil
}

func (g *ProviderEParaksts) ParseSerialNumber(serialNumber string) string {
	serialNumber = serialNumber[6:]
	return strings.Replace(serialNumber, "-", "", -1)
}
