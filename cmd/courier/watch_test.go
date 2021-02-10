package courier

import (
	"context"
	"net/http"
	"testing"

	"github.com/ory/kratos/driver"
	"github.com/ory/kratos/driver/config"
	"github.com/ory/x/configx"
	"github.com/ory/x/logrusx"
	"github.com/stretchr/testify/require"
)

func TestStartCourier(t *testing.T) {
	c := config.MustNew(logrusx.New("", ""),
		configx.WithConfigFiles("../../internal/.kratos.yaml"))
	t.Run("case=without metrics", func(t *testing.T) {
		r := driver.NewRegistryDefault().WithConfig(c)
		StartCourier(context.Background(), r)
		_, err := http.Get(c.MetricsListenOn())
		require.Error(t, err)
	})

}
