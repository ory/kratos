// Copyright © 2024 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package claims_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/selfservice/strategy/oidc/claims"
)

func TestClaimsValidate(t *testing.T) {
	require.Error(t, new(claims.Claims).Validate())
	require.Error(t, (&claims.Claims{Issuer: "not-empty"}).Validate())
	require.Error(t, (&claims.Claims{Issuer: "not-empty"}).Validate())
	require.Error(t, (&claims.Claims{Subject: "not-empty"}).Validate())
	require.Error(t, (&claims.Claims{Subject: "not-empty"}).Validate())
	require.NoError(t, (&claims.Claims{Issuer: "not-empty", Subject: "not-empty"}).Validate())
}
