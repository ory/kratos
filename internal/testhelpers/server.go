package testhelpers

import (
	"context"
	"strings"
	"testing"

	"github.com/gobuffalo/httptest"

	"github.com/ory/kratos/driver"
	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/x"
)

func NewKratosServer(t *testing.T, reg driver.Registry) (public, admin *httptest.Server) {
	return NewKratosServerWithRouters(t, reg, x.NewRouterPublic(), x.NewRouterAdmin())
}

func NewKratosServerWithCSRF(t *testing.T, reg driver.Registry) (public, admin *httptest.Server) {
	public, admin, _, _ = NewKratosServerWithCSRFAndRouters(t, reg)
	return
}

func NewKratosServerWithCSRFAndRouters(t *testing.T, reg driver.Registry) (public, admin *httptest.Server, rp *x.RouterPublic, ra *x.RouterAdmin) {
	rp, ra = x.NewRouterPublic(), x.NewRouterAdmin()
	csrfHandler := x.NewTestCSRFHandler(rp, reg)
	reg.WithCSRFHandler(csrfHandler)
	public = httptest.NewServer(x.NewTestCSRFHandler(rp, reg))
	admin = httptest.NewServer(ra)

	// Workaround for:
	// - https://github.com/golang/go/issues/12610
	// - https://github.com/golang/go/issues/31054
	public.URL = strings.Replace(public.URL, "127.0.0.1", "localhost", -1)

	if len(reg.Config(context.Background()).Source().String(config.ViperKeySelfServiceLoginUI)) == 0 {
		reg.Config(context.Background()).MustSet(config.ViperKeySelfServiceLoginUI, "http://NewKratosServerWithCSRF/you-forgot-to-set-me/login")
	}
	reg.Config(context.Background()).MustSet(config.ViperKeyPublicBaseURL, public.URL)
	reg.Config(context.Background()).MustSet(config.ViperKeyAdminBaseURL, admin.URL)

	reg.RegisterRoutes(context.Background(), rp, ra)

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
	if len(reg.Config(context.Background()).Source().String(config.ViperKeySelfServiceLoginUI)) == 0 {
		reg.Config(context.Background()).MustSet(config.ViperKeySelfServiceLoginUI, "http://NewKratosServerWithRouters/you-forgot-to-set-me/login")
	}
	reg.Config(context.Background()).MustSet(config.ViperKeyPublicBaseURL, public.URL)
	reg.Config(context.Background()).MustSet(config.ViperKeyAdminBaseURL, admin.URL)

	reg.RegisterRoutes(context.Background(), public.Config.Handler.(*x.RouterPublic), admin.Config.Handler.(*x.RouterAdmin))
}

func NewKratosServers(t *testing.T) (public, admin *httptest.Server) {
	public = httptest.NewServer(x.NewRouterPublic())
	admin = httptest.NewServer(x.NewRouterAdmin())

	t.Cleanup(public.Close)
	t.Cleanup(admin.Close)
	return
}
