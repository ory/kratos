// Copyright © 2024 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package oidc_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/selfservice/strategy/oidc"
	"github.com/ory/kratos/x"
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
	strat := oidc.NewStrategy(reg)

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

		stateParam, pkce, err := strat.GenerateState(context.Background(), provider, x.NewUUID())
		require.NoError(t, err)
		require.NotEmpty(t, stateParam)

		state, err := oidc.DecryptState(context.Background(), reg.Cipher(context.Background()), stateParam)
		require.NoError(t, err)

		if tc.pkce {
			require.NotEmpty(t, pkce)
			require.NotEmpty(t, oidc.PKCEVerifier(state))
		} else {
			require.Empty(t, pkce)
			require.Empty(t, oidc.PKCEVerifier(state))
		}
	}

	t.Run("OAuth1", func(t *testing.T) {
		for _, provider := range []oidc.Provider{
			oidc.NewProviderX(&oidc.Configuration{IssuerURL: supported.URL, PKCE: "force"}, reg),
			oidc.NewProviderX(&oidc.Configuration{IssuerURL: supported.URL, PKCE: "never"}, reg),
			oidc.NewProviderX(&oidc.Configuration{IssuerURL: supported.URL, PKCE: "auto"}, reg),
		} {
			stateParam, pkce, err := strat.GenerateState(context.Background(), provider, x.NewUUID())
			require.NoError(t, err)
			require.NotEmpty(t, stateParam)
			assert.Empty(t, pkce)

			state, err := oidc.DecryptState(context.Background(), reg.Cipher(context.Background()), stateParam)
			require.NoError(t, err)
			assert.Empty(t, oidc.PKCEVerifier(state))
		}
	})
}
