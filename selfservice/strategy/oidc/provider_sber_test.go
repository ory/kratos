// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package oidc_test

import (
	"context"
	"net/url"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"

	"github.com/ory/kratos/pkg"
	"github.com/ory/kratos/selfservice/strategy/oidc"

	"testing"
)

func TestProviderSber_OAuth2(t *testing.T) {
	_, reg := pkg.NewVeryFastRegistryWithoutDB(t)

	p := oidc.NewProviderSber(&oidc.Configuration{
		ClientID:     "client",
		ClientSecret: "secret",
		Scope:        []string{"openid"},
	}, reg)

	c, err := p.(oidc.OAuth2Provider).OAuth2(context.Background())
	require.NoError(t, err)

	assert.Contains(t, c.Endpoint.AuthURL, "sberbank")
	assert.Contains(t, c.Endpoint.TokenURL, "sber")
}

func TestProviderSber_AuthCodeURLOptions(t *testing.T) {
	_, reg := pkg.NewVeryFastRegistryWithoutDB(t)

	p := oidc.NewProviderSber(&oidc.Configuration{}, reg)

	opts := p.(oidc.OAuth2Provider).AuthCodeURLOptions(nil)

	assert.NotEmpty(t, opts)
}

func TestProviderSber_Claims_IDToken(t *testing.T) {
	_, reg := pkg.NewVeryFastRegistryWithoutDB(t)

	p := oidc.NewProviderSber(&oidc.Configuration{}, reg)

	token := (&oauth2.Token{}).WithExtra(map[string]interface{}{
		"id_token": fakeJWTJWKS,
	})

	claims, err := p.(oidc.OAuth2Provider).Claims(context.Background(), token, url.Values{})
	require.NoError(t, err)

	assert.NotEmpty(t, claims.Subject)
}
