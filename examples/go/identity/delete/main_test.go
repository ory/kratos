// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"testing"

	"github.com/ory/kratos/internal/testhelpers"

	"github.com/ory/kratos/examples/go/pkg"
)

func TestFunc(t *testing.T) {
	publicURL, _ := testhelpers.StartE2EServer(t, "../../pkg/stub/kratos.yaml", nil)
	client = pkg.NewSDKForSelfHosted(publicURL)
	deleteIdentity()
}
