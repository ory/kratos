// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package config_test

import (
	"context"
	"io"
	"testing"

	confighelpers "github.com/ory/kratos/driver/config/testhelpers"

	"github.com/julienschmidt/httprouter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/internal"
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
	router := httprouter.New()
	config.NewConfigHashHandler(&configProvider{cfg: cfg}, router)
	ts := confighelpers.NewConfigurableTestServer(router)
	t.Cleanup(ts.Close)

	// first request, get baseline hash
	res, err := ts.Client(ctx).Get(ts.URL + "/health/config")
	require.NoError(t, err)
	defer res.Body.Close()
	require.Equal(t, 200, res.StatusCode)
	first, err := io.ReadAll(res.Body)
	require.NoError(t, err)

	// second request, no config change
	res, err = ts.Client(ctx).Get(ts.URL + "/health/config")
	require.NoError(t, err)
	defer res.Body.Close()
	require.Equal(t, 200, res.StatusCode)
	second, err := io.ReadAll(res.Body)
	require.NoError(t, err)
	assert.Equal(t, first, second)

	// third request, with config change
	res, err = ts.Client(confighelpers.WithConfigValue(ctx, config.ViperKeySessionDomain, "foobar")).Get(ts.URL + "/health/config")
	require.NoError(t, err)
	defer res.Body.Close()
	require.Equal(t, 200, res.StatusCode)
	third, err := io.ReadAll(res.Body)
	require.NoError(t, err)
	assert.NotEqual(t, first, third)

	// fourth request, no config change
	res, err = ts.Client(ctx).Get(ts.URL + "/health/config")
	require.NoError(t, err)
	defer res.Body.Close()
	require.Equal(t, 200, res.StatusCode)
	fourth, err := io.ReadAll(res.Body)
	require.NoError(t, err)
	assert.Equal(t, first, fourth)
}
