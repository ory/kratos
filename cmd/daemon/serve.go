// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package daemon

import (
	stdctx "context"
	"crypto/tls"
	"net/http"

	"github.com/pkg/errors"
	"github.com/rs/cors"
	"github.com/spf13/cobra"
	"github.com/urfave/negroni"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"golang.org/x/net/context"
	"golang.org/x/sync/errgroup"

	"github.com/ory/analytics-go/v4"
	"github.com/ory/graceful"
	"github.com/ory/x/healthx"
	"github.com/ory/x/metricsx"
	"github.com/ory/x/networkx"
	"github.com/ory/x/otelx"
	prometheus "github.com/ory/x/prometheusx"
	"github.com/ory/x/reqlog"
	"github.com/ory/x/servicelocatorx"

	"github.com/ory/kratos/cmd/courier"
	"github.com/ory/kratos/driver"
	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/schema"
	"github.com/ory/kratos/selfservice/errorx"
	"github.com/ory/kratos/selfservice/flow/login"
	"github.com/ory/kratos/selfservice/flow/logout"
	"github.com/ory/kratos/selfservice/flow/recovery"
	"github.com/ory/kratos/selfservice/flow/registration"
	"github.com/ory/kratos/selfservice/flow/settings"
	"github.com/ory/kratos/selfservice/flow/verification"
	"github.com/ory/kratos/selfservice/strategy/link"
	"github.com/ory/kratos/selfservice/strategy/oidc"
	"github.com/ory/kratos/session"
	"github.com/ory/kratos/x"
)

type options struct {
	ctx stdctx.Context
}

func NewOptions(ctx stdctx.Context, opts []Option) *options {
	o := new(options)
	o.ctx = ctx
	for _, f := range opts {
		f(o)
	}
	return o
}

type Option func(*options)

func WithContext(ctx stdctx.Context) Option {
	return func(o *options) {
		o.ctx = ctx
	}
}

func ServePublic(r driver.Registry, cmd *cobra.Command, args []string, slOpts *servicelocatorx.Options, opts []Option) error {
	modifiers := NewOptions(cmd.Context(), opts)
	ctx := modifiers.ctx

	c := r.Config()
	l := r.Logger()
	n := negroni.New()

	for _, mw := range slOpts.HTTPMiddlewares() {
		n.UseFunc(mw)
	}

	publicLogger := reqlog.NewMiddlewareFromLogger(
		l,
		"public#"+c.SelfPublicURL(ctx).String(),
	)

	if r.Config().DisablePublicHealthRequestLog(ctx) {
		publicLogger.ExcludePaths(healthx.AliveCheckPath, healthx.ReadyCheckPath)
	}

	n.Use(publicLogger)
	n.Use(x.HTTPLoaderContextMiddleware(r))
	n.Use(sqa(ctx, cmd, r))

	n.Use(r.PrometheusManager())

	router := x.NewRouterPublic()
	csrf := x.NewCSRFHandler(router, r)

	n.UseFunc(x.CleanPath) // Prevent double slashes from breaking CSRF.
	r.WithCSRFHandler(csrf)
	n.UseHandler(r.CSRFHandler())

	// Disable CSRF for these endpoints
	csrf.DisablePath(healthx.AliveCheckPath)
	csrf.DisablePath(healthx.ReadyCheckPath)
	csrf.DisablePath(healthx.VersionPath)
	csrf.DisablePath(prometheus.MetricsPrometheusPath)

	r.RegisterPublicRoutes(ctx, router)
	r.PrometheusManager().RegisterRouter(router.Router)

	var handler http.Handler = n
	options, enabled := r.Config().CORS(ctx, "public")
	if enabled {
		handler = cors.New(options).Handler(handler)
	}

	certs := c.GetTLSCertificatesForPublic(ctx)

	if tracer := r.Tracer(ctx); tracer.IsLoaded() {
		handler = otelx.TraceHandler(handler, otelhttp.WithTracerProvider(tracer.Provider()))
	}

	// #nosec G112 - the correct settings are set by graceful.WithDefaults
	server := graceful.WithDefaults(&http.Server{
		Handler:   handler,
		TLSConfig: &tls.Config{GetCertificate: certs, MinVersion: tls.VersionTLS12},
	})
	addr := c.PublicListenOn(ctx)

	l.Printf("Starting the public httpd on: %s", addr)
	if err := graceful.Graceful(func() error {
		listener, err := networkx.MakeListener(addr, c.PublicSocketPermission(ctx))
		if err != nil {
			return err
		}

		if certs == nil {
			return server.Serve(listener)
		}
		return server.ServeTLS(listener, "", "")
	}, server.Shutdown); err != nil {
		if !errors.Is(err, context.Canceled) {
			l.Errorf("Failed to gracefully shutdown public httpd: %s", err)
			return err
		}
	}
	l.Println("Public httpd was shutdown gracefully")
	return nil
}

