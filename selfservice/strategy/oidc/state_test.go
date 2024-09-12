// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package oidc_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/cipher"
	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/selfservice/strategy/oidc"
	"github.com/ory/kratos/x"
)

func TestGenerateState(t *testing.T) {
	conf, reg := internal.NewFastRegistryWithMocks(t)
	_ = conf
	strat := oidc.NewStrategy(reg)
	ctx := context.Background()
	ciph := reg.Cipher(ctx)
	_, ok := ciph.(*cipher.Noop)
	require.False(t, ok)

	var expectProvider string
	assertions := func(t *testing.T) {
		flowID := x.NewUUID()

		stateParam, pkce, err := strat.GenerateState(ctx, &testProvider{}, flowID)
		require.NoError(t, err)
		require.NotEmpty(t, stateParam)
		assert.Empty(t, pkce)

		state, err := oidc.ParseStateCompatiblity(ctx, ciph, stateParam)
		require.NoError(t, err)
		assert.Equal(t, flowID.Bytes(), state.FlowId)
		assert.Empty(t, oidc.PKCEVerifier(state))
		assert.Equal(t, expectProvider, state.ProviderId)
	}

	t.Run("case=old-style", func(t *testing.T) {
		expectProvider = ""
		assertions(t)
	})
	t.Run("case=new-style", func(t *testing.T) {
		oidc.TestHookEnableNewStyleState(t)
		expectProvider = "test-provider"
		assertions(t)
	})
}

type testProvider struct{}

func (t *testProvider) Config() *oidc.Configuration {
	return &oidc.Configuration{ID: "test-provider", PKCE: "never"}
}
