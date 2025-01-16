// Copyright © 2024 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package oidc_test

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/selfservice/strategy/oidc"
)

func TestProviderJackson(t *testing.T) {
	_, reg := internal.NewVeryFastRegistryWithoutDB(t)

	j := oidc.NewProviderJackson(&oidc.Configuration{
		Provider:  "jackson",
		IssuerURL: "https://www.jackson.com/oauth",
		AuthURL:   "https://www.jackson.com/oauth/auth",
		TokenURL:  "https://www.jackson.com/api/oauth/token",
		Mapper:    "file://./stub/hydra.schema.json",
		Scope:     []string{"email", "profile"},
		ID:        "some-id",
	}, reg)
	assert.NotNil(t, j)

	c, err := j.(oidc.OAuth2Provider).OAuth2(context.Background())
	require.NoError(t, err)

	assert.True(t, strings.HasSuffix(c.RedirectURL, "/self-service/methods/saml/callback/some-id"))
}
