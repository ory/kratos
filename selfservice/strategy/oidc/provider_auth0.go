// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package oidc

import (
	"context"
	"encoding/json"
	"io"
	"net/url"
	"path"
	"time"

	"github.com/ory/x/stringsx"

	"github.com/tidwall/sjson"

	"github.com/hashicorp/go-retryablehttp"

	"github.com/ory/x/httpx"

	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
	"golang.org/x/oauth2"

	"github.com/ory/herodot"
)

type ProviderAuth0 struct {
	*ProviderGenericOIDC
}

func NewProviderAuth0(
	config *Configuration,
	reg dependencies,
) *ProviderAuth0 {
	return &ProviderAuth0{
		ProviderGenericOIDC: &ProviderGenericOIDC{
			config: config,
			reg:    reg,
		},
	}
}

func (g *ProviderAuth0) oauth2(ctx context.Context) (*oauth2.Config, error) {
	endpoint, err := url.Parse(g.config.IssuerURL)
	if err != nil {
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("%s", err))
	}

	authUrl := *endpoint
	tokenUrl := *endpoint

	authUrl.Path = path.Join(authUrl.Path, "/authorize")
	tokenUrl.Path = path.Join(tokenUrl.Path, "/oauth/token")

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

func (g *ProviderAuth0) OAuth2(ctx context.Context) (*oauth2.Config, error) {
	return g.oauth2(ctx)
}

func (g *ProviderAuth0) Claims(ctx context.Context, exchange *oauth2.Token, query url.Values) (*Claims, error) {
	o, err := g.OAuth2(ctx)
	if err != nil {
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("%s", err))
	}

	client := g.reg.HTTPClient(ctx, httpx.ResilientClientWithClient(o.Client(ctx, exchange)))
	u, err := url.Parse(g.config.IssuerURL)
	if err != nil {
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("%s", err))
	}
	u.Path = path.Join(u.Path, "/userinfo")
	req, err := retryablehttp.NewRequest("GET", u.String(), nil)
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

	// Once auth0 fixes this bug, all this workaround can be removed.
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("%s", err))
	}

	b, err = authZeroUpdatedAtWorkaround(b)
	if err != nil {
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("%s", err))
	}

	// Once we get here, we know that if there is an updated_at field in the json, it is the correct type.
	var claims Claims
	if err := json.Unmarshal(b, &claims); err != nil {
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("%s", err))
	}

	claims.Issuer = stringsx.Coalesce(claims.Issuer, g.config.IssuerURL)
	return &claims, nil
}

// There is a bug in the response from Auth0. The updated_at field may be a string and not an int64.
// https://community.auth0.com/t/oidc-id-token-claim-updated-at-violates-oidc-specification-breaks-rp-implementations/24098
// We work around this by reading the json generically (as map[string]inteface{} and looking at the updated_at field
// if it exists. If it's the wrong type (string), we fill out the claims by hand.
func authZeroUpdatedAtWorkaround(body []byte) ([]byte, error) {
	// Force updatedAt to be an int if given as a string in the response.
	if updatedAtField := gjson.GetBytes(body, "updated_at"); updatedAtField.Exists() && updatedAtField.Type == gjson.String {
		t, err := time.Parse(time.RFC3339, updatedAtField.String())
		if err != nil {
			return nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("bad time format in updated_at"))
		}
		body, err = sjson.SetBytes(body, "updated_at", t.Unix())
		if err != nil {
			return nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("%s", err))
		}
	}
	return body, nil
}
