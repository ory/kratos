package courier

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/ory/kratos/internal"
	"github.com/phayes/freeport"
	"github.com/stretchr/testify/require"
)

func TestStartCourier(t *testing.T) {
	t.Run("case=without metrics", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		_, r := internal.NewFastRegistryWithMocks(t)
		go StartCourier(ctx, r)
		time.Sleep(time.Second)
		require.Equal(t, r.Config(ctx).CourierExposeMetricsPort(), 0)
		cancel()

	})
	t.Run("case=with metrics", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		_, r := internal.NewFastRegistryWithMocks(t)
		port, err := freeport.GetFreePort()
		require.NoError(t, err)
		r.Config(ctx).Set("expose-metrics-port", port)
		go StartCourier(ctx, r)
		time.Sleep(time.Second)
		res, err := http.Get("http://" + r.Config(ctx).MetricsListenOn() + "/metrics/prometheus")
		require.NoError(t, err)
		require.Equal(t, 200, res.StatusCode)
		cancel()
	})

}
