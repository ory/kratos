package testhelpers

import (
	"testing"

	"github.com/gobuffalo/httptest"

	"github.com/ory/viper"

	"github.com/ory/kratos/driver"
	"github.com/ory/kratos/driver/configuration"
	"github.com/ory/kratos/x"
)

func NewKratosServer(t *testing.T, reg driver.Registry) (public, admin *httptest.Server) {
	rp := x.NewRouterPublic()
	ra := x.NewRouterAdmin()

	public = httptest.NewServer(rp)
	admin = httptest.NewServer(ra)

	viper.Set(configuration.ViperKeyURLsSelfPublic, public.URL)
	viper.Set(configuration.ViperKeyURLsSelfAdmin, admin.URL)

	reg.RegisterRoutes(rp, ra)

	t.Cleanup(public.Close)
	t.Cleanup(admin.Close)

	return
}
