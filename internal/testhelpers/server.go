// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package testhelpers

import (
	"context"
	"strings"
	"testing"

	"github.com/gobuffalo/httptest"
	"github.com/urfave/negroni"

	"github.com/ory/kratos/driver"
	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/x"
	"github.com/ory/kratos/x/nosurfx"
)

func NewKratosServer(t *testing.T, reg driver.Registry) (public, admin *httptest.Server) {
	return NewKratosServerWithRouters(t, reg, x.NewRouterPublic(reg), x.NewRouterAdmin(reg))
}

func NewKratosServerWithCSRF(t *testing.T, reg driver.Registry) (public, admin *httptest.Server) {
	public, admin, _, _ = NewKratosServerWithCSRFAndRouters(t, reg)
	return
}

func NewKratosServerWithCSRFAndRouters(t *testing.T, reg driver.Registry) (public, admin *httptest.Server, rp *x.RouterPublic, ra *x.RouterAdmin) {
	rp, ra = x.NewRouterPublic(reg), x.NewRouterAdmin(reg)
	csrfHandler := nosurfx.NewTestCSRFHandler(rp, reg)
	reg.WithCSRFHandler(csrfHandler)

	ran := negroni.New()
	ran.UseFunc(x.RedirectAdminMiddleware)
	ran.UseHandler(ra)

	rpn := negroni.New()
	rpn.UseFunc(x.HTTPLoaderContextMiddleware(reg))
	rpn.UseHandler(rp)

	public = httptest.NewServer(nosurfx.NewTestCSRFHandler(rpn, reg))
	admin = httptest.NewServer(ran)
	ctx := t.Context()

	// Workaround for:
	// - https://github.com/golang/go/issues/12610
	// - https://github.com/golang/go/issues/31054
	public.URL = strings.ReplaceAll(public.URL, "127.0.0.1", "localhost")

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

	InitKratosServers(t, reg, public, admin, rp, ra)

	t.Cleanup(public.Close)
	t.Cleanup(admin.Close)
	return
}

func InitKratosServers(t *testing.T, reg driver.Registry, public, admin *httptest.Server, rp *x.RouterPublic, ra *x.RouterAdmin) {
	ctx := t.Context()
	if len(reg.Config().GetProvider(ctx).String(config.ViperKeySelfServiceLoginUI)) == 0 {
		reg.Config().MustSet(ctx, config.ViperKeySelfServiceLoginUI, "http://NewKratosServerWithRouters/you-forgot-to-set-me/login")
	}
	reg.Config().MustSet(ctx, config.ViperKeyPublicBaseURL, public.URL)
	reg.Config().MustSet(ctx, config.ViperKeyAdminBaseURL, admin.URL)

	reg.RegisterPublicRoutes(ctx, rp)
	reg.RegisterAdminRoutes(ctx, ra)
}
