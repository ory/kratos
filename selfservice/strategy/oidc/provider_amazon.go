// Copyright © 2025 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package oidc

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"slices"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/pkg/errors"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/amazon"

	"github.com/ory/herodot"
	"github.com/ory/x/httpx"
	"github.com/ory/x/otelx"
)

var _ OAuth2Provider = (*ProviderAmazon)(nil)

var amazonSupportedScopes = []string{"profile", "profile:user_id", "postal_code"}

type ProviderAmazon struct {
	*ProviderGenericOIDC
	amazonProfileURL string // Only overriden in tests.
}

type amazonProfileResponse struct {
	UserId     string `json:"user_id"`
	Email      string `json:"email"`
	Name       string `json:"name"`
	PostalCode string `json:"postal_code"`
}

func NewProviderAmazon(
	config *Configuration,
	reg Dependencies,
) Provider {
	config.IssuerURL = amazon.Endpoint.AuthURL
	const amazonProfileURL string = "https://api.amazon.com/user/profile"

	return &ProviderAmazon{
		ProviderGenericOIDC: &ProviderGenericOIDC{
			config: config,
			reg:    reg,
		},
		amazonProfileURL: amazonProfileURL,
	}
}

// Only to be used in tests.
func (p *ProviderAmazon) SetProfileURL(url string) {
	p.amazonProfileURL = url
}

func (p *ProviderAmazon) Config() *Configuration {
	return p.config
}

func (p *ProviderAmazon) oauth2(ctx context.Context) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     p.config.ClientID,
		ClientSecret: p.config.ClientSecret,
		Endpoint:     amazon.Endpoint,
		Scopes:       p.config.Scope,
		RedirectURL:  p.config.Redir(p.reg.Config().OIDCRedirectURIBase(ctx)),
	}
}

func (p *ProviderAmazon) validateConfiguration() error {
	for _, s := range p.config.Scope {
		if !slices.Contains(amazonSupportedScopes, s) {
			return errors.WithStack(
				herodot.ErrMisconfiguration.WithReasonf("scope %s not supported. Supported: %+v", s, amazonSupportedScopes))
		}
	}
	if p.config.PKCE == "auto" {
		return errors.WithStack(herodot.ErrMisconfiguration.WithReason("pkce:auto is not supported because Amazon does not support PKCE discovery"))
	}

	return nil
}

func (p *ProviderAmazon) OAuth2(ctx context.Context) (*oauth2.Config, error) {
	// This is as good a place as any to validate the configuration.
	if err := p.validateConfiguration(); err != nil {
		return nil, err
	}

	return p.oauth2(ctx), nil
}

func (p *ProviderAmazon) AuthCodeURLOptions(r ider) []oauth2.AuthCodeOption {
	return []oauth2.AuthCodeOption{}
}

func (p *ProviderAmazon) Claims(ctx context.Context, exchange *oauth2.Token, query url.Values) (_ *Claims, err error) {
	ctx, span := p.reg.Tracer(ctx).Tracer().Start(ctx, "selfservice.strategy.oidc.ProviderAmazon.Claims")
	defer otelx.End(span, &err)

	_, client := httpx.SetOAuth2(ctx, p.reg.HTTPClient(ctx), p.oauth2(ctx), exchange)

	req, err := retryablehttp.NewRequestWithContext(ctx, http.MethodGet, p.amazonProfileURL, nil)
	if err != nil {
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithReason("failed to create HTTP request").WithDetail("url", p.amazonProfileURL).WithError(err.Error()))
	}
	req.Header.Set("x-amz-access-token", exchange.AccessToken)
	resp, err := client.Do(req)
	if err != nil {
		return nil, errors.WithStack(herodot.ErrUpstreamError.WithReason("failed to make HTTP request").WithDetail("url", p.amazonProfileURL).WithError(err.Error()))
	}
	defer func() { _ = resp.Body.Close() }()
	body := io.LimitReader(resp.Body, 64*1024) // 64 KiB

	if resp.StatusCode != http.StatusOK {
		rawResponse, _ := io.ReadAll(body)
		return nil, errors.WithStack(herodot.ErrUpstreamError.WithReason("non 200 response").WithDetail("url", p.amazonProfileURL).WithDetail("external_error", string(rawResponse)).
			WithDetail("external_status_code", resp.StatusCode))
	}

	profile := amazonProfileResponse{}
	if err := json.NewDecoder(body).Decode(&profile); err != nil {
		rawResponse, _ := io.ReadAll(body)
		return nil, errors.WithStack(herodot.ErrUpstreamError.WithDetail("url", p.amazonProfileURL).WithDetail("raw_response", rawResponse).WithError(err.Error()))
	}

	claims := &Claims{
		Subject:  profile.UserId,
		Issuer:   amazon.Endpoint.TokenURL,
		Name:     profile.Name,
		Email:    profile.Email,
		Zoneinfo: profile.PostalCode,
	}

	return claims, nil
}
