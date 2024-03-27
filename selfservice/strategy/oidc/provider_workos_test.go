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

	p := oidc.NewProviderWorkOS(&oidc.Configuration{
		Provider:             "workos",
		ID:                   "demo_organization",
		WorkOSOrganizationId: "demo_organization_id",
		ClientID:             "client",
		ClientSecret:         "secret",
		Mapper:               "file://./stub/hydra.schema.json",
		RequestedClaims:      nil,
		Scope:                []string{},
	}, reg)

	c, err := p.(oidc.OAuth2Provider).OAuth2(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "https://api.workos.com/sso/token", c.Endpoint.TokenURL)
	assert.Equal(t, "https://api.workos.com/sso/authorize?organization=demo_organization_id", c.Endpoint.AuthURL)
}
