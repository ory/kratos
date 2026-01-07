// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package testhelpers

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/urfave/negroni"

	"github.com/ory/kratos/driver"
	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/x"
	"github.com/ory/kratos/x/nosurfx"
	"github.com/ory/x/httprouterx"
)

func NewKratosServer(t *testing.T, reg driver.Registry) (public, admin *httptest.Server) {
	return NewKratosServerWithRouters(t, reg, httprouterx.NewTestRouterPublic(t), httprouterx.NewTestRouterAdminWithPrefix(t))
}

func NewKratosServerWithCSRF(t *testing.T, reg driver.Registry) (public, admin *httptest.Server) {
	public, admin, _, _ = NewKratosServerWithCSRFAndRouters(t, reg)
	return
}

func NewKratosServerWithCSRFAndRouters(t *testing.T, reg driver.Registry) (public, admin *httptest.Server, rp *httprouterx.RouterPublic, ra *httprouterx.RouterAdmin) {
	rp, ra = httprouterx.NewTestRouterPublic(t), httprouterx.NewTestRouterAdminWithPrefix(t)
	csrfHandler := nosurfx.NewTestCSRFHandler(rp, reg)
	reg.WithCSRFHandler(csrfHandler)

	ran := negroni.New()
	ran.UseFunc(httprouterx.AddAdminPrefixIfNotPresentNegroni)
	ran.UseHandler(ra)

	rpn := negroni.New()
	rpn.UseFunc(x.HTTPLoaderContextMiddleware(reg))
	rpn.UseHandler(rp)

	public = httptest.NewServer(nosurfx.NewTestCSRFHandler(rpn, reg))
	admin = httptest.NewServer(ran)

	public.URL = strings.ReplaceAll(public.URL, "127.0.0.1", "localhost")

	initKratosServers(t, reg, public, admin, rp, ra)

	t.Cleanup(public.Close)
	t.Cleanup(admin.Close)
	return
}

func NewKratosServerWithRouters(t *testing.T, reg driver.Registry, rp *httprouterx.RouterPublic, ra *httprouterx.RouterAdmin) (public, admin *httptest.Server) {
	np := negroni.New()
	np.UseFunc(httprouterx.TrimTrailingSlashNegroni)
	np.UseHandler(rp)

	na := negroni.New()
	na.UseFunc(httprouterx.AddAdminPrefixIfNotPresentNegroni)
	na.UseFunc(httprouterx.TrimTrailingSlashNegroni)
	na.UseHandler(ra)

	public = httptest.NewServer(np)
	admin = httptest.NewServer(na)

	initKratosServers(t, reg, public, admin, rp, ra)

	t.Cleanup(public.Close)
	t.Cleanup(admin.Close)
	return
}

func initKratosServers(t *testing.T, reg driver.Registry, public, admin *httptest.Server, rp *httprouterx.RouterPublic, ra *httprouterx.RouterAdmin) {
	ctx := t.Context()
	if len(reg.Config().GetProvider(ctx).String(config.ViperKeySelfServiceLoginUI)) == 0 {
		reg.Config().MustSet(ctx, config.ViperKeySelfServiceLoginUI, "http://NewKratosServerWithRouters/you-forgot-to-set-me/login")
	}
	reg.Config().MustSet(ctx, config.ViperKeyPublicBaseURL, public.URL)
	reg.Config().MustSet(ctx, config.ViperKeyAdminBaseURL, admin.URL)

	reg.RegisterPublicRoutes(ctx, rp)
	reg.RegisterAdminRoutes(ctx, ra)
}