func ServeAdmin(r driver.Registry, cmd *cobra.Command, args []string, slOpts *servicelocatorx.Options, opts []Option) error {
	modifiers := NewOptions(cmd.Context(), opts)
	ctx := modifiers.ctx

	c := r.Config()
	l := r.Logger()
	n := negroni.New()

	for _, mw := range slOpts.HTTPMiddlewares() {
		n.UseFunc(mw)
	}

	adminLogger := reqlog.NewMiddlewareFromLogger(
		l,
		"admin#"+c.SelfPublicURL(ctx).String(),
	)

	if r.Config().DisableAdminHealthRequestLog(ctx) {
		adminLogger.ExcludePaths(x.AdminPrefix+healthx.AliveCheckPath, x.AdminPrefix+healthx.ReadyCheckPath, x.AdminPrefix+prometheus.MetricsPrometheusPath)
	}
	n.Use(adminLogger)
	n.UseFunc(x.RedirectAdminMiddleware)
	n.Use(x.HTTPLoaderContextMiddleware(r))
	n.Use(sqa(ctx, cmd, r))
	n.Use(r.PrometheusManager())

	router := x.NewRouterAdmin()
	r.RegisterAdminRoutes(ctx, router)
	r.PrometheusManager().RegisterRouter(router.Router)

	n.UseHandler(router)
	certs := c.GetTLSCertificatesForAdmin(ctx)

	var handler http.Handler = n
	if tracer := r.Tracer(ctx); tracer.IsLoaded() {
		handler = otelx.TraceHandler(handler,
			otelhttp.WithTracerProvider(tracer.Provider()),
			otelhttp.WithFilter(func(req *http.Request) bool {
				return req.URL.Path != x.AdminPrefix+prometheus.MetricsPrometheusPath
			}),
		)
	}

	// #nosec G112 - the correct settings are set by graceful.WithDefaults
	server := graceful.WithDefaults(&http.Server{
		Handler:   handler,
		TLSConfig: &tls.Config{GetCertificate: certs, MinVersion: tls.VersionTLS12},
	})

	addr := c.AdminListenOn(ctx)

	l.Printf("Starting the admin httpd on: %s", addr)
	if err := graceful.Graceful(func() error {
		listener, err := networkx.MakeListener(addr, c.AdminSocketPermission(ctx))
		if err != nil {
			return err
		}

		if certs == nil {
			return server.Serve(listener)
		}
		return server.ServeTLS(listener, "", "")
	}, server.Shutdown); err != nil {
		if !errors.Is(err, context.Canceled) {
			l.Errorf("Failed to gracefully shutdown admin httpd: %s", err)
			return err
		}
	}
	l.Println("Admin httpd was shutdown gracefully")
	return nil
}

func sqa(ctx stdctx.Context, cmd *cobra.Command, d driver.Registry) *metricsx.Service {
	// Creates only ones
	// instance
	return metricsx.New(
		cmd,
		d.Logger(),
		d.Config().GetProvider(ctx),
		&metricsx.Options{
			Service:       "ory-kratos",
			ClusterID:     metricsx.Hash(d.Persister().NetworkID(ctx).String()),
			IsDevelopment: d.Config().IsInsecureDevMode(ctx),
			WriteKey:      "qQlI6q8Q4WvkzTjKQSor4sHYOikHIvvi",
			WhitelistedPaths: []string{
				"/",
				healthx.AliveCheckPath,
				healthx.ReadyCheckPath,
				healthx.VersionPath,

				oidc.RouteBase,

				login.RouteInitBrowserFlow,
				login.RouteInitAPIFlow,
				login.RouteGetFlow,
				login.RouteSubmitFlow,

				logout.RouteInitBrowserFlow,
				logout.RouteSubmitFlow,
				logout.RouteAPIFlow,

				registration.RouteInitBrowserFlow,
				registration.RouteInitAPIFlow,
				registration.RouteGetFlow,
				registration.RouteSubmitFlow,

				session.RouteWhoami,

				x.AdminPrefix + "/" + schema.SchemasPath,
				x.AdminPrefix + identity.RouteCollection,

				settings.RouteInitBrowserFlow,
				settings.RouteInitAPIFlow,
				settings.RouteGetFlow,
				settings.RouteSubmitFlow,

				verification.RouteInitAPIFlow,
				verification.RouteInitBrowserFlow,
				verification.RouteGetFlow,
				verification.RouteSubmitFlow,

				recovery.RouteInitAPIFlow,
				recovery.RouteInitBrowserFlow,
				recovery.RouteGetFlow,
				recovery.RouteSubmitFlow,

				link.RouteAdminCreateRecoveryLink,

				errorx.RouteGet,
				prometheus.MetricsPrometheusPath,
			},
			BuildVersion: config.Version,
			BuildHash:    config.Commit,
			BuildTime:    config.Date,
			Config: &analytics.Config{
				Endpoint: "https://sqa.ory.sh",
			},
		},
	)
}

func bgTasks(d driver.Registry, cmd *cobra.Command, args []string, slOpts *servicelocatorx.Options, opts []Option) error {
	modifiers := NewOptions(cmd.Context(), opts)
	ctx := modifiers.ctx

	if d.Config().IsBackgroundCourierEnabled(ctx) {
		return courier.Watch(ctx, d)
	}

	return nil
}

func ServeAll(d driver.Registry, slOpts *servicelocatorx.Options, opts []Option) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		mods := NewOptions(cmd.Context(), opts)
		ctx := mods.ctx

		g, ctx := errgroup.WithContext(ctx)
		cmd.SetContext(ctx)
		opts = append(opts, WithContext(ctx))

		g.Go(func() error {
			return ServePublic(d, cmd, args, slOpts, opts)
		})
		g.Go(func() error {
			return ServeAdmin(d, cmd, args, slOpts, opts)
		})
		g.Go(func() error {
			return bgTasks(d, cmd, args, slOpts, opts)
		})
		return g.Wait()
	}
}
