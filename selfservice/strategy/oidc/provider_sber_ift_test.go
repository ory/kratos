// Copyright © 2026 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package oidc_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/herodot"

	"github.com/ory/kratos/pkg"
	"github.com/ory/kratos/selfservice/strategy/oidc"
)

func TestProviderSberIft_DefaultPKCE(t *testing.T) {
	_, reg := pkg.NewVeryFastRegistryWithoutDB(t)

	cfg := &oidc.Configuration{
		ID:           "sber-ift",
		ClientID:     "client",
		ClientSecret: "secret",
	}
	p := oidc.NewProviderSberIft(cfg, reg)

	assert.Equal(t, "auto", p.Config().PKCE)
}

func TestProviderSberIft_ExchangeRequest(t *testing.T) {
	_, reg := pkg.NewVeryFastRegistryWithoutDB(t)

	var (
		gotMethod      string
		gotContentType string
		gotAccept      string
		gotRqUID       string
		gotForm        url.Values
	)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotContentType = r.Header.Get("Content-Type")
		gotAccept = r.Header.Get("accept")
		gotRqUID = r.Header.Get("RqUID")

		err := r.ParseForm()
		require.NoError(t, err)
		gotForm = r.Form

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"access_token":  "token",
			"token_type":    "bearer",
			"expires_in":    60,
			"id_token":      "id-token",
			"refresh_token": "refresh-token",
		})
	}))
	t.Cleanup(ts.Close)

	cfg := &oidc.Configuration{
		ID:           "sber-ift",
		ClientID:     "client-id",
		ClientSecret: "client-secret",
		Scope:        []string{"openid"},
		AuthURL:      ts.URL + "/authorize",
		TokenURL:     ts.URL,
	}
	p := oidc.NewProviderSberIft(cfg, reg).(oidc.OAuth2TokenExchanger)

	token, err := p.Exchange(context.Background(), "auth-code")
	require.NoError(t, err)
	require.NotNil(t, token)

	assert.Equal(t, http.MethodPost, gotMethod)
	assert.Contains(t, gotContentType, "application/x-www-form-urlencoded")
	assert.Equal(t, "application/json", gotAccept)
	assert.Len(t, gotRqUID, 32)

	assert.Equal(t, "authorization_code", gotForm.Get("grant_type"))
	assert.Equal(t, "auth-code", gotForm.Get("code"))
	assert.Equal(t, "client-id", gotForm.Get("client_id"))
	assert.Equal(t, "client-secret", gotForm.Get("client_secret"))
	assert.NotEmpty(t, gotForm.Get("redirect_uri"))
}

func TestProviderSberIft_ExchangeUnauthorized(t *testing.T) {
	_, reg := pkg.NewVeryFastRegistryWithoutDB(t)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = io.WriteString(w, `{"error":"invalid_client","client_secret":"super-secret","access_token":"very-secret-token"}`)
	}))
	t.Cleanup(ts.Close)

	cfg := &oidc.Configuration{
		ID:           "sber-ift",
		ClientID:     "client-id",
		ClientSecret: "client-secret",
		Scope:        []string{"openid"},
		AuthURL:      ts.URL + "/authorize",
		TokenURL:     ts.URL,
	}
	p := oidc.NewProviderSberIft(cfg, reg).(oidc.OAuth2TokenExchanger)

	_, err := p.Exchange(context.Background(), "auth-code")
	require.Error(t, err)

	var he *herodot.DefaultError
	require.ErrorAs(t, err, &he, "%+v", err)
	reason := he.Reason()
	assert.Contains(t, reason, "sber token exchange unauthorized")
	assert.Contains(t, reason, "stage=token_exchange")
	assert.Contains(t, reason, "http_status=401")
	assert.Contains(t, reason, "provider=sber-ift")
	assert.Contains(t, reason, "curl_request=")
	assert.Contains(t, reason, "--request POST")
	assert.Contains(t, reason, "--header 'rquid:")
}
