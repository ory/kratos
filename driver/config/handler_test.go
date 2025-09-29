// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package config_test

import (
	"context"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/x"
	"github.com/ory/x/contextx"
)

type configProvider struct {
	cfg *config.Config
}

func (c *configProvider) Config() *config.Config {
	return c.cfg
}

func TestNewConfigHashHandler(t *testing.T) {
	ctx := context.Background()
	cfg := internal.NewConfigurationWithDefaults(t)
	router := x.NewTestRouterPublic(t)
	config.NewConfigHashHandler(&configProvider{cfg: cfg}, router)
	ts := contextx.NewConfigurableTestServer(router)
	t.Cleanup(ts.Close)

	// first request, get baseline hash
	res, err := ts.Client(ctx).Get(ts.URL + "/health/config")
	require.NoError(t, err)
	defer func() { _ = res.Body.Close() }()
	require.Equal(t, 200, res.StatusCode)
	first, err := io.ReadAll(res.Body)
	require.NoError(t, err)

	// second request, no config change
	res, err = ts.Client(ctx).Get(ts.URL + "/health/config")
	require.NoError(t, err)
	defer func() { _ = res.Body.Close() }()
	require.Equal(t, 200, res.StatusCode)
	second, err := io.ReadAll(res.Body)
	require.NoError(t, err)
	assert.Equal(t, first, second)

	// third request, with config change
	res, err = ts.Client(contextx.WithConfigValue(ctx, config.ViperKeySessionDomain, "foobar")).Get(ts.URL + "/health/config")
	require.NoError(t, err)
	defer func() { _ = res.Body.Close() }()
	require.Equal(t, 200, res.StatusCode)
	third, err := io.ReadAll(res.Body)
	require.NoError(t, err)
	assert.NotEqual(t, first, third)

	// fourth request, no config change
	res, err = ts.Client(ctx).Get(ts.URL + "/health/config")
	require.NoError(t, err)
	defer func() { _ = res.Body.Close() }()
	require.Equal(t, 200, res.StatusCode)
	fourth, err := io.ReadAll(res.Body)
	require.NoError(t, err)
	assert.Equal(t, first, fourth)
}
