// Copyright © 2024 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package oidc

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/pkg/errors"
	"golang.org/x/oauth2"

	"github.com/ory/herodot"
	"github.com/ory/x/httpx"
	"github.com/ory/x/stringslice"
	"github.com/ory/x/stringsx"
)

var _ OAuth2Provider = (*ProviderTwitterV2)(nil)

const (
	twitterV2AuthURL     = "https://twitter.com/i/oauth2/authorize"
	twitterV2TokenURL    = "https://api.twitter.com/2/oauth2/token"
	twitterV2UserInfoURL = "https://api.twitter.com/2/users/me"
)

type ProviderTwitterV2 struct {
	config *Configuration
	reg    Dependencies
}

func NewProviderTwitterV2(
	config *Configuration,
	reg Dependencies,
) Provider {
	return &ProviderTwitterV2{
		config: config,
		reg:    reg,
	}
}

func (t *ProviderTwitterV2) Config() *Configuration {
	return t.config
}

func (t *ProviderTwitterV2) oauth2(ctx context.Context) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     t.config.ClientID,
		ClientSecret: t.config.ClientSecret,
		Endpoint: oauth2.Endpoint{
			AuthURL:  twitterV2AuthURL,
			TokenURL: twitterV2TokenURL,
		},
		RedirectURL: t.config.Redir(t.reg.Config().OIDCRedirectURIBase(ctx)),
		Scopes:      t.config.Scope,
	}
}

func (t *ProviderTwitterV2) OAuth2(ctx context.Context) (*oauth2.Config, error) {
	return t.oauth2(ctx), nil
}

func (t *ProviderTwitterV2) AuthCodeURLOptions(r ider) []oauth2.AuthCodeOption {
	// Twitter OAuth2 requires PKCE by default, which is handled by the framework
	// We can optionally add code_challenge_method here if needed
	return []oauth2.AuthCodeOption{
		oauth2.SetAuthURLParam("code_challenge_method", "S256"),
	}
}

func (t *ProviderTwitterV2) Claims(ctx context.Context, exchange *oauth2.Token, query url.Values) (*Claims, error) {
	// Verify granted scopes
	grantedScopes := stringsx.Splitx(fmt.Sprintf("%s", exchange.Extra("scope")), " ")
	for _, check := range t.Config().Scope {
		if !stringslice.Has(grantedScopes, check) {
			return nil, errors.WithStack(ErrScopeMissing)
		}
	}

	// Set up HTTP client with OAuth2 token
	ctx, client := httpx.SetOAuth2(ctx, t.reg.HTTPClient(ctx), t.oauth2(ctx), exchange)

	// Construct user info URL with requested fields
	userInfoURL := t.buildUserInfoURL()

	// Make request to Twitter API
	req, err := retryablehttp.NewRequest("GET", userInfoURL, nil)
	if err != nil {
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("Unable to create Twitter API request: %s", err))
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, errors.WithStack(herodot.ErrUpstreamError.WithWrap(err).WithReasonf("Unable to fetch user info from Twitter: %s", err))
	}
	defer resp.Body.Close()

	if err := logUpstreamError(t.reg.Logger(), resp); err != nil {
		return nil, err
	}

	var response twitterV2UserResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, errors.WithStack(herodot.ErrUpstreamError.WithWrap(err).WithReasonf("Unable to decode Twitter user response: %s", err))
	}

	if response.Data == nil {
		return nil, errors.WithStack(herodot.ErrUpstreamError.WithReasonf("Twitter API returned no user data"))
	}

	user := response.Data

	// Build claims from Twitter user data
	claims := &Claims{
		Issuer:            twitterV2TokenURL,
		Subject:           user.ID,
		Name:              user.Name,
		Nickname:          user.Username,
		PreferredUsername: user.Username,
		RawClaims:         make(map[string]interface{}),
	}

	// Add optional fields if present
	if user.ProfileImageURL != "" {
		claims.Picture = user.ProfileImageURL
	}

	if user.Description != "" {
		claims.RawClaims["description"] = user.Description
	}

	if user.Verified {
		claims.RawClaims["verified"] = user.Verified
	}

	return claims, nil
}

func (t *ProviderTwitterV2) buildUserInfoURL() string {
	// Twitter API v2 requires specifying which user fields to return
	// Default fields: id, name, username
	// Optional fields can be requested via user.fields parameter
	fields := []string{
		"id",
		"name",
		"username",
		"profile_image_url",
		"description",
		"verified",
	}

	u, _ := url.Parse(twitterV2UserInfoURL)
	q := u.Query()
	q.Set("user.fields", strings.Join(fields, ","))
	u.RawQuery = q.Encode()

	return u.String()
}

// twitterV2UserResponse represents the response from Twitter API v2 /users/me endpoint
type twitterV2UserResponse struct {
	Data *twitterV2User `json:"data"`
}

// twitterV2User represents a Twitter user from API v2
type twitterV2User struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	Username        string `json:"username"`
	ProfileImageURL string `json:"profile_image_url,omitempty"`
	Description     string `json:"description,omitempty"`
	Verified        bool   `json:"verified,omitempty"`
}
