// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/examples/go/pkg"
)

func TestFunc(t *testing.T) {
	client = pkg.TestClient(t)
	require.NotEmpty(t, createIdentity().Id)
}
