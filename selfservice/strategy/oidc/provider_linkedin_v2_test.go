// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package oidc_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/selfservice/strategy/oidc"
)

func TestProviderLinkedInV2_Discovery(t *testing.T) {
	_, reg := internal.NewVeryFastRegistryWithoutDB(t)

	p := oidc.NewProviderLinkedInV2(&oidc.Configuration{
		Provider:        "linkedin_v2",
		ID:              "valid",
		ClientID:        "client",
		ClientSecret:    "secret",
		Mapper:          "file://./stub/hydra.schema.json",
		RequestedClaims: nil,
		Scope:           []string{"email", "profile", "offline_access"},
	}, reg)

	c, err := p.(oidc.OAuth2Provider).OAuth2(context.Background())
	require.NoError(t, err)
	assert.Contains(t, c.Scopes, "openid")
	assert.Equal(t, "https://www.linkedin.com/oauth/v2/accessToken", c.Endpoint.TokenURL)
}
