// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package daemon

import (
	"context"
	"crypto/tls"
	"net/http"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/cors"
	"github.com/spf13/cobra"
	"github.com/urfave/negroni"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"golang.org/x/sync/errgroup"

	"github.com/ory/analytics-go/v5"
	"github.com/ory/graceful"
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
	"github.com/ory/kratos/x/nosurfx"
	"github.com/ory/x/healthx"
	"github.com/ory/x/metricsx"
	"github.com/ory/x/networkx"
	"github.com/ory/x/otelx"
	"github.com/ory/x/otelx/semconv"
	prometheus "github.com/ory/x/prometheusx"
	"github.com/ory/x/reqlog"
	"github.com/ory/x/urlx"
)

func init() {
	graceful.DefaultShutdownTimeout = 120 * time.Second
}

func servePublic(ctx context.Context, r *driver.RegistryDefault, cmd *cobra.Command) (func() error, error) {
	cfg := r.Config().ServePublic(ctx)
	l := r.Logger()
	n := negroni.New()

	for _, mw := range r.HTTPMiddlewares() {
		n.Use(mw)
	}

	publicLogger := reqlog.NewMiddlewareFromLogger(l, "public#"+cfg.BaseURL.String())

	if cfg.RequestLog.DisableHealth {
		publicLogger.ExcludePaths(healthx.AliveCheckPath, healthx.ReadyCheckPath)
	}

	n.UseFunc(semconv.Middleware)
	n.Use(publicLogger)
	n.Use(x.HTTPLoaderContextMiddleware(r))
	n.Use(sqa(ctx, cmd, r))

	router := x.NewRouterPublic(r)
	csrf := nosurfx.NewCSRFHandler(router, r)

	// we need to always load the CORS middleware even if it is disabled, to allow hot-enabling CORS
	n.UseFunc(func(w http.ResponseWriter, req *http.Request, next http.HandlerFunc) {
		cfg, enabled := r.Config().CORSPublic(req.Context())
		if !enabled {
			next(w, req)
			return
		}
		cors.New(cfg).ServeHTTP(w, req, next)
	})

	n.UseFunc(x.CleanPath) // Prevent double slashes from breaking CSRF.
	r.WithCSRFHandler(csrf)
	n.UseHandler(http.MaxBytesHandler(r.CSRFHandler(), 5*1024*1024 /* 5 MB */))

	// Disable CSRF for these endpoints
	csrf.DisablePath(healthx.AliveCheckPath)
	csrf.DisablePath(healthx.ReadyCheckPath)
	csrf.DisablePath(healthx.VersionPath)
	csrf.DisablePath(prometheus.MetricsPrometheusPath)

	r.RegisterPublicRoutes(ctx, router)

	var handler http.Handler = n
	if tracer := r.Tracer(ctx); tracer.IsLoaded() {
		handler = otelx.TraceHandler(handler, otelhttp.WithTracerProvider(tracer.Provider()))
	}

	certFunc, err := cfg.TLS.GetCertFunc(ctx, l, "public")
	if err != nil {
		return nil, err
	}

	//#nosec G112 -- the correct settings are set by graceful.WithDefaults
	server := graceful.WithDefaults(&http.Server{
		Handler:           handler,
		TLSConfig:         &tls.Config{GetCertificate: certFunc, MinVersion: tls.VersionTLS12},
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       120 * time.Second,
	})
	addr := cfg.GetAddress()

	return func() error {
		l.Printf("Starting the public httpd on: %s", addr)
		if err := graceful.GracefulContext(ctx, func() error {
			listener, err := networkx.MakeListener(addr, &cfg.Socket)
			if err != nil {
				return err
			}

			if certFunc == nil {
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
	}, nil
}

func serveAdmin(ctx context.Context, r *driver.RegistryDefault, cmd *cobra.Command) (func() error, error) {
	cfg := r.Config().ServeAdmin(ctx)
	l := r.Logger()
	n := negroni.New()

	for _, mw := range r.HTTPMiddlewares() {
		n.Use(mw)
	}

	adminLogger := reqlog.NewMiddlewareFromLogger(l, "admin#"+cfg.BaseURL.String())

	if cfg.RequestLog.DisableHealth {
		adminLogger.ExcludePaths(x.AdminPrefix+healthx.AliveCheckPath, x.AdminPrefix+healthx.ReadyCheckPath, x.AdminPrefix+prometheus.MetricsPrometheusPath)
	}
	n.UseFunc(semconv.Middleware)
	n.Use(adminLogger)
	n.UseFunc(x.RedirectAdminMiddleware)
	n.Use(x.HTTPLoaderContextMiddleware(r))
	n.Use(sqa(ctx, cmd, r))

	router := x.NewRouterAdmin(r)
	r.RegisterAdminRoutes(ctx, router)

	n.UseHandler(http.MaxBytesHandler(router, 5*1024*1024 /* 5 MB */))

	var handler http.Handler = n
	if tracer := r.Tracer(ctx); tracer.IsLoaded() {
		handler = otelx.TraceHandler(handler,
			otelhttp.WithTracerProvider(tracer.Provider()),
			otelhttp.WithFilter(func(req *http.Request) bool {
				return req.URL.Path != x.AdminPrefix+prometheus.MetricsPrometheusPath
			}),
		)
	}

	certFunc, err := cfg.TLS.GetCertFunc(ctx, l, "admin")
	if err != nil {
		return nil, err
	}

	//#nosec G112 -- the correct settings are set by graceful.WithDefaults
	server := graceful.WithDefaults(&http.Server{
		Handler:           handler,
		TLSConfig:         &tls.Config{GetCertificate: certFunc, MinVersion: tls.VersionTLS12},
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      120 * time.Second,
		IdleTimeout:       600 * time.Second,
	})

	addr := cfg.GetAddress()

	return func() error {
		l.Printf("Starting the admin httpd on: %s", addr)
		if err := graceful.GracefulContext(ctx, func() error {
			listener, err := networkx.MakeListener(addr, &cfg.Socket)
			if err != nil {
				return err
			}

			if certFunc == nil {
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
	}, nil
}

func sqa(ctx context.Context, cmd *cobra.Command, d driver.Registry) *metricsx.Service {
	urls := []string{
		d.Config().ServePublic(ctx).BaseURL.Host,
		d.Config().ServeAdmin(ctx).BaseURL.Host,
		d.Config().SelfServiceFlowLoginUI(ctx).Host,
		d.Config().SelfServiceFlowSettingsUI(ctx).Host,
		d.Config().SelfServiceFlowErrorURL(ctx).Host,
		d.Config().SelfServiceFlowRegistrationUI(ctx).Host,
		d.Config().SelfServiceFlowRecoveryUI(ctx).Host,
		d.Config().ServePublic(ctx).Host,
		d.Config().ServeAdmin(ctx).Host,
	}
	if c, y := d.Config().CORSPublic(ctx); y {
		urls = append(urls, c.AllowedOrigins...)
	}
	host := urlx.ExtractPublicAddress(urls...)

	// Creates only ones
	// instance
	return metricsx.New(
		cmd,
		d.Logger(),
		d.Config().GetProvider(ctx),
		&metricsx.Options{
			Service:       "kratos",
			DeploymentId:  metricsx.Hash(d.Persister().NetworkID(ctx).String()),
			DBDialect:     d.Persister().GetConnection(ctx).Dialect.Details().Dialect,
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
				Endpoint:             "https://sqa.ory.sh",
				GzipCompressionLevel: 6,
				BatchMaxSize:         500 * 1000,
				BatchSize:            1000,
				Interval:             time.Hour * 6,
			},
			Hostname: host,
		},
	)
}

func courierTask(ctx context.Context, d driver.Registry) func() error {
	return func() error {
		if d.Config().IsBackgroundCourierEnabled(ctx) {
			return courier.Watch(ctx, d)
		}
		return nil
	}
}

func ServeAll(d *driver.RegistryDefault) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ctx := cmd.Context()
		g, ctx := errgroup.WithContext(ctx)
		cmd.SetContext(ctx)

		// construct all tasks upfront to avoid race conditions
		publicSrv, err := servePublic(ctx, d, cmd)
		if err != nil {
			return errors.WithStack(err)
		}
		adminSrv, err := serveAdmin(ctx, d, cmd)
		if err != nil {
			return errors.WithStack(err)
		}
		tasks := []func() error{
			publicSrv,
			adminSrv,
			courierTask(ctx, d),
		}
		for _, task := range tasks {
			g.Go(task)
		}
		return g.Wait()
	}
}
