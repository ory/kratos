// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package oidc_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/cipher"
	oidcv1 "github.com/ory/kratos/gen/oidc/v1"
	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/selfservice/flow/registration"
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

	flow := &registration.Flow{
		ID: x.NewUUID(),
	}

	stateParam, pkce, err := strat.GenerateState(ctx, &testProvider{}, flow)
	require.NoError(t, err)
	require.NotEmpty(t, stateParam)
	assert.Empty(t, pkce)

	state, err := oidc.DecryptState(ctx, ciph, stateParam)
	require.NoError(t, err)
	assert.Equal(t, flow.GetID().Bytes(), state.FlowId)
	assert.Empty(t, oidc.PKCEVerifier(state))
	assert.Equal(t, "test-provider", state.ProviderId)
	assert.Equal(t, oidcv1.FlowKind_FLOW_KIND_REGISTRATION, state.FlowKind)
}

type testProvider struct{}

func (t *testProvider) Config() *oidc.Configuration {
	return &oidc.Configuration{ID: "test-provider", PKCE: "never"}
}
