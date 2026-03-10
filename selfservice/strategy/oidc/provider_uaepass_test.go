// Copyright © 2025 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package oidc_test

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"

	"github.com/ory/kratos/pkg"
	"github.com/ory/kratos/selfservice/strategy/oidc"
)

func TestUAEPassOIDCClaims(t *testing.T) {
	t.Parallel()

	handler := http.NewServeMux()
	expectedAccessToken := "my-access-token"

	handler.HandleFunc("GET /userinfo", func(w http.ResponseWriter, r *http.Request) {
		// Verify Bearer token authentication
		authHeader := r.Header.Get("Authorization")
		if authHeader != "Bearer "+expectedAccessToken {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte(`{"error": "invalid_token"}`))
			return
		}

		// UAE PASS citizen/resident profile response
		userProfile := `{
			"sub": "uaepass-user-123456",
			"email": "user@example.ae",
			"fullnameEN": "Ahmed Mohammed Al Rashid",
			"fullnameAR": "أحمد محمد الراشد",
			"firstnameEN": "Ahmed",
			"firstnameAR": "أحمد",
			"lastnameEN": "Al Rashid",
			"lastnameAR": "الراشد",
			"uuid": "550e8400-e29b-41d4-a716-446655440000",
			"unifiedID": "784-1990-1234567-1",
			"idn": "784199012345671",
			"userType": "SOP1",
			"nationalityEN": "ARE",
			"gender": "M",
			"dob": "1990-05-15",
			"mobile": "+971501234567"
		}`

		w.Header().Set("Content-Type", "application/json")
		_, err := w.Write([]byte(userProfile))
		require.NoError(t, err)
	})

	uaePassAPI := httptest.NewServer(handler)
	t.Cleanup(uaePassAPI.Close)

	_, reg := pkg.NewFastRegistryWithMocks(t)
	p := oidc.NewProviderUAEPass(&oidc.Configuration{
		Scope: []string{"urn:uae:digitalid:profile:general"},
	}, reg).(*oidc.ProviderUAEPass)
	p.SetUserInfoURL(t, uaePassAPI.URL+"/userinfo")

	claims, err := p.Claims(t.Context(), &oauth2.Token{AccessToken: expectedAccessToken}, nil)
	require.NoError(t, err)
	require.NotNil(t, claims)

	// Verify standard claims
	assert.Equal(t, "uaepass-user-123456", claims.Subject)
	assert.Equal(t, "user@example.ae", claims.Email)
	assert.Equal(t, "Ahmed Mohammed Al Rashid", claims.Name)
	assert.Equal(t, "Ahmed", claims.GivenName)
	assert.Equal(t, "Al Rashid", claims.LastName)

	// Verify raw claims contain UAE PASS specific data
	require.NotNil(t, claims.RawClaims)
	assert.Equal(t, "550e8400-e29b-41d4-a716-446655440000", claims.RawClaims["uuid"])
	assert.Equal(t, "784-1990-1234567-1", claims.RawClaims["unifiedID"])
	assert.Equal(t, "784199012345671", claims.RawClaims["idn"])
	assert.Equal(t, "SOP1", claims.RawClaims["userType"])
	assert.Equal(t, "ARE", claims.RawClaims["nationalityEN"])
	assert.Equal(t, "M", claims.RawClaims["gender"])
	assert.Equal(t, "1990-05-15", claims.RawClaims["dob"])
	assert.Equal(t, "+971501234567", claims.RawClaims["mobile"])
	assert.Equal(t, "أحمد محمد الراشد", claims.RawClaims["fullnameAR"])
}

