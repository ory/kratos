// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package oidc_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	_ "embed"

	"github.com/golang-jwt/jwt/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/selfservice/strategy/oidc"
)

func TestDecodeQuery(t *testing.T) {
	query := url.Values{
		"user": []string{`{"name": {"firstName": "first", "lastName": "last"}, "email": "email@email.com"}`},
	}

	for k, tc := range []struct {
		claims     *oidc.Claims
		familyName string
		givenName  string
		lastName   string
	}{
		{claims: &oidc.Claims{}, familyName: "last", givenName: "first", lastName: "last"},
		{claims: &oidc.Claims{FamilyName: "fam"}, familyName: "fam", givenName: "first", lastName: "last"},
		{claims: &oidc.Claims{FamilyName: "fam", GivenName: "giv"}, familyName: "fam", givenName: "giv", lastName: "last"},
		{claims: &oidc.Claims{FamilyName: "fam", GivenName: "giv", LastName: "las"}, familyName: "fam", givenName: "giv", lastName: "las"},
	} {
		t.Run(fmt.Sprintf("case=%d", k), func(t *testing.T) {
			a := oidc.NewProviderApple(&oidc.Configuration{}, nil).(*oidc.ProviderApple)
			a.DecodeQuery(query, tc.claims)
			assert.Equal(t, tc.familyName, tc.claims.FamilyName)
			assert.Equal(t, tc.givenName, tc.claims.GivenName)
			assert.Equal(t, tc.lastName, tc.claims.LastName)
			// Never extract email from the query, as the same info can be extracted and verified from the ID token.
			assert.Empty(t, tc.claims.Email)
		})
	}
}

func TestAppleVerify(t *testing.T) {
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
			Issuer:    "https://appleid.apple.com",
			Subject:   "acme@ory.sh",
			Audience:  jwt.ClaimStrings{aud},
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		}
	}
	t.Run("case=successful verification", func(t *testing.T) {
		_, reg := internal.NewFastRegistryWithMocks(t)
		apple := oidc.NewProviderApple(&oidc.Configuration{
			ClientID: "com.example.app",
		}, reg).(*oidc.ProviderApple)
		apple.JWKSUrl = ts.URL
		token := createIdToken(t, makeClaims("com.example.app"))

		c, err := apple.Verify(context.Background(), token)
		require.NoError(t, err)
		assert.Equal(t, "acme@ory.sh", c.Email)
		assert.Equal(t, "acme@ory.sh", c.Subject)
		assert.Equal(t, "https://appleid.apple.com", c.Issuer)
	})

	t.Run("case=fails due to client_id mismatch", func(t *testing.T) {
		_, reg := internal.NewFastRegistryWithMocks(t)
		apple := oidc.NewProviderApple(&oidc.Configuration{
			ClientID: "com.example.app",
		}, reg).(*oidc.ProviderApple)
		apple.JWKSUrl = ts.URL
		token := createIdToken(t, makeClaims("com.different-example.app"))

		_, err := apple.Verify(context.Background(), token)
		require.Error(t, err)
		assert.Equal(t, `token audience didn't match allowed audiences: [com.example.app] oidc: expected audience "com.example.app" got ["com.different-example.app"]`, err.Error())
	})

	t.Run("case=fails due to jwks mismatch", func(t *testing.T) {
		_, reg := internal.NewFastRegistryWithMocks(t)
		apple := oidc.NewProviderApple(&oidc.Configuration{
			ClientID: "com.example.app",
		}, reg).(*oidc.ProviderApple)
		apple.JWKSUrl = tsOtherJWKS.URL
		token := createIdToken(t, makeClaims("com.example.app"))

		_, err := apple.Verify(context.Background(), token)
		require.Error(t, err)
		assert.Equal(t, "failed to verify signature: failed to verify id token signature", err.Error())
	})

	t.Run("case=succeedes with additional id token audience", func(t *testing.T) {
		_, reg := internal.NewFastRegistryWithMocks(t)
		apple := oidc.NewProviderApple(&oidc.Configuration{
			ClientID:                   "something.else.app",
			AdditionalIDTokenAudiences: []string{"com.example.app"},
		}, reg).(*oidc.ProviderApple)
		apple.JWKSUrl = ts.URL
		token := createIdToken(t, makeClaims("com.example.app"))

		_, err := apple.Verify(context.Background(), token)
		require.NoError(t, err)
	})
}
