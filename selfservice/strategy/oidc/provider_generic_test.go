// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package oidc_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/ory/kratos/driver"
	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/selfservice/strategy/oidc"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/selfservice/flow/login"
	"github.com/ory/kratos/x"
	"github.com/ory/x/contextx"
)

func makeOIDCClaims() json.RawMessage {
	claims, _ := json.Marshal(map[string]interface{}{
		"id_token": map[string]interface{}{
			"email": map[string]bool{
				"essential": true,
			},
			"email_verified": map[string]bool{
				"essential": true,
			},
		},
	})
	return claims
}

func makeAuthCodeURL(t *testing.T, r *login.Flow, reg *driver.RegistryDefault) string {
	p := oidc.NewProviderGenericOIDC(&oidc.Configuration{
		Provider:        "generic",
		ID:              "valid",
		ClientID:        "client",
		ClientSecret:    "secret",
		IssuerURL:       "https://accounts.google.com",
		Mapper:          "file://./stub/hydra.schema.json",
		RequestedClaims: makeOIDCClaims(),
	}, reg)
	c, err := p.(oidc.OAuth2Provider).OAuth2(context.Background())
	require.NoError(t, err)
	return c.AuthCodeURL("state", p.(oidc.OAuth2Provider).AuthCodeURLOptions(r)...)
}

func TestProviderGenericOIDC_AddAuthCodeURLOptions(t *testing.T) {
	ctx := context.Background()
	conf, reg := internal.NewFastRegistryWithMocks(t)
	conf.MustSet(ctx, config.ViperKeyPublicBaseURL, "https://ory.sh")
	t.Run("case=redirectURI is public base url", func(t *testing.T) {
		r := &login.Flow{ID: x.NewUUID(), Refresh: true}
		actual, err := url.ParseRequestURI(makeAuthCodeURL(t, r, reg))
		require.NoError(t, err)
		assert.Contains(t, actual.Query().Get("redirect_uri"), "https://ory.sh")
	})

	t.Run("case=redirectURI is public base url", func(t *testing.T) {
		conf.MustSet(ctx, config.ViperKeyOIDCBaseRedirectURL, "https://example.org")
		t.Cleanup(func() {
			conf.MustSet(ctx, config.ViperKeyOIDCBaseRedirectURL, nil)
		})
		r := &login.Flow{ID: x.NewUUID(), Refresh: true}
		actual, err := url.ParseRequestURI(makeAuthCodeURL(t, r, reg))
		require.NoError(t, err)
		assert.Contains(t, actual.Query().Get("redirect_uri"), "https://example.org")
	})

	t.Run("case=expect prompt to be login with forced flag", func(t *testing.T) {
		r := &login.Flow{
			ID:      x.NewUUID(),
			Refresh: true,
		}
		assert.Contains(t, makeAuthCodeURL(t, r, reg), "prompt=login")
	})

	t.Run("case=expect prompt to not be login without forced flag", func(t *testing.T) {
		r := &login.Flow{
			ID: x.NewUUID(),
		}
		assert.NotContains(t, makeAuthCodeURL(t, r, reg), "prompt=login")
	})

	t.Run("case=expect requested claims to be set", func(t *testing.T) {
		r := &login.Flow{
			ID: x.NewUUID(),
		}
		assert.Contains(t, makeAuthCodeURL(t, r, reg), "claims="+url.QueryEscape(string(makeOIDCClaims())))
	})
}

func TestProviderGenericOIDC_UseOIDCDiscoveryIssuer(t *testing.T) {
	// Simulate an OIDC provider (like Azure AD B2C) where the issuer in the
	// discovery document does not match the discovery URL.
	mismatchedIssuer := "http://different-issuer.example.com"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprintf(w, `{
			"issuer": %q,
			"authorization_endpoint": "http://%s/authorize",
			"token_endpoint": "http://%s/token",
			"jwks_uri": "http://%s/keys",
			"id_token_signing_alg_values_supported": ["RS256"]
		}`, mismatchedIssuer, r.Host, r.Host, r.Host)
	}))
	t.Cleanup(server.Close)

	_, reg := internal.NewFastRegistryWithMocks(t)
	ctx := contextx.WithConfigValue(context.Background(), config.ViperKeyPublicBaseURL, "https://ory.sh")

	t.Run("case=fails when issuer does not match discovery URL", func(t *testing.T) {
		p := oidc.NewProviderGenericOIDC(&oidc.Configuration{
			Provider:               "generic",
			ID:                     "test",
			ClientID:               "client",
			ClientSecret:           "secret",
			IssuerURL:              server.URL,
			UseOIDCDiscoveryIssuer: false,
		}, reg)

		_, err := p.(oidc.OAuth2Provider).OAuth2(ctx)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Invalid configuration")
	})

	t.Run("case=succeeds when use_oidc_discovery_issuer is true", func(t *testing.T) {
		p := oidc.NewProviderGenericOIDC(&oidc.Configuration{
			Provider:               "generic",
			ID:                     "test",
			ClientID:               "client",
			ClientSecret:           "secret",
			IssuerURL:              server.URL,
			UseOIDCDiscoveryIssuer: true,
		}, reg)

		c, err := p.(oidc.OAuth2Provider).OAuth2(ctx)
		require.NoError(t, err)
		assert.Contains(t, c.Endpoint.AuthURL, server.URL)
	})

	t.Run("case=uses discovered endpoints not config auth_url/token_url", func(t *testing.T) {
		p := oidc.NewProviderGenericOIDC(&oidc.Configuration{
			Provider:               "generic",
			ID:                     "test",
			ClientID:               "client",
			ClientSecret:           "secret",
			IssuerURL:              server.URL,
			AuthURL:                "https://should-be-ignored.example.com/authorize",
			TokenURL:               "https://should-be-ignored.example.com/token",
			UseOIDCDiscoveryIssuer: true,
		}, reg)

		c, err := p.(oidc.OAuth2Provider).OAuth2(ctx)
		require.NoError(t, err)
		assert.NotContains(t, c.Endpoint.AuthURL, "should-be-ignored")
		assert.Contains(t, c.Endpoint.AuthURL, server.URL)
	})
}
