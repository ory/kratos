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

	"github.com/ory/x/httpx"
	"github.com/ory/x/stringsx"

	"github.com/tidwall/sjson"

	"github.com/hashicorp/go-retryablehttp"

	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
	"golang.org/x/oauth2"

	"github.com/ory/herodot"
)

type ProviderSalesforce struct {
	*ProviderGenericOIDC
}

func NewProviderSalesforce(
	config *Configuration,
	reg Dependencies,
) Provider {
	return &ProviderSalesforce{
		ProviderGenericOIDC: &ProviderGenericOIDC{
			config: config,
			reg:    reg,
		},
	}
}

func (g *ProviderSalesforce) oauth2(ctx context.Context) (*oauth2.Config, error) {
	endpoint, err := url.Parse(g.config.IssuerURL)
	if err != nil {
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("%s", err))
	}

	authUrl := *endpoint
	tokenUrl := *endpoint

	authUrl.Path = path.Join(authUrl.Path, "/services/oauth2/authorize")
	tokenUrl.Path = path.Join(tokenUrl.Path, "/services/oauth2/token")

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

func (g *ProviderSalesforce) OAuth2(ctx context.Context) (*oauth2.Config, error) {
	return g.oauth2(ctx)
}

func (g *ProviderSalesforce) Claims(ctx context.Context, exchange *oauth2.Token, query url.Values) (*Claims, error) {
	o, err := g.OAuth2(ctx)
	if err != nil {
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("%s", err))
	}

	u, err := url.Parse(g.config.IssuerURL)
	if err != nil {
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("%s", err))
	}
	u.Path = path.Join(u.Path, "/services/oauth2/userinfo")

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

	// Once Salesforce fixes this bug, all this workaround can be removed.
	b, err := io.ReadAll(io.LimitReader(resp.Body, 1024*1024))
	if err != nil {
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("%s", err))
	}

	b, err = salesforceUpdatedAtWorkaround(b)
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

// There is a bug in the response from Salesforce. The updated_at field may be a string and not an int64.
// https://help.salesforce.com/s/articleView?id=sf.remoteaccess_using_userinfo_endpoint.htm&type=5
// We work around this by reading the json generically (as map[string]inteface{} and looking at the updated_at field
// if it exists. If it's the wrong type (string), we fill out the claims by hand.
func salesforceUpdatedAtWorkaround(body []byte) ([]byte, error) {
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
