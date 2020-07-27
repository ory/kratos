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

	if len(viper.GetString(configuration.ViperKeySelfServiceLoginUI)) == 0 {
		viper.Set(configuration.ViperKeySelfServiceLoginUI, "http://NewKratosServerWithCSRF/you-forgot-to-set-me/login")
	}
	viper.Set(configuration.ViperKeyPublicBaseURL, public.URL)
	viper.Set(configuration.ViperKeyAdminBaseURL, admin.URL)

	panic("asdf")
	reg.RegisterRoutes(rp, ra)

	t.Cleanup(public.Close)
	t.Cleanup(admin.Close)
	return
}

func NewKratosServerWithRouters(t *testing.T, reg driver.Registry, rp *x.RouterPublic, ra *x.RouterAdmin) (public, admin *httptest.Server) {
	public = httptest.NewServer(rp)
	admin = httptest.NewServer(ra)

	InitKratosServers(t, reg, public, admin)

	t.Cleanup(public.Close)
	t.Cleanup(admin.Close)
	return
}

func InitKratosServers(t *testing.T, reg driver.Registry, public, admin *httptest.Server) {
	if len(viper.GetString(configuration.ViperKeySelfServiceLoginUI)) == 0 {
		viper.Set(configuration.ViperKeySelfServiceLoginUI, "http://NewKratosServerWithRouters/you-forgot-to-set-me/login")
	}
	viper.Set(configuration.ViperKeyPublicBaseURL, public.URL)
	viper.Set(configuration.ViperKeyAdminBaseURL, admin.URL)

	reg.RegisterRoutes(public.Config.Handler.(*x.RouterPublic), admin.Config.Handler.(*x.RouterAdmin))
}

func NewKratosServers(t *testing.T) (public, admin *httptest.Server) {
	public = httptest.NewServer(x.NewRouterPublic())
	admin = httptest.NewServer(x.NewRouterAdmin())

	t.Cleanup(public.Close)
	t.Cleanup(admin.Close)
	return
}
