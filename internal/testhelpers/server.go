// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package testhelpers

import (
	"context"
	"strings"
	"testing"

	"github.com/urfave/negroni"

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
	ran := negroni.New()
	ran.UseFunc(x.RedirectAdminMiddleware)
	ran.UseHandler(ra)
	public = httptest.NewServer(x.NewTestCSRFHandler(rp, reg))
	admin = httptest.NewServer(ran)
	ctx := context.Background()

	// Workaround for:
	// - https://github.com/golang/go/issues/12610
	// - https://github.com/golang/go/issues/31054
	public.URL = strings.Replace(public.URL, "127.0.0.1", "localhost", -1)

	if len(reg.Config().GetProvider(ctx).String(config.ViperKeySelfServiceLoginUI)) == 0 {
		reg.Config().MustSet(ctx, config.ViperKeySelfServiceLoginUI, "http://NewKratosServerWithCSRF/you-forgot-to-set-me/login")
	}
	reg.Config().MustSet(ctx, config.ViperKeyPublicBaseURL, public.URL)
	reg.Config().MustSet(ctx, config.ViperKeyAdminBaseURL, admin.URL)

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
	ctx := context.Background()
	if len(reg.Config().GetProvider(ctx).String(config.ViperKeySelfServiceLoginUI)) == 0 {
		reg.Config().MustSet(ctx, config.ViperKeySelfServiceLoginUI, "http://NewKratosServerWithRouters/you-forgot-to-set-me/login")
	}
	reg.Config().MustSet(ctx, config.ViperKeyPublicBaseURL, public.URL)
	reg.Config().MustSet(ctx, config.ViperKeyAdminBaseURL, admin.URL)

	reg.RegisterRoutes(context.Background(), public.Config.Handler.(*x.RouterPublic), admin.Config.Handler.(*x.RouterAdmin))
}

func NewKratosServers(t *testing.T) (public, admin *httptest.Server) {
	public = httptest.NewServer(x.NewRouterPublic())
	admin = httptest.NewServer(x.NewRouterAdmin())

	public.URL = strings.Replace(public.URL, "127.0.0.1", "localhost", -1)
	t.Cleanup(public.Close)
	t.Cleanup(admin.Close)
	return
}
