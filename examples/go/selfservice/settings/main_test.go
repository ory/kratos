// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"testing"

	ory "github.com/ory/kratos/internal/httpclient"

	"github.com/stretchr/testify/assert"

	"github.com/ory/kratos/internal/testhelpers"

	"github.com/ory/kratos/examples/go/pkg"

	"github.com/stretchr/testify/require"
)

func TestSettings(t *testing.T) {
	publicURL, _ := testhelpers.StartE2EServer(t, "../../pkg/stub/kratos.yaml", nil)
	client = pkg.NewSDKForSelfHosted(publicURL)

	email, password := pkg.RandomCredentials()
	result := changePassword(email, password)
	require.NotEmpty(t, result.Id)
	assert.EqualValues(t, ory.SETTINGSFLOWSTATE_SUCCESS, result.State)

	email, password = pkg.RandomCredentials()
	result = changeTraits(email, password)
	require.NotEmpty(t, result.Id)
	assert.EqualValues(t, ory.SETTINGSFLOWSTATE_SUCCESS, result.State)
	assert.Equal(t, "not-"+email, result.Identity.Traits.(map[string]interface{})["email"].(string))
}
