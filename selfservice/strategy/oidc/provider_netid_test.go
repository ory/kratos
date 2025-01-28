// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package oidc_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/selfservice/strategy/oidc"
)

func TestNetidProvider(t *testing.T) {
	t.Skip("can't test this automatically, because the token is only valid for a short time")
	_, reg := internal.NewVeryFastRegistryWithoutDB(t)

	p := oidc.NewProviderNetID(&oidc.Configuration{
		ClientID: "9b56b26a-e93d-4fce-8f16-951a9858f23e",
	}, reg)

	rawToken := `...`

	claims, err := p.(oidc.IDTokenVerifier).Verify(context.Background(), rawToken)
	require.NoError(t, err)
	require.NotNil(t, claims)
}
