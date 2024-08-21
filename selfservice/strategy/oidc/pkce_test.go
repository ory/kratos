// Copyright © 2024 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package oidc_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/selfservice/strategy/oidc"
	"github.com/ory/x/sqlxx"
)

func TestPKCESupport(t *testing.T) {
	supported := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `{"issuer": "http://%s", "code_challenge_methods_supported":["S256"]}`, r.Host)
	}))
	t.Cleanup(supported.Close)
	notSupported := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `{"issuer": "http://%s", "code_challenge_methods_supported": ["plain"]}`, r.Host)
	}))
	t.Cleanup(notSupported.Close)

	conf, reg := internal.NewFastRegistryWithMocks(t)
	_ = conf

	for _, tc := range []struct {
		c    *oidc.Configuration
		pkce bool
	}{
		{c: &oidc.Configuration{IssuerURL: supported.URL, PKCE: "force"}, pkce: true},
		{c: &oidc.Configuration{IssuerURL: supported.URL, PKCE: "never"}, pkce: false},
		{c: &oidc.Configuration{IssuerURL: supported.URL, PKCE: "auto"}, pkce: true},
		{c: &oidc.Configuration{IssuerURL: supported.URL, PKCE: ""}, pkce: true}, // same as auto

		{c: &oidc.Configuration{IssuerURL: notSupported.URL, PKCE: "force"}, pkce: true},
		{c: &oidc.Configuration{IssuerURL: notSupported.URL, PKCE: "never"}, pkce: false},
		{c: &oidc.Configuration{IssuerURL: notSupported.URL, PKCE: "auto"}, pkce: false},
		{c: &oidc.Configuration{IssuerURL: notSupported.URL, PKCE: ""}, pkce: false}, // same as auto

		{c: &oidc.Configuration{IssuerURL: "", PKCE: "force"}, pkce: true},
		{c: &oidc.Configuration{IssuerURL: "", PKCE: "never"}, pkce: false},
		{c: &oidc.Configuration{IssuerURL: "", PKCE: "auto"}, pkce: false},
		{c: &oidc.Configuration{IssuerURL: "", PKCE: ""}, pkce: false}, // same as auto

	} {
		provider := oidc.NewProviderGenericOIDC(tc.c, reg)

		var flow testFlow
		pkce, err := oidc.MaybeUsePKCE(context.Background(), reg, provider, &flow)
		require.NoError(t, err)
		require.Equal(t, tc.pkce, pkce)
		if tc.pkce {
			require.NotEmpty(t, oidc.PKCEChallenge(&flow))
			require.NotEmpty(t, oidc.PKCEVerifier(&flow))
		} else {
			require.Empty(t, oidc.PKCEChallenge(&flow))
			require.Empty(t, oidc.PKCEVerifier(&flow))
		}
	}

	t.Run("OAuth1", func(t *testing.T) {
		var flow testFlow

		force := oidc.NewProviderX(&oidc.Configuration{IssuerURL: supported.URL, PKCE: "force"}, reg)
		pkce, err := oidc.MaybeUsePKCE(context.Background(), reg, force, &flow)
		require.ErrorContains(t, err, "Provider does not support OAuth2, cannot force PKCE")
		require.False(t, pkce)

		never := oidc.NewProviderX(&oidc.Configuration{IssuerURL: supported.URL, PKCE: "never"}, reg)
		pkce, err = oidc.MaybeUsePKCE(context.Background(), reg, never, &flow)
		require.NoError(t, err)
		require.False(t, pkce)

		auto := oidc.NewProviderX(&oidc.Configuration{IssuerURL: supported.URL, PKCE: "auto"}, reg)
		pkce, err = oidc.MaybeUsePKCE(context.Background(), reg, auto, &flow)
		require.NoError(t, err)
		require.False(t, pkce)
	})
}

type testFlow struct {
	context sqlxx.JSONRawMessage
}

func (t *testFlow) EnsureInternalContext() {
	if t.context == nil {
		t.context = sqlxx.JSONRawMessage("{}")
	}
}

func (t *testFlow) GetInternalContext() sqlxx.JSONRawMessage {
	return t.context
}

func (t *testFlow) SetInternalContext(c sqlxx.JSONRawMessage) {
	t.context = c
}
