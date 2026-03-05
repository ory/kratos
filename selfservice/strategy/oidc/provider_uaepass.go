// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package oidc

import (
	"context"
	"encoding/json"
	"net/url"

	"github.com/hashicorp/go-retryablehttp"

	"github.com/ory/x/httpx"

	"github.com/pkg/errors"
	"golang.org/x/oauth2"

	"github.com/ory/herodot"
)

var _ OAuth2Provider = (*ProviderUAEPASS)(nil)

type ProviderUAEPASS struct {
	config *Configuration
	reg    Dependencies
}

func NewProviderUAEPASS(config *Configuration, reg Dependencies) Provider {
	return &ProviderUAEPASS{
		config: config,
		reg:    reg,
	}
}

func (p *ProviderUAEPASS) Config() *Configuration {
	return p.config
}

// oauth2 returns the OAuth2 config with UAE PASS endpoints.
// Uses config.AuthURL/TokenURL if set, otherwise defaults to staging.
func (p *ProviderUAEPASS) oauth2(ctx context.Context) *oauth2.Config {
	authURL := p.config.AuthURL
	if authURL == "" {
		authURL = "https://stg-id.uaepass.ae/idshub/authorize"
	}
	tokenURL := p.config.TokenURL
	if tokenURL == "" {
		tokenURL = "https://stg-id.uaepass.ae/idshub/token"
	}

	return &oauth2.Config{
		ClientID:     p.config.ClientID,
		ClientSecret: p.config.ClientSecret,
		Endpoint: oauth2.Endpoint{
			AuthURL:   authURL,
			TokenURL:  tokenURL,
			AuthStyle: oauth2.AuthStyleInHeader, // client_secret_basic
		},
		// Use scopes from config directly — do NOT add "openid" as UAE PASS does not support it.
		Scopes:      p.config.Scope,
		RedirectURL: p.config.Redir(p.reg.Config().OIDCRedirectURIBase(ctx)),
	}
}

func (p *ProviderUAEPASS) OAuth2(ctx context.Context) (*oauth2.Config, error) {
	return p.oauth2(ctx), nil
}

// AuthCodeURLOptions adds acr_values and ui_locales required/supported by UAE PASS.
func (p *ProviderUAEPASS) AuthCodeURLOptions(r ider) []oauth2.AuthCodeOption {
	return []oauth2.AuthCodeOption{
		oauth2.SetAuthURLParam("acr_values",
			"urn:safelayer:tws:policies:authentication:level:low"),
	}
}

// userinfoURL returns the userinfo endpoint. If config.IssuerURL is set,
// it derives the URL as {issuer_url}/userinfo. Otherwise, it falls back
// to the UAE PASS staging userinfo endpoint.
func (p *ProviderUAEPASS) userinfoURL() string {
	if p.config.IssuerURL != "" {
		return p.config.IssuerURL + "/userinfo"
	}
	return "https://stg-id.uaepass.ae/idshub/userinfo"
}

// Claims fetches user info from the UAE PASS userinfo endpoint and maps the
// response to Kratos Claims. All raw fields are preserved in RawClaims so
// that downstream Jsonnet mappers can access UAE PASS-specific attributes
// like userType, idn, nationalityEN, etc.
func (p *ProviderUAEPASS) Claims(ctx context.Context, exchange *oauth2.Token, query url.Values) (*Claims, error) {
	o, err := p.OAuth2(ctx)
	if err != nil {
		return nil, err
	}

	ctx, client := httpx.SetOAuth2(ctx, p.reg.HTTPClient(ctx), o, exchange)
	req, err := retryablehttp.NewRequestWithContext(ctx, "GET", p.userinfoURL(), nil)
	if err != nil {
		return nil, errors.WithStack(
			herodot.ErrInternalServerError.WithWrap(err).WithReasonf("%s", err))
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, errors.WithStack(
			herodot.ErrUpstreamError.WithWrap(err).WithReasonf("%s", err))
	}
	defer func() { _ = resp.Body.Close() }()

	if err := logUpstreamError(p.reg.Logger(), resp); err != nil {
		return nil, err
	}

	// UAE PASS userinfo response — captures all documented attributes across
	// SOP1, SOP2, and SOP3 account types for both Citizens/Residents and Visitors.
	// See https://docs.uaepass.ae/resources/attributes-list
	var user struct {
		Sub           string `json:"sub"`
		UUID          string `json:"uuid"`
		Email         string `json:"email"`
		Mobile        string `json:"mobile"`
		UserType      string `json:"userType"`
		FirstnameEN   string `json:"firstnameEN"`
		LastnameEN    string `json:"lastnameEN"`
		FullnameEN    string `json:"fullnameEN"`
		FirstnameAR   string `json:"firstnameAR"`
		LastnameAR    string `json:"lastnameAR"`
		FullnameAR    string `json:"fullnameAR"`
		Gender        string `json:"gender"`
		NationalityEN string `json:"nationalityEN"`
		NationalityAR string `json:"nationalityAR"`
		Idn           string `json:"idn"`
		IdType        string `json:"idType"`
		SpUUID        string `json:"spuuid"`
		TitleEN       string `json:"titleEN"`
		TitleAR       string `json:"titleAR"`
		ProfileType   string `json:"profileType"`
		UnifiedId     string `json:"unifiedId"`
	}

	// Decode into both the typed struct and a raw map for RawClaims.
	var rawClaims map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&rawClaims); err != nil {
		return nil, errors.WithStack(
			herodot.ErrUpstreamError.WithWrap(err).WithReasonf("%s", err))
	}

	// Re-marshal and unmarshal into the typed struct. This is the standard
	// pattern used by other Kratos providers (e.g., NetID, generic) for
	// populating both typed fields and RawClaims.
	raw, err := json.Marshal(rawClaims)
	if err != nil {
		return nil, errors.WithStack(
			herodot.ErrUpstreamError.WithWrap(err).WithReasonf("%s", err))
	}
	if err := json.Unmarshal(raw, &user); err != nil {
		return nil, errors.WithStack(
			herodot.ErrUpstreamError.WithWrap(err).WithReasonf("%s", err))
	}

	// Prefer UUID as the stable subject identifier; fall back to sub.
	subject := user.UUID
	if subject == "" {
		subject = user.Sub
	}

	return &Claims{
		Issuer:      "https://id.uaepass.ae",
		Subject:     subject,
		GivenName:   user.FirstnameEN,
		FamilyName:  user.LastnameEN,
		Name:        user.FullnameEN,
		Email:       user.Email,
		Gender:      user.Gender,
		PhoneNumber: user.Mobile,
		RawClaims:   rawClaims,
	}, nil
}
