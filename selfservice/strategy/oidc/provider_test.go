// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package oidc

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
)

func TestClaimsValidate(t *testing.T) {
	require.Error(t, new(Claims).Validate())
	require.Error(t, (&Claims{Issuer: "not-empty"}).Validate())
	require.Error(t, (&Claims{Issuer: "not-empty"}).Validate())
	require.Error(t, (&Claims{Subject: "not-empty"}).Validate())
	require.Error(t, (&Claims{Subject: "not-empty"}).Validate())
	require.NoError(t, (&Claims{Issuer: "not-empty", Subject: "not-empty"}).Validate())
}

type TestProvider struct {
	*ProviderGenericOIDC
}

func NewTestProvider(c *Configuration, reg Dependencies) Provider {
	return &TestProvider{
		ProviderGenericOIDC: NewProviderGenericOIDC(c, reg).(*ProviderGenericOIDC),
	}
}

func RegisterTestProvider(t *testing.T, id string) {
	supportedProviders[id] = func(c *Configuration, reg Dependencies) Provider {
		return NewTestProvider(c, reg)
	}
	t.Cleanup(func() {
		delete(supportedProviders, id)
	})
}

var _ IDTokenVerifier = new(TestProvider)

func (t *TestProvider) Verify(_ context.Context, token string) (*Claims, error) {
	if token == "error" {
		return nil, fmt.Errorf("stub error")
	}
	c := Claims{}
	if err := json.Unmarshal([]byte(token), &c); err != nil {
		return nil, err
	}
	return &c, nil
}

// SetUserInfoURL allows overriding the userinfo URL for testing purposes.
func (p *ProviderUAEPass) SetUserInfoURL(t *testing.T, url string) {
	t.Helper()
	p.userinfoURL = url
}

// upstreamParam reports whether the option list produced by
// UpstreamParameters contains a given upstream parameter, and its value.
// The options are applied through a real oauth2.Config.AuthCodeURL so the
// test observes exactly what is sent to the upstream provider.
func upstreamParam(t *testing.T, opts []oauth2.AuthCodeOption, key string) (string, bool) {
	t.Helper()
	conf := &oauth2.Config{
		ClientID: "client",
		Endpoint: oauth2.Endpoint{AuthURL: "https://idp.example.com/auth"},
	}
	authURL, err := url.Parse(conf.AuthCodeURL("state", opts...))
	require.NoError(t, err)
	q := authURL.Query()
	if !q.Has(key) {
		return "", false
	}
	return q.Get(key), true
}

func TestUpstreamParameters(t *testing.T) {
	t.Parallel()

	upstream := map[string]string{
		"login_hint": "victim@example.com",
		"hd":         "example.com",
		"prompt":     "select_account",
		"auth_type":  "reauthenticate",
		"acr_values": "urn:example:acr",
		"unknown":    "dropped",
	}

	opts := UpstreamParameters(upstream)

	t.Run("case=login_hint is never forwarded", func(t *testing.T) {
		// A caller-supplied login_hint cannot be trusted and can steer
		// account/tenant/client selection on some IdPs, enabling a client-swap
		// attack (HackerOne #3239672, ory-corp/cloud#8955).
		_, ok := upstreamParam(t, opts, "login_hint")
		assert.False(t, ok, "login_hint must not be forwarded")
	})

	t.Run("case=other allowed parameters are forwarded", func(t *testing.T) {
		for key, want := range map[string]string{
			"hd":         "example.com",
			"prompt":     "select_account",
			"auth_type":  "reauthenticate",
			"acr_values": "urn:example:acr",
		} {
			got, ok := upstreamParam(t, opts, key)
			assert.True(t, ok, "%s must be forwarded", key)
			assert.Equal(t, want, got)
		}
	})

	t.Run("case=unknown parameters are dropped", func(t *testing.T) {
		_, ok := upstreamParam(t, opts, "unknown")
		assert.False(t, ok, "unknown parameters must not be forwarded")
	})
}

func TestLocale(t *testing.T) {
	// test json unmarshal
	for _, tc := range []struct {
		name      string
		json      string
		expected  string
		assertErr assert.ErrorAssertionFunc
	}{{
		name:     "empty",
		json:     `{}`,
		expected: "",
	}, {
		name:     "empty string locale",
		json:     `{"locale":""}`,
		expected: "",
	}, {
		name:      "invalid string locale",
		json:      `{"locale":"""}`,
		assertErr: assert.Error,
	}, {
		name:     "string locale",
		json:     `{"locale":"en-US"}`,
		expected: "en-US",
	}, {
		name:     "linkedin locale",
		json:     `{"locale":{"country":"US","language":"en","ignore":"me"}}`,
		expected: "en-US",
	}, {
		name:     "missing country linkedin locale",
		json:     `{"locale":{"language":"en"}}`,
		expected: "en",
	}, {
		name:     "missing language linkedin locale",
		json:     `{"locale":{"country":"US"}}`,
		expected: "US",
	}, {
		name:     "invalid linkedin locale",
		json:     `{"locale":{"invalid":"me"}}`,
		expected: "",
	}} {
		t.Run(tc.name, func(t *testing.T) {
			var c Claims
			err := json.Unmarshal([]byte(tc.json), &c)
			if tc.assertErr != nil {
				tc.assertErr(t, err)
				return
			}
			require.NoError(t, err)
			assert.EqualValues(t, tc.expected, c.Locale)
		})
	}
}
