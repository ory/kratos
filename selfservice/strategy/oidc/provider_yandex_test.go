// Copyright © 2026 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package oidc_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/pkg"
	"github.com/ory/kratos/selfservice/strategy/oidc"
)

func TestProviderYandexOAuthEndpoints(t *testing.T) {
	_, reg := pkg.NewVeryFastRegistryWithoutDB(t)

	t.Run("defaults to oauth.yandex.com", func(t *testing.T) {
		p := oidc.NewProviderYandex(&oidc.Configuration{
			ID:           "yandex",
			Provider:     "yandex",
			ClientID:     "client",
			ClientSecret: "secret",
			Mapper:       "file://./stub/hydra.schema.json",
		}, reg).(oidc.OAuth2Provider)

		c, err := p.OAuth2(context.Background())
		require.NoError(t, err)
		assert.Equal(t, "https://oauth.yandex.com/authorize", c.Endpoint.AuthURL)
		assert.Equal(t, "https://oauth.yandex.com/token", c.Endpoint.TokenURL)
	})

	t.Run("honors auth_url and token_url overrides", func(t *testing.T) {
		p := oidc.NewProviderYandex(&oidc.Configuration{
			ID:           "yandex",
			Provider:     "yandex",
			ClientID:     "client",
			ClientSecret: "secret",
			Mapper:       "file://./stub/hydra.schema.json",
			AuthURL:      "https://oauth.yandex.ru/authorize",
			TokenURL:     "https://oauth.yandex.ru/token",
		}, reg).(oidc.OAuth2Provider)

		c, err := p.OAuth2(context.Background())
		require.NoError(t, err)
		assert.Equal(t, "https://oauth.yandex.ru/authorize", c.Endpoint.AuthURL)
		assert.Equal(t, "https://oauth.yandex.ru/token", c.Endpoint.TokenURL)
	})
}
