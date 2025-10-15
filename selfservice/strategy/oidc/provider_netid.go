// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package oidc

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"slices"
	"strings"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/pkg/errors"
	"golang.org/x/oauth2"

	"github.com/ory/herodot"
	"github.com/ory/x/httpx"
	"github.com/ory/x/urlx"
)

const (
	defaultBrokerScheme = "https"
	defaultBrokerHost   = "broker.netid.de"
)

var _ OAuth2Provider = (*ProviderNetID)(nil)

type ProviderNetID struct {
	*ProviderGenericOIDC
}

func NewProviderNetID(
	config *Configuration,
	reg Dependencies,
) Provider {
	config.IssuerURL = fmt.Sprintf("%s://%s/", defaultBrokerScheme, defaultBrokerHost)
	if !slices.Contains(config.Scope, oidc.ScopeOpenID) {
		config.Scope = append(config.Scope, oidc.ScopeOpenID)
	}

	return &ProviderNetID{
		ProviderGenericOIDC: &ProviderGenericOIDC{
			config: config,
			reg:    reg,
		},
	}
}

func (n *ProviderNetID) OAuth2(ctx context.Context) (*oauth2.Config, error) {
	return n.oAuth2(ctx)
}

func (n *ProviderNetID) oAuth2(ctx context.Context) (*oauth2.Config, error) {
	u := n.brokerURL()

	authURL := urlx.AppendPaths(u, "/authorize")
	tokenURL := urlx.AppendPaths(u, "/token")

	return &oauth2.Config{
		ClientID:     n.config.ClientID,
		ClientSecret: n.config.ClientSecret,
		Endpoint: oauth2.Endpoint{
			AuthURL:  authURL.String(),
			TokenURL: tokenURL.String(),
		},
		Scopes:      n.config.Scope,
		RedirectURL: n.config.Redir(n.reg.Config().OIDCRedirectURIBase(ctx)),
	}, nil
}

func (n *ProviderNetID) Claims(ctx context.Context, exchange *oauth2.Token, _ url.Values) (*Claims, error) {
	o, err := n.OAuth2(ctx)
	if err != nil {
		return nil, err
	}

	ctx, client := httpx.SetOAuth2(ctx, n.reg.HTTPClient(ctx), o, exchange)
	req, err := retryablehttp.NewRequestWithContext(ctx, "GET", urlx.AppendPaths(n.brokerURL(), "/userinfo").String(), nil)
	if err != nil {
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("%s", err))
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, errors.WithStack(herodot.ErrUpstreamError.WithWrap(err).WithReasonf("%s", err))
	}
	defer func() { _ = resp.Body.Close() }()

	if err := logUpstreamError(n.reg.Logger(), resp); err != nil {
		return nil, err
	}

	p, err := n.provider(ctx)
	if err != nil {
		return nil, err
	}

	raw, ok := exchange.Extra("id_token").(string)
	if !ok || len(raw) == 0 {
		return nil, errors.WithStack(ErrIDTokenMissing)
	}

	claims, err := n.verifyAndDecodeClaimsWithProvider(ctx, p, raw)
	if err != nil {
		return nil, err
	}

	var userinfo Claims
	if err := json.NewDecoder(resp.Body).Decode(&userinfo); err != nil {
		return nil, errors.WithStack(herodot.ErrUpstreamError.WithWrap(err).WithReasonf("%s", err))
	}
	userinfo.Issuer = claims.Issuer
	userinfo.Subject = claims.Subject
	return &userinfo, nil
}

func (n *ProviderNetID) Verify(ctx context.Context, rawIDToken string) (*Claims, error) {
	provider, err := n.provider(ctx)
	if err != nil {
		return nil, err
	}

	req, err := retryablehttp.NewRequestWithContext(ctx, "POST", urlx.AppendPaths(n.brokerURL(), "/token").String(), strings.NewReader(url.Values{
		"grant_type":  {"netid_fedcm"},
		"fedcm_token": {rawIDToken},
	}.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Origin", n.config.NetIDTokenOriginHeader)
	res, err := n.reg.HTTPClient(ctx).Do(req)
	if err != nil {
		return nil, err
	}

	token := struct {
		IDToken string `json:"id_token"`
	}{}

	if err := json.NewDecoder(res.Body).Decode(&token); err != nil {
		return nil, err
	}

	idToken, err := provider.VerifierContext(
		n.withHTTPClientContext(ctx),
		&oidc.Config{ClientID: n.config.ClientID},
	).Verify(ctx, token.IDToken)
	if err != nil {
		return nil, err
	}

	var (
		claims    Claims
		rawClaims map[string]any
	)

	if err = idToken.Claims(&claims); err != nil {
		return nil, err
	}
	if err = idToken.Claims(&rawClaims); err != nil {
		return nil, err
	}
	claims.RawClaims = rawClaims

	return &claims, nil
}

func (n *ProviderNetID) brokerURL() *url.URL {
	return &url.URL{Scheme: defaultBrokerScheme, Host: defaultBrokerHost}
}
