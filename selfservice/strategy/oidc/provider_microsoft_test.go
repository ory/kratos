// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package oidc_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	_ "embed"

	"github.com/golang-jwt/jwt/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/selfservice/strategy/oidc"
)

func TestMicrosoftVerify(t *testing.T) {
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
			Issuer:    "https://login.microsoftonline.com/tenant_id/v2.0",
			Subject:   "acme@ory.sh",
			Audience:  jwt.ClaimStrings{aud},
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		}
	}
	t.Run("case=successful verification", func(t *testing.T) {
		_, reg := internal.NewFastRegistryWithMocks(t)
		apple := oidc.NewProviderMicrosoft(&oidc.Configuration{
			ClientID: "com.example.app",
			Tenant:   "tenant_id",
		}, reg).(*oidc.ProviderMicrosoft)
		apple.JWKSUrl = ts.URL
		token := createIdToken(t, makeClaims("com.example.app"))

		c, err := apple.Verify(context.Background(), token)
		require.NoError(t, err)
		assert.Equal(t, "acme@ory.sh", c.Email)
		assert.Equal(t, "acme@ory.sh", c.Subject)
		assert.Equal(t, "https://login.microsoftonline.com/tenant_id/v2.0", c.Issuer)
	})

	t.Run("case=fails due to client_id mismatch", func(t *testing.T) {
		_, reg := internal.NewFastRegistryWithMocks(t)
		apple := oidc.NewProviderMicrosoft(&oidc.Configuration{
			ClientID: "com.example.app",
			Tenant:   "tenant_id",
		}, reg).(*oidc.ProviderMicrosoft)
		apple.JWKSUrl = ts.URL
		token := createIdToken(t, makeClaims("com.different-example.app"))

		_, err := apple.Verify(context.Background(), token)
		require.Error(t, err)
		assert.Equal(t, `token audience didn't match allowed audiences: [com.example.app] oidc: expected audience "com.example.app" got ["com.different-example.app"]`, err.Error())
	})

	t.Run("case=fails due to jwks mismatch", func(t *testing.T) {
		_, reg := internal.NewFastRegistryWithMocks(t)
		apple := oidc.NewProviderMicrosoft(&oidc.Configuration{
			ClientID: "com.example.app",
			Tenant:   "tenant_id",
		}, reg).(*oidc.ProviderMicrosoft)
		apple.JWKSUrl = tsOtherJWKS.URL
		token := createIdToken(t, makeClaims("com.example.app"))

		_, err := apple.Verify(context.Background(), token)
		require.Error(t, err)
		assert.Equal(t, "failed to verify signature: failed to verify id token signature", err.Error())
	})

	t.Run("case=fails due to wrong issuer tenant", func(t *testing.T) {
		_, reg := internal.NewFastRegistryWithMocks(t)
		apple := oidc.NewProviderMicrosoft(&oidc.Configuration{
			ClientID: "com.example.app",
			Tenant:   "wrong_tenant_id",
		}, reg).(*oidc.ProviderMicrosoft)
		apple.JWKSUrl = tsOtherJWKS.URL
		token := createIdToken(t, makeClaims("com.example.app"))

		_, err := apple.Verify(context.Background(), token)
		require.Error(t, err)
		assert.Equal(t, "oidc: id token issued by a different provider, expected \"https://login.microsoftonline.com/wrong_tenant_id/v2.0\" got \"https://login.microsoftonline.com/tenant_id/v2.0\"", err.Error())
	})

	t.Run("case=succeedes with additional id token audience", func(t *testing.T) {
		_, reg := internal.NewFastRegistryWithMocks(t)
		apple := oidc.NewProviderMicrosoft(&oidc.Configuration{
			ClientID:                   "something.else.app",
			Tenant:                     "tenant_id",
			AdditionalIDTokenAudiences: []string{"com.example.app"},
		}, reg).(*oidc.ProviderMicrosoft)
		apple.JWKSUrl = ts.URL
		token := createIdToken(t, makeClaims("com.example.app"))

		_, err := apple.Verify(context.Background(), token)
		require.NoError(t, err)
	})
}
