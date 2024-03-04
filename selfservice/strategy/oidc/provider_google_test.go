// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package oidc_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"

	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/selfservice/flow/login"
	"github.com/ory/kratos/selfservice/strategy/oidc"
	"github.com/ory/kratos/x"
)

func TestProviderGoogle_Scope(t *testing.T) {
	_, reg := internal.NewVeryFastRegistryWithoutDB(t)

	p := oidc.NewProviderGoogle(&oidc.Configuration{
		Provider:        "google",
		ID:              "valid",
		ClientID:        "client",
		ClientSecret:    "secret",
		Mapper:          "file://./stub/hydra.schema.json",
		RequestedClaims: nil,
		Scope:           []string{"email", "profile", "offline_access"},
	}, reg)

	c, _ := p.(oidc.OAuth2Provider).OAuth2(context.Background())
	assert.NotContains(t, c.Scopes, "offline_access")
}

func TestProviderGoogle_AccessType(t *testing.T) {
	_, reg := internal.NewVeryFastRegistryWithoutDB(t)

	p := oidc.NewProviderGoogle(&oidc.Configuration{
		Provider:        "google",
		ID:              "valid",
		ClientID:        "client",
		ClientSecret:    "secret",
		Mapper:          "file://./stub/hydra.schema.json",
		RequestedClaims: nil,
		Scope:           []string{"email", "profile", "offline_access"},
	}, reg)

	r := &login.Flow{
		ID: x.NewUUID(),
	}

	options := p.(oidc.OAuth2Provider).AuthCodeURLOptions(r)
	assert.Contains(t, options, oauth2.AccessTypeOffline)
}

func TestGoogleVerify(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write(publicJWKS)
	}))

	tsOtherJWKS := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write(publicJWKS2)
	}))

	makeClaims := func(aud string) jwt.RegisteredClaims {
		return jwt.RegisteredClaims{
			Issuer:    "https://accounts.google.com",
			Subject:   "acme@ory.sh",
			Audience:  jwt.ClaimStrings{aud},
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		}
	}
	createProvider := func(jwksUrl string) *oidc.ProviderGoogle {
		_, reg := internal.NewVeryFastRegistryWithoutDB(t)
		p := oidc.NewProviderGoogle(&oidc.Configuration{
			Provider:        "google",
			ID:              "valid",
			ClientID:        "com.example.app",
			ClientSecret:    "secret",
			Mapper:          "file://./stub/hydra.schema.json",
			RequestedClaims: nil,
			Scope:           []string{"email", "profile", "offline_access"},
		}, reg).(*oidc.ProviderGoogle)
		p.JWKSUrl = jwksUrl
		return p
	}
	t.Run("case=successful verification", func(t *testing.T) {
		p := createProvider(ts.URL)
		token := createIdToken(t, makeClaims("com.example.app"))

		c, err := p.Verify(context.Background(), token)
		require.NoError(t, err)
		assert.Equal(t, "acme@ory.sh", c.Email)
		assert.Equal(t, "acme@ory.sh", c.Subject)
		assert.Equal(t, "https://accounts.google.com", c.Issuer)
	})

	t.Run("case=fails due to client_id mismatch", func(t *testing.T) {
		p := createProvider(ts.URL)
		token := createIdToken(t, makeClaims("com.different-example.app"))

		_, err := p.Verify(context.Background(), token)
		require.Error(t, err)
		assert.Equal(t, `token audience didn't match allowed audiences: [com.example.app] oidc: expected audience "com.example.app" got ["com.different-example.app"]`, err.Error())
	})

	t.Run("case=fails due to jwks mismatch", func(t *testing.T) {
		p := createProvider(tsOtherJWKS.URL)
		token := createIdToken(t, makeClaims("com.example.app"))

		_, err := p.Verify(context.Background(), token)
		require.Error(t, err)
		assert.Equal(t, "failed to verify signature: failed to verify id token signature", err.Error())
	})

	t.Run("case=succeedes with additional id token audience", func(t *testing.T) {
		_, reg := internal.NewFastRegistryWithMocks(t)
		google := oidc.NewProviderGoogle(&oidc.Configuration{
			ClientID:                   "something.else.app",
			AdditionalIDTokenAudiences: []string{"com.example.app"},
		}, reg).(*oidc.ProviderGoogle)
		google.JWKSUrl = ts.URL
		token := createIdToken(t, makeClaims("com.example.app"))

		_, err := google.Verify(context.Background(), token)
		require.NoError(t, err)
	})
}