func TestUAEPassOIDCClaimsVisitor(t *testing.T) {
	t.Parallel()

	handler := http.NewServeMux()
	expectedAccessToken := "visitor-access-token"

	handler.HandleFunc("GET /userinfo", func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader != "Bearer "+expectedAccessToken {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		// UAE PASS visitor profile response (limited data)
		userProfile := `{
			"sub": "uaepass-visitor-789",
			"email": "visitor@example.com",
			"fullnameEN": "John Smith",
			"firstnameEN": "John",
			"lastnameEN": "Smith",
			"uuid": "660e8400-e29b-41d4-a716-446655440001",
			"userType": "SOP3"
		}`

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(userProfile))
	})

	uaePassAPI := httptest.NewServer(handler)
	t.Cleanup(uaePassAPI.Close)

	_, reg := pkg.NewFastRegistryWithMocks(t)
	p := oidc.NewProviderUAEPass(&oidc.Configuration{
		Scope: []string{"urn:uae:digitalid:profile"},
	}, reg).(*oidc.ProviderUAEPass)
	p.SetUserInfoURL(t, uaePassAPI.URL+"/userinfo")

	claims, err := p.Claims(t.Context(), &oauth2.Token{AccessToken: expectedAccessToken}, nil)
	require.NoError(t, err)
	require.NotNil(t, claims)

	assert.Equal(t, "uaepass-visitor-789", claims.Subject)
	assert.Equal(t, "visitor@example.com", claims.Email)
	assert.Equal(t, "John Smith", claims.Name)
	assert.Equal(t, "SOP3", claims.RawClaims["userType"])
}

func TestUAEPassOIDCClaimsError(t *testing.T) {
	t.Parallel()

	handler := http.NewServeMux()

	handler.HandleFunc("GET /userinfo", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error": "invalid_token", "error_description": "The access token is expired"}`))
	})

	uaePassAPI := httptest.NewServer(handler)
	t.Cleanup(uaePassAPI.Close)

	_, reg := pkg.NewFastRegistryWithMocks(t)
	p := oidc.NewProviderUAEPass(&oidc.Configuration{
		Scope: []string{"urn:uae:digitalid:profile:general"},
	}, reg).(*oidc.ProviderUAEPass)
	p.SetUserInfoURL(t, uaePassAPI.URL+"/userinfo")

	claims, err := p.Claims(t.Context(), &oauth2.Token{AccessToken: "invalid-token"}, nil)
	require.Error(t, err)
	require.Nil(t, claims)
}

func TestUAEPassIssuerURL(t *testing.T) {
	t.Parallel()

	_, reg := pkg.NewFastRegistryWithMocks(t)

	// Test staging via IssuerURL
	pStaging := oidc.NewProviderUAEPass(&oidc.Configuration{
		IssuerURL: "https://stg-id.uaepass.ae",
		Scope:     []string{"urn:uae:digitalid:profile:general"},
	}, reg).(*oidc.ProviderUAEPass)

	oauth2Config, err := pStaging.OAuth2(t.Context())
	require.NoError(t, err)
	assert.Equal(t, "https://stg-id.uaepass.ae/idshub/authorize", oauth2Config.Endpoint.AuthURL)
	assert.Equal(t, "https://stg-id.uaepass.ae/idshub/token", oauth2Config.Endpoint.TokenURL)

	// Test production via IssuerURL
	pProduction := oidc.NewProviderUAEPass(&oidc.Configuration{
		IssuerURL: "https://id.uaepass.ae",
		Scope:     []string{"urn:uae:digitalid:profile:general"},
	}, reg).(*oidc.ProviderUAEPass)

	oauth2Config, err = pProduction.OAuth2(t.Context())
	require.NoError(t, err)
	assert.Equal(t, "https://id.uaepass.ae/idshub/authorize", oauth2Config.Endpoint.AuthURL)
	assert.Equal(t, "https://id.uaepass.ae/idshub/token", oauth2Config.Endpoint.TokenURL)

	// Test default (production) when IssuerURL is not set
	pDefault := oidc.NewProviderUAEPass(&oidc.Configuration{
		Scope: []string{"urn:uae:digitalid:profile:general"},
	}, reg).(*oidc.ProviderUAEPass)

	oauth2Config, err = pDefault.OAuth2(t.Context())
	require.NoError(t, err)
	assert.Equal(t, "https://id.uaepass.ae/idshub/authorize", oauth2Config.Endpoint.AuthURL)
	assert.Equal(t, "https://id.uaepass.ae/idshub/token", oauth2Config.Endpoint.TokenURL)

	// Test trailing slash is handled correctly
	pTrailingSlash := oidc.NewProviderUAEPass(&oidc.Configuration{
		IssuerURL: "https://stg-id.uaepass.ae/",
		Scope:     []string{"urn:uae:digitalid:profile:general"},
	}, reg).(*oidc.ProviderUAEPass)

	oauth2Config, err = pTrailingSlash.OAuth2(t.Context())
	require.NoError(t, err)
	assert.Equal(t, "https://stg-id.uaepass.ae/idshub/authorize", oauth2Config.Endpoint.AuthURL)
	assert.Equal(t, "https://stg-id.uaepass.ae/idshub/token", oauth2Config.Endpoint.TokenURL)
}

