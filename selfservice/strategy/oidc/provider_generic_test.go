// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package oidc_test

import (
	"context"
	"encoding/json"
	"net/url"
	"testing"

	"golang.org/x/oauth2"

	"github.com/ory/kratos/driver"
	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/selfservice/strategy/oidc"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/selfservice/flow/login"
	"github.com/ory/kratos/x"
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

func makeAuthCodeURL(t *testing.T, r *login.Flow, reg *driver.RegistryDefault, pkcsMethods ...string) string {
	var pkcsMethod string
	if len(pkcsMethods) > 0 {
		pkcsMethod = pkcsMethods[0]
	} else {
		pkcsMethod = ""
	}
	p := oidc.NewProviderGenericOIDC(&oidc.Configuration{
		Provider:        "generic",
		ID:              "valid",
		ClientID:        "client",
		ClientSecret:    "secret",
		IssuerURL:       "https://accounts.google.com",
		Mapper:          "file://./stub/hydra.schema.json",
		PKCSMethod:      pkcsMethod,
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

func TestProviderGenericOIDC_PKCS(t *testing.T) {
	ctx := context.Background()
	conf, reg := internal.NewFastRegistryWithMocks(t)
	conf.MustSet(ctx, config.ViperKeyPublicBaseURL, "https://ory.sh")

	t.Run("case=PKCSMethod is set to S256", func(t *testing.T) {
		r := &login.Flow{ID: x.NewUUID(), Refresh: true}
		reg.LoginFlowPersister().CreateLoginFlow(ctx, r)
		err := oidc.SetPKCSContext(r, oidc.PkcsContext{
			Method:   "S256",
			Verifier: oauth2.GenerateVerifier(),
		})
		require.NoError(t, err)
		err = reg.LoginFlowPersister().UpdateLoginFlow(ctx, r)
		require.NoError(t, err)
		actual, err := url.ParseRequestURI(makeAuthCodeURL(t, r, reg, "S256"))
		require.NoError(t, err)
		assert.Contains(t, actual.Query(), "code_challenge")
		t.Logf("code_challenge: %s", actual.Query().Get("code_challenge"))
		assert.Contains(t, actual.Query().Get("code_challenge_method"), "S256")
		t.Logf("code_challenge_method: %s", actual.Query().Get("code_challenge_method"))
	})
	t.Run("case=PKCSMethod is set to plain", func(t *testing.T) {
		r := &login.Flow{ID: x.NewUUID(), Refresh: true}
		reg.LoginFlowPersister().CreateLoginFlow(ctx, r)
		verifier := oauth2.GenerateVerifier()
		err := oidc.SetPKCSContext(r, oidc.PkcsContext{
			Method:   "plain",
			Verifier: verifier,
		})
		require.NoError(t, err)
		err = reg.LoginFlowPersister().UpdateLoginFlow(ctx, r)
		require.NoError(t, err)
		actual, err := url.ParseRequestURI(makeAuthCodeURL(t, r, reg, "plain"))
		require.NoError(t, err)
		assert.Contains(t, actual.Query(), "code_challenge")
		t.Logf("code_challenge: %s", actual.Query().Get("code_challenge"))
		assert.Contains(t, actual.Query().Get("code_challenge_method"), "plain")
		t.Logf("code_challenge_method: %s", actual.Query().Get("code_challenge_method"))
		assert.Equal(t, actual.Query().Get("code_challenge"), verifier)
	})
	t.Run("case=PKCSMethod is empty", func(t *testing.T) {
		r := &login.Flow{ID: x.NewUUID(), Refresh: true}
		actual, err := url.ParseRequestURI(makeAuthCodeURL(t, r, reg))
		require.NoError(t, err)
		assert.NotContains(t, actual.Query(), "code_challenge")
		t.Logf("code_challenge: %s", actual.Query().Get("code_challenge"))
		assert.NotContains(t, actual.Query(), "code_challenge_method")
		t.Logf("code_challenge_method: %s", actual.Query().Get("code_challenge_method"))
	})
}
