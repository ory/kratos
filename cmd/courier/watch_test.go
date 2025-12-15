// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package courier

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/phayes/freeport"
	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/internal"
	"github.com/ory/x/configx"
)

func TestStartCourier(t *testing.T) {
	t.Run("case=without metrics", func(t *testing.T) {
		_, r := internal.NewFastRegistryWithMocks(t)
		go func() { _ = StartCourier(t.Context(), r) }()
		time.Sleep(time.Second)
		require.Equal(t, r.Config().CourierExposeMetricsPort(t.Context()), 0)
	})

	t.Run("case=with metrics", func(t *testing.T) {
		port, err := freeport.GetFreePort()
		require.NoError(t, err)
		_, r := internal.NewFastRegistryWithMocks(t, configx.WithValue("expose-metrics-port", port))
		go func() { _ = StartCourier(t.Context(), r) }()
		time.Sleep(time.Second)
		res, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/metrics/prometheus", port))
		require.NoError(t, err)
		require.Equal(t, 200, res.StatusCode)
	})
}
