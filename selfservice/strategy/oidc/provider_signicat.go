// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package oidc

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/sirupsen/logrus"

	"github.com/ory/x/httpx"

	"github.com/pkg/errors"
	"golang.org/x/oauth2"

	"github.com/ory/herodot"
)

const (
	signicatBaseUrl = "https://client.sandbox.signicat.com"
)

type ProviderSignicat struct {
	*ProviderGenericOIDC
}

type UserInfoResponse struct {
	IdpId             string `json:"idp_id,omitempty"`
	FamilyName        string `json:"family_name,omitempty"`
	GivenName         string `json:"given_name,omitempty"`
	Birthdate         string `json:"birthdate,omitempty"`
	Nin               string `json:"nin,omitempty"`
	NinType           string `json:"nin_type,omitempty"`
	NinIssuingCountry string `json:"nin_issuing_country,omitempty"`
	Sub               string `json:"sub,omitempty"`
	SubLegacy         string `json:"sub_legacy,omitempty"`
}

func NewProviderSignicat(
	config *Configuration,
	reg Dependencies,
) Provider {
	return &ProviderSignicat{
		ProviderGenericOIDC: &ProviderGenericOIDC{
			config: config,
			reg:    reg,
		},
	}
}

func (g *ProviderSignicat) oauth2(ctx context.Context) (*oauth2.Config, error) {
	endpoint, err := g.endpoint()
	if err != nil {
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("%s", err))
	}

	authUrl := *endpoint
	tokenUrl := *endpoint

	authUrl.Path = path.Join(authUrl.Path, "/connect/authorize")
	tokenUrl.Path = path.Join(tokenUrl.Path, "/connect/token")

	return &oauth2.Config{
		ClientID:     g.config.ClientID,
		ClientSecret: g.config.ClientSecret,
		Endpoint: oauth2.Endpoint{
			AuthURL:  authUrl.String(),
			TokenURL: tokenUrl.String(),
		},
		Scopes:      g.config.Scope,
		RedirectURL: g.config.Redir(g.reg.Config().OIDCRedirectURIBase(ctx)),
	}, nil
}

func (g *ProviderSignicat) OAuth2(ctx context.Context) (*oauth2.Config, error) {
	return g.oauth2(ctx)
}

func (g *ProviderSignicat) Claims(ctx context.Context, exchange *oauth2.Token, query url.Values) (*Claims, error) {
	o, err := g.OAuth2(ctx)
	if err != nil {
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("%s", err))
	}

	client := g.reg.HTTPClient(ctx, httpx.ResilientClientDisallowInternalIPs(), httpx.ResilientClientWithClient(o.Client(ctx, exchange)))

	u, err := g.endpoint()
	if err != nil {
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("%s", err))
	}
	u.Path = path.Join(u.Path, "/connect/userinfo")
	req, err := retryablehttp.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("%s", err))
	}

	req.Header.Set("Accept", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("%s", err))
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("Expected the Signicat userinfo endpoint to return a 200 OK response but got %d instead", resp.StatusCode))
	}

	var claims Claims

	if err := json.NewDecoder(resp.Body).Decode(&claims); err != nil {
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("%s", err)) // FAILS!
	}

	g.ProviderGenericOIDC.reg.Logger().WithField("provider", g.config.Provider).WithFields(logrus.Fields{
		"claims": fmt.Sprintf("%+v\n", claims),
	}).Debug("Received claims from Signicat userinfo endpoint.")

	return &claims, nil
}

func (g *ProviderSignicat) endpoint() (*url.URL, error) {
	var e = signicatBaseUrl
	if len(g.config.IssuerURL) > 0 {
		e = g.config.IssuerURL
	}
	return url.Parse(e)
}
