// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package config_test

import (
	"context"
	"io"
	"net/http/httptest"
	"testing"

	"github.com/julienschmidt/httprouter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/internal"
)

func TestNewConfigHashHandler(t *testing.T) {
	ctx := context.Background()
	conf, reg := internal.NewFastRegistryWithMocks(t)
	router := httprouter.New()
	config.NewConfigHashHandler(reg, router)
	ts := httptest.NewServer(router)
	t.Cleanup(ts.Close)
	res, err := ts.Client().Get(ts.URL + "/health/config")
	require.NoError(t, err)
	defer res.Body.Close()
	require.Equal(t, 200, res.StatusCode)
	first, err := io.ReadAll(res.Body)
	require.NoError(t, err)

	res, err = ts.Client().Get(ts.URL + "/health/config")
	require.NoError(t, err)
	defer res.Body.Close()
	require.Equal(t, 200, res.StatusCode)
	second, err := io.ReadAll(res.Body)
	require.NoError(t, err)
	assert.Equal(t, first, second)

	require.NoError(t, conf.Set(ctx, config.ViperKeySessionDomain, "foobar"))

	res, err = ts.Client().Get(ts.URL + "/health/config")
	require.NoError(t, err)
	defer res.Body.Close()
	require.Equal(t, 200, res.StatusCode)
	second, err = io.ReadAll(res.Body)
	require.NoError(t, err)
	assert.NotEqual(t, first, second)
}
