// Copyright © 2024 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package oidc_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"

	"github.com/ory/herodot"
	"github.com/ory/kratos/pkg"
	"github.com/ory/kratos/selfservice/strategy/oidc"
)

func TestProviderJackson(t *testing.T) {
	_, reg := pkg.NewVeryFastRegistryWithoutDB(t)

	j := oidc.NewProviderJackson(&oidc.Configuration{
		Provider:  "jackson",
		IssuerURL: "https://www.jackson.com/oauth",
		AuthURL:   "https://www.jackson.com/oauth/auth",
		TokenURL:  "https://www.jackson.com/api/oauth/token",
		Mapper:    "file://./stub/hydra.schema.json",
		Scope:     []string{"email", "profile"},
		ID:        "some-id",
	}, reg)
	assert.NotNil(t, j)

	c, err := j.(oidc.OAuth2Provider).OAuth2(context.Background())
	require.NoError(t, err)

	assert.True(t, strings.HasSuffix(c.RedirectURL, "/self-service/methods/saml/callback/some-id"))
}

func TestProviderJacksonClaims(t *testing.T) {
	jwks := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		_, _ = w.Write(publicJWKS)
	}))
	t.Cleanup(jwks.Close)

	otherJWKS := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		_, _ = w.Write(publicJWKS2)
	}))
	t.Cleanup(otherJWKS.Close)

	const issuer = "https://sp.example.org/saml"

	newJackson := func(t *testing.T, upstream *httptest.Server) oidc.OAuth2Provider {
		_, reg := pkg.NewVeryFastRegistryWithoutDB(t)
		return oidc.NewProviderJackson(&oidc.Configuration{
			Provider:  "jackson",
			ID:        "jackson",
			ClientID:  "client-id",
			IssuerURL: issuer,
			AuthURL:   upstream.URL + "/api/oauth/authorize",
			TokenURL:  upstream.URL + "/api/oauth/token",
			Mapper:    "file://./stub/hydra.schema.json",
		}, reg).(oidc.OAuth2Provider)
	}

	makeExchange := func(t *testing.T, iss string) *oauth2.Token {
		raw := createIDToken(t, jwt.RegisteredClaims{
			Issuer:    iss,
			Subject:   "acme@ory.sh",
			Audience:  jwt.ClaimStrings{"client-id"},
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		})
		return (&oauth2.Token{}).WithExtra(map[string]interface{}{"id_token": raw})
	}

	t.Run("case=accepts the configured issuer", func(t *testing.T) {
		claims, err := newJackson(t, jwks).Claims(t.Context(), makeExchange(t, issuer), url.Values{})
		require.NoError(t, err)
		assert.Equal(t, issuer, claims.Issuer)
		assert.Equal(t, "acme@ory.sh", claims.Subject)
		assert.Equal(t, "acme@ory.sh", claims.Email)
	})

	requireReasonContains := func(t *testing.T, err error, substr string) {
		t.Helper()
		he := new(herodot.DefaultError)
		require.ErrorAs(t, err, &he)
		assert.Contains(t, he.ReasonField, substr)
	}

	t.Run("case=rejects an unknown issuer", func(t *testing.T) {
		_, err := newJackson(t, jwks).Claims(t.Context(), makeExchange(t, "https://old.example.org/saml"), url.Values{})
		requireReasonContains(t, err, "https://old.example.org/saml")
	})

	t.Run("case=accepts an additional trusted issuer", func(t *testing.T) {
		t.Setenv("SAML_SP_TRUSTED_ISSUERS", "https://old.example.org/saml, https://other.example.org/saml")

		claims, err := newJackson(t, jwks).Claims(t.Context(), makeExchange(t, "https://old.example.org/saml"), url.Values{})
		require.NoError(t, err)
		assert.Equal(t, "https://old.example.org/saml", claims.Issuer)

		claims, err = newJackson(t, jwks).Claims(t.Context(), makeExchange(t, "https://other.example.org/saml"), url.Values{})
		require.NoError(t, err)
		assert.Equal(t, "https://other.example.org/saml", claims.Issuer)

		claims, err = newJackson(t, jwks).Claims(t.Context(), makeExchange(t, issuer), url.Values{})
		require.NoError(t, err)
		assert.Equal(t, issuer, claims.Issuer)
	})

	t.Run("case=rejects an issuer outside the trusted list", func(t *testing.T) {
		t.Setenv("SAML_SP_TRUSTED_ISSUERS", "https://old.example.org/saml")

		_, err := newJackson(t, jwks).Claims(t.Context(), makeExchange(t, "https://evil.example.org/saml"), url.Values{})
		requireReasonContains(t, err, "https://evil.example.org/saml")
	})

	t.Run("case=rejects a token signed by an unknown key even for a trusted issuer", func(t *testing.T) {
		t.Setenv("SAML_SP_TRUSTED_ISSUERS", "https://old.example.org/saml")

		_, err := newJackson(t, otherJWKS).Claims(t.Context(), makeExchange(t, "https://old.example.org/saml"), url.Values{})
		requireReasonContains(t, err, "failed to verify signature")
	})

	t.Run("case=rejects a missing id token", func(t *testing.T) {
		_, err := newJackson(t, jwks).Claims(t.Context(), &oauth2.Token{}, url.Values{})
		require.Error(t, err)
	})
}