func TestUAEPassConfigurationValidation(t *testing.T) {
	t.Parallel()

	_, reg := pkg.NewFastRegistryWithMocks(t)

	// Test PKCE auto falls back to force (no error)
	pPKCEAuto := oidc.NewProviderUAEPass(&oidc.Configuration{
		PKCE:  "auto",
		Scope: []string{"urn:uae:digitalid:profile:general"},
	}, reg).(*oidc.ProviderUAEPass)

	_, err := pPKCEAuto.OAuth2(t.Context())
	require.NoError(t, err)
	assert.Equal(t, "force", pPKCEAuto.Config().PKCE)

	// Test PKCE unset also falls back to force
	pPKCEUnset := oidc.NewProviderUAEPass(&oidc.Configuration{
		Scope: []string{"urn:uae:digitalid:profile:general"},
	}, reg).(*oidc.ProviderUAEPass)

	_, err = pPKCEUnset.OAuth2(t.Context())
	require.NoError(t, err)
	assert.Equal(t, "force", pPKCEUnset.Config().PKCE)

	// Test valid configuration
	pValid := oidc.NewProviderUAEPass(&oidc.Configuration{
		PKCE:  "never",
		Scope: []string{"urn:uae:digitalid:profile:general"},
	}, reg).(*oidc.ProviderUAEPass)

	oauth2Config, err := pValid.OAuth2(t.Context())
	require.NoError(t, err)
	require.NotNil(t, oauth2Config)
}

func TestUAEPassAuthFlowDefaultsToWeb(t *testing.T) {
	t.Parallel()

	_, reg := pkg.NewFastRegistryWithMocks(t)

	p := oidc.NewProviderUAEPass(&oidc.Configuration{
		Scope: []string{"urn:uae:digitalid:profile:general"},
	}, reg).(*oidc.ProviderUAEPass)

	options := p.AuthCodeURLOptions(nil)
	require.Len(t, options, 1)

	oauth2Config, err := p.OAuth2(t.Context())
	require.NoError(t, err)

	authURL := oauth2Config.AuthCodeURL("test-state", options...)
	parsedURL, err := url.Parse(authURL)
	require.NoError(t, err)

	acrValue := parsedURL.Query().Get("acr_values")
	assert.Equal(t, oidc.UAEPassACRWeb, acrValue)
}

func TestUAEPassMobileACRViaUpstreamParams(t *testing.T) {
	t.Parallel()

	_, reg := pkg.NewFastRegistryWithMocks(t)

	p := oidc.NewProviderUAEPass(&oidc.Configuration{
		Scope: []string{"urn:uae:digitalid:profile:general"},
	}, reg).(*oidc.ProviderUAEPass)

	oauth2Config, err := p.OAuth2(t.Context())
	require.NoError(t, err)

	// Simulate the ordering in getAuthRedirectURL: provider defaults first, then upstream params.
	// Upstream acr_values should override the provider's default web ACR value.
	var opts []oauth2.AuthCodeOption
	opts = append(opts, p.AuthCodeURLOptions(nil)...)
	opts = append(opts, oidc.UpstreamParameters(map[string]string{
		"acr_values": "urn:digitalid:authentication:flow:mobileondevice",
	})...)

	authURL := oauth2Config.AuthCodeURL("test-state", opts...)
	parsedURL, err := url.Parse(authURL)
	require.NoError(t, err)

	acrValue := parsedURL.Query().Get("acr_values")
	assert.Equal(t, "urn:digitalid:authentication:flow:mobileondevice", acrValue)
}
