// Copyright © 2025 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package oidc

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/pkg/errors"
	"golang.org/x/oauth2"

	"github.com/ory/herodot"
	"github.com/ory/x/httpx"
	"github.com/ory/x/otelx"
)

var _ OAuth2Provider = (*ProviderUAEPass)(nil)

// UAE PASS scopes for different user types
// See: https://docs.uaepass.ae/feature-guides/authentication/web-application/endpoints
//
// For UAE Citizens/Residents:
//   - urn:uae:digitalid:profile:general - General profile information
//   - urn:uae:digitalid:profile:general:profileType - Profile type (SOP1/SOP2/SOP3)
//   - urn:uae:digitalid:profile:general:unifiedId - Unified ID
//
// For Visitors:
//   - urn:uae:digitalid:profile - Basic profile
//
// Common scopes:
//   - openid - Required for OIDC flows (returns id_token)

// ProviderUAEPass implements the OAuth2Provider interface for UAE PASS.
// UAE PASS is the UAE's official digital identity platform that allows
// citizens, residents, and visitors to authenticate with government and
// private sector services.
type ProviderUAEPass struct {
	*ProviderGenericOIDC
	baseURL     string // Base URL for UAE PASS (e.g., https://id.uaepass.ae or https://stg-id.uaepass.ae)
	userinfoURL string // Can be overridden in tests
}

// NewProviderUAEPass creates a new UAE PASS provider.
// The IssuerURL configuration determines which UAE PASS environment to use:
// - Production: https://id.uaepass.ae (default if not set)
// - Staging: https://stg-id.uaepass.ae
func NewProviderUAEPass(
	config *Configuration,
	reg Dependencies,
) Provider {
	// Use IssuerURL if provided, otherwise default to production
	baseURL := config.IssuerURL
	if baseURL == "" {
		baseURL = "https://id.uaepass.ae"
	}

	// Normalize: remove trailing slash if present
	baseURL = strings.TrimSuffix(baseURL, "/")

	// UAE PASS doesn't support PKCE discovery, so treat "auto" (or unset) as "force"
	// since UAE PASS does support PKCE.
	if config.PKCE == "" || config.PKCE == "auto" {
		config.PKCE = "force"
	}

	return &ProviderUAEPass{
		ProviderGenericOIDC: &ProviderGenericOIDC{
			config: config,
			reg:    reg,
		},
		baseURL:     baseURL,
		userinfoURL: baseURL + "/idshub/userinfo",
	}
}

// oauth2 returns the OAuth2 configuration for UAE PASS.
func (p *ProviderUAEPass) oauth2(ctx context.Context) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     p.config.ClientID,
		ClientSecret: p.config.ClientSecret,
		Endpoint: oauth2.Endpoint{
			AuthURL:   p.baseURL + "/idshub/authorize",
			TokenURL:  p.baseURL + "/idshub/token",
			AuthStyle: oauth2.AuthStyleInHeader,
		},
		Scopes:      p.config.Scope,
		RedirectURL: p.config.Redir(p.reg.Config().OIDCRedirectURIBase(ctx)),
	}
}

// OAuth2 returns the OAuth2 configuration for UAE PASS.
func (p *ProviderUAEPass) OAuth2(ctx context.Context) (*oauth2.Config, error) {
	return p.oauth2(ctx), nil
}

// UAE PASS acr_values for different authentication levels
// UAEPassACRWeb is the default authentication level for web-based authentication
const UAEPassACRWeb = "urn:safelayer:tws:policies:authentication:level:low" //nolint:gosec // not a credential, this is an ACR URN

// AuthCodeURLOptions returns OAuth2 authorization URL options.
// UAE PASS requires acr_values parameter for authentication.
// The default is web-based authentication. To use mobile app-to-app
// authentication, pass acr_values via upstream_parameters.
func (p *ProviderUAEPass) AuthCodeURLOptions(r ider) []oauth2.AuthCodeOption {
	return []oauth2.AuthCodeOption{
		oauth2.SetAuthURLParam("acr_values", UAEPassACRWeb),
	}
}

// Claims fetches user claims from the UAE PASS userinfo endpoint.
// UAE PASS requires the access token to be passed as a Bearer token in the Authorization header.
func (p *ProviderUAEPass) Claims(ctx context.Context, exchange *oauth2.Token, query url.Values) (_ *Claims, err error) {
	ctx, span := p.reg.Tracer(ctx).Tracer().Start(ctx, "selfservice.strategy.oidc.ProviderUAEPass.Claims")
	defer otelx.End(span, &err)

	ctx, client := httpx.SetOAuth2(ctx, p.reg.HTTPClient(ctx), p.oauth2(ctx), exchange)

	req, err := retryablehttp.NewRequestWithContext(ctx, http.MethodGet, p.userinfoURL, nil)
	if err != nil {
		return nil, errors.WithStack(herodot.ErrInternalServerError.
			WithReason("failed to create HTTP request").
			WithDetail("url", p.userinfoURL).
			WithError(err.Error()))
	}

	// UAE PASS expects the access token as a Bearer token in the Authorization header
	req.Header.Set("Authorization", "Bearer "+exchange.AccessToken)

	resp, err := client.Do(req)
	if err != nil {
		return nil, errors.WithStack(herodot.ErrUpstreamError.
			WithReason("failed to make HTTP request to UAE PASS userinfo endpoint").
			WithDetail("url", p.userinfoURL).
			WithError(err.Error()))
	}
	defer func() { _ = resp.Body.Close() }()

	body := io.LimitReader(resp.Body, 64*1024) // 64 KiB limit

	if resp.StatusCode != http.StatusOK {
		rawResponse, _ := io.ReadAll(body)
		return nil, errors.WithStack(herodot.ErrUpstreamError.
			WithReason("UAE PASS userinfo endpoint returned non-200 status").
			WithDetail("url", p.userinfoURL).
			WithDetail("external_error", string(rawResponse)).
			WithDetail("external_status_code", resp.StatusCode))
	}

	var rawClaims map[string]interface{}
	if err := json.NewDecoder(body).Decode(&rawClaims); err != nil {
		return nil, errors.WithStack(herodot.ErrUpstreamError.
			WithReason("failed to decode UAE PASS userinfo response").
			WithDetail("url", p.userinfoURL).
			WithError(err.Error()))
	}

	str := func(key string) string {
		if v, ok := rawClaims[key].(string); ok {
			return v
		}
		return ""
	}

	// Map UAE PASS profile to standard OIDC claims
	claims := &Claims{
		Subject:   str("sub"),
		Issuer:    p.baseURL,
		Email:     str("email"),
		Name:      str("fullnameEN"),
		GivenName: str("firstnameEN"),
		LastName:  str("lastnameEN"),
		RawClaims: rawClaims,
	}

	return claims, nil
}
