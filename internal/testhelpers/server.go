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
	return NewKratosServerWithRouters(t, reg, x.NewRouterPublic(), x.NewRouterAdmin())

}

func NewKratosServerWithCSRF(t *testing.T, reg driver.Registry) (public, admin *httptest.Server) {
	rp, ra := x.NewRouterPublic(), x.NewRouterAdmin()
	public = httptest.NewServer(x.NewTestCSRFHandler(rp, reg))
	admin = httptest.NewServer(ra)

	viper.Set(configuration.ViperKeyURLsLogin, "http://NewKratosServerWithCSRF/i-am-a-mock-value")
	viper.Set(configuration.ViperKeyURLsSelfPublic, public.URL)
	viper.Set(configuration.ViperKeyURLsSelfAdmin, admin.URL)

	reg.RegisterRoutes(rp, ra)

	t.Cleanup(public.Close)
	t.Cleanup(admin.Close)
	return
}

func NewKratosServerWithRouters(t *testing.T, reg driver.Registry, rp *x.RouterPublic, ra *x.RouterAdmin) (public, admin *httptest.Server) {
	public = httptest.NewServer(rp)
	admin = httptest.NewServer(ra)

	viper.Set(configuration.ViperKeyURLsLogin, "http://NewKratosServerWithRouters/i-am-a-mock-value")
	viper.Set(configuration.ViperKeyURLsSelfPublic, public.URL)
	viper.Set(configuration.ViperKeyURLsSelfAdmin, admin.URL)

	reg.RegisterRoutes(rp, ra)

	t.Cleanup(public.Close)
	t.Cleanup(admin.Close)

	return
}
