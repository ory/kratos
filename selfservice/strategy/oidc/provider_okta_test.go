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

func TestProviderWorkOS(t *testing.T) {
	_, reg := internal.NewVeryFastRegistryWithoutDB(t)

	p := oidc.NewProviderOkta(&oidc.Configuration{
		Provider:        "okta",
		ID:              "test-okta-id",
		ClientID:        "test-client-id",
		ClientSecret:    "test-client-secret",
		IssuerURL:       "https://foo.okta.com",
		Mapper:          "file://./stub/oidc.facebook.jsonnet",
		RequestedClaims: nil,
		Scope:           []string{},
	}, reg)

	c, err := p.(oidc.OAuth2Provider).OAuth2(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "https://foo.okta.com/oauth2/v1/token", c.Endpoint.TokenURL)
	assert.Equal(t, "https://foo.okta.com/oauth2/v1/authorize", c.Endpoint.AuthURL)
}
