// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"testing"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/examples/go/pkg"
	ory "github.com/ory/kratos/internal/httpclient"
	"github.com/ory/kratos/internal/testhelpers"
)

func TestFunc(t *testing.T) {
	publicURL, _ := testhelpers.StartE2EServer(t, "../../pkg/stub/kratos.yaml", nil)
	client = pkg.NewSDKForSelfHosted(publicURL)

	flow := performVerification("dev+" + uuid.Must(uuid.NewV4()).String() + "@ory.sh")
	require.NotEmpty(t, flow.Id)
	assert.EqualValues(t, ory.VERIFICATIONFLOWSTATE_SENT_EMAIL, flow.State)
}
