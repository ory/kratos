package daemon

import (
	"crypto/tls"
	"net/http"
	"sync"

	"github.com/ory/kratos/selfservice/flow/recovery"

	"github.com/ory/x/reqlog"

	"github.com/ory/kratos/cmd/courier"
	"github.com/ory/kratos/driver/config"

	"github.com/rs/cors"

	prometheus "github.com/ory/x/prometheusx"

	"github.com/ory/analytics-go/v4"

	"github.com/ory/x/healthx"
	"github.com/ory/x/networkx"

	stdctx "context"

	"github.com/spf13/cobra"
	"github.com/urfave/negroni"

	"github.com/ory/graceful"
	"github.com/ory/x/metricsx"

	"github.com/ory/kratos/driver"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/selfservice/errorx"
	"github.com/ory/kratos/selfservice/flow/login"
	"github.com/ory/kratos/selfservice/flow/logout"
	"github.com/ory/kratos/selfservice/flow/registration"
	"github.com/ory/kratos/selfservice/flow/settings"
	"github.com/ory/kratos/selfservice/flow/verification"
	"github.com/ory/kratos/selfservice/strategy/link"
	"github.com/ory/kratos/selfservice/strategy/oidc"
	"github.com/ory/kratos/session"
	"github.com/ory/kratos/x"
)

type options struct {
	mwf []func(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc)
	ctx stdctx.Context
}

func newOptions(ctx stdctx.Context, opts []Option) *options {
	o := new(options)
	o.ctx = ctx
	for _, f := range opts {
		f(o)
	}
	return o
}

type Option func(*options)

func WithRootMiddleware(m func(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc)) Option {
	return func(o *options) {
		o.mwf = append(o.mwf, m)
	}
}

func WithContext(ctx stdctx.Context) Option {
	return func(o *options) {
		o.ctx = ctx
	}
}

func ServePublic(r driver.Registry, wg *sync.WaitGroup, cmd *cobra.Command, args []string, opts ...Option) {
	defer wg.Done()
	modifiers := newOptions(cmd.Context(), opts)
	ctx := modifiers.ctx

	c := r.Config(cmd.Context())
	l := r.Logger()
	n := negroni.New()
	for _, mw := range modifiers.mwf {
		n.UseFunc(mw)
	}
	publicLogger := reqlog.NewMiddlewareFromLogger(
		l,
		"public#"+c.SelfPublicURL().String(),
	)
	if r.Config(ctx).DisablePublicHealthRequestLog() {
		publicLogger.ExcludePaths(healthx.AliveCheckPath, healthx.ReadyCheckPath)
	}
	n.Use(publicLogger)
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

	if tracer := r.Tracer(ctx); tracer.IsLoaded() {
		n.Use(tracer)
	}

	var handler http.Handler = n
	options, enabled := r.Config(ctx).CORS("public")
	if enabled {
		handler = cors.New(options).Handler(handler)
	}

	certs := c.GetTSLCertificatesForPublic()
	server := graceful.WithDefaults(&http.Server{
		Handler:   handler,
		TLSConfig: &tls.Config{Certificates: certs, MinVersion: tls.VersionTLS12},
	})
	addr := c.PublicListenOn()

	l.Printf("Starting the public httpd on: %s", addr)
	if err := graceful.Graceful(func() error {
		listener, err := networkx.MakeListener(addr, c.PublicSocketPermission())
		if err != nil {
			return err
		}

		if certs == nil {
			return server.Serve(listener)
		}
		return server.ServeTLS(listener, "", "")
	}, server.Shutdown); err != nil {
		l.Fatalf("Failed to gracefully shutdown public httpd: %s", err)
	}
	l.Println("Public httpd was shutdown gracefully")
}

func ServeAdmin(r driver.Registry, wg *sync.WaitGroup, cmd *cobra.Command, args []string, opts ...Option) {
	defer wg.Done()
	modifiers := newOptions(cmd.Context(), opts)
	ctx := modifiers.ctx

	c := r.Config(ctx)
	l := r.Logger()
	n := negroni.New()
	for _, mw := range modifiers.mwf {
		n.UseFunc(mw)
	}
	adminLogger := reqlog.NewMiddlewareFromLogger(
		l,
		"admin#"+c.SelfPublicURL().String(),
	)
	if r.Config(ctx).DisableAdminHealthRequestLog() {
		adminLogger.ExcludePaths(healthx.AliveCheckPath, healthx.ReadyCheckPath)
	}
	n.Use(adminLogger)
	n.Use(sqa(ctx, cmd, r))
	n.Use(r.PrometheusManager())

	router := x.NewRouterAdmin()
	r.RegisterAdminRoutes(ctx, router)
	r.PrometheusManager().RegisterRouter(router.Router)

	if tracer := r.Tracer(ctx); tracer.IsLoaded() {
		n.Use(tracer)
	}

	n.UseHandler(router)
	certs := c.GetTSLCertificatesForAdmin()
	server := graceful.WithDefaults(&http.Server{
		Handler:   n,
		TLSConfig: &tls.Config{Certificates: certs, MinVersion: tls.VersionTLS12},
	})
	addr := c.AdminListenOn()

	l.Printf("Starting the admin httpd on: %s", addr)
	if err := graceful.Graceful(func() error {
		listener, err := networkx.MakeListener(addr, c.AdminSocketPermission())
		if err != nil {
			return err
		}

		if certs == nil {
			return server.Serve(listener)
		}
		return server.ServeTLS(listener, "", "")
	}, server.Shutdown); err != nil {
		l.Fatalf("Failed to gracefully shutdown admin httpd: %s", err)
	}
	l.Println("Admin httpd was shutdown gracefully")
}

func sqa(ctx stdctx.Context, cmd *cobra.Command, d driver.Registry) *metricsx.Service {
	// Creates only ones
	// instance
	return metricsx.New(
		cmd,
		d.Logger(),
		d.Config(ctx).Source(),
		&metricsx.Options{
			Service:       "ory-kratos",
			ClusterID:     metricsx.Hash(d.Persister().NetworkID().String()),
			IsDevelopment: d.Config(ctx).IsInsecureDevMode(),
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
				identity.RouteCollection,

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

func bgTasks(d driver.Registry, wg *sync.WaitGroup, cmd *cobra.Command, args []string, opts ...Option) {
	defer wg.Done()
	modifiers := newOptions(cmd.Context(), opts)
	ctx := modifiers.ctx

	if d.Config(ctx).IsBackgroundCourierEnabled() {
		go courier.Watch(ctx, d)
	}
}

func ServeAll(d driver.Registry, opts ...Option) func(cmd *cobra.Command, args []string) {
	return func(cmd *cobra.Command, args []string) {
		var wg sync.WaitGroup
		wg.Add(3)
		go ServePublic(d, &wg, cmd, args, opts...)
		go ServeAdmin(d, &wg, cmd, args, opts...)
		go bgTasks(d, &wg, cmd, args, opts...)
		wg.Wait()
	}
}
