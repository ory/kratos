// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package oidc_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/cipher"
	oidcv1 "github.com/ory/kratos/gen/oidc/v1"
	"github.com/ory/kratos/pkg"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/selfservice/flow/registration"
	"github.com/ory/kratos/selfservice/strategy/oidc"
	"github.com/ory/kratos/x"
)

func TestGenerateState(t *testing.T) {
	_, reg := pkg.NewFastRegistryWithMocks(t)
	strat := oidc.NewStrategy(reg)
	ciph := reg.Cipher(t.Context())
	_, ok := ciph.(*cipher.Noop)
	require.False(t, ok)

	flow := &registration.Flow{
		ID: x.NewUUID(),
	}

	stateParam, pkce, err := strat.GenerateState(t.Context(), &testProvider{}, flow, "https://teststatehost")
	require.NoError(t, err)
	require.NotEmpty(t, stateParam)
	assert.Empty(t, pkce)

	state, err := oidc.DecryptState(t.Context(), ciph, stateParam)
	require.NoError(t, err)
	assert.Equal(t, flow.GetID().Bytes(), state.FlowId)
	assert.Empty(t, oidc.PKCEVerifier(state))
	assert.Equal(t, "test-provider", state.ProviderId)
	assert.Equal(t, oidcv1.FlowKind_FLOW_KIND_REGISTRATION, state.FlowKind)
	assert.Equal(t, "https://teststatehost", state.RequestBaseUrl)
}

// TestGenerateStateUsesFlowCapturedBaseURL is the regression test for the
// init→submit bridge. The Ory CLI / Tunnel only sends the
// Ory-Base-URL-Rewrite header on the flow-*init* request; the provider
// submit that reaches GenerateState does not carry it. The base URL is
// therefore captured onto the flow's InternalContext at init, and
// GenerateState must prefer that captured value over the request-derived
// fallback — otherwise the social-sign-in callback never gets redirected
// back onto the developer host (the bug this guards against).
func TestGenerateStateUsesFlowCapturedBaseURL(t *testing.T) {
	_, reg := pkg.NewFastRegistryWithMocks(t)
	strat := oidc.NewStrategy(reg)
	ciph := reg.Cipher(t.Context())

	t.Run("flow-captured base URL wins over the submit fallback", func(t *testing.T) {
		f := &registration.Flow{ID: x.NewUUID(), InternalContext: []byte("{}")}
		require.NoError(t, flow.SetRequestBaseURL(f, "http://localhost:4000"))

		// The fallback (what the submit request resolved to, e.g. the bare
		// projects.oryapis host) must be ignored in favor of the captured value.
		stateParam, _, err := strat.GenerateState(t.Context(), &testProvider{}, f, "https://slug.projects.oryapis.com")
		require.NoError(t, err)
		state, err := oidc.DecryptState(t.Context(), ciph, stateParam)
		require.NoError(t, err)
		assert.Equal(t, "http://localhost:4000", state.RequestBaseUrl,
			"http scheme + host captured at init must survive to the callback")
	})

	t.Run("no captured base URL falls back to the passed value", func(t *testing.T) {
		f := &registration.Flow{ID: x.NewUUID(), InternalContext: []byte("{}")}
		stateParam, _, err := strat.GenerateState(t.Context(), &testProvider{}, f, "https://slug.projects.oryapis.com")
		require.NoError(t, err)
		state, err := oidc.DecryptState(t.Context(), ciph, stateParam)
		require.NoError(t, err)
		assert.Equal(t, "https://slug.projects.oryapis.com", state.RequestBaseUrl)
	})
}

type testProvider struct{}

func (t *testProvider) Config() *oidc.Configuration {
	return &oidc.Configuration{ID: "test-provider", PKCE: "never"}
}
