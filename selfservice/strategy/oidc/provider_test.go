// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package oidc

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestClaimsValidate(t *testing.T) {
	require.Error(t, new(Claims).Validate())
	require.Error(t, (&Claims{Issuer: "not-empty"}).Validate())
	require.Error(t, (&Claims{Issuer: "not-empty"}).Validate())
	require.Error(t, (&Claims{Subject: "not-empty"}).Validate())
	require.Error(t, (&Claims{Subject: "not-empty"}).Validate())
	require.NoError(t, (&Claims{Issuer: "not-empty", Subject: "not-empty"}).Validate())
}
