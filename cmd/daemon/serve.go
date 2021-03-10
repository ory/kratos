package daemon

import (
	"net/http"
	"strings"
	"sync"

	"github.com/ory/x/reqlog"

	"github.com/ory/kratos/cmd/courier"
	"github.com/ory/kratos/driver/config"

	"github.com/rs/cors"

	"github.com/ory/kratos/metrics/prometheus"

	"github.com/ory/analytics-go/v4"

	"github.com/ory/x/healthx"

	"github.com/gorilla/context"
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
	"github.com/ory/kratos/selfservice/strategy/password"
	"github.com/ory/kratos/selfservice/strategy/profile"
	"github.com/ory/kratos/session"
	"github.com/ory/kratos/x"
)

type options struct {
	mwf []func(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc)
}

func newOptions(opts []Option) *options {
	o := new(options)
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

func ServePublic(r driver.Registry, wg *sync.WaitGroup, cmd *cobra.Command, args []string, opts ...Option) {
	defer wg.Done()
	modifiers := newOptions(opts)

	ctx := cmd.Context()

	c := r.Config(cmd.Context())
	l := r.Logger()
	n := negroni.New()
	for _, mw := range modifiers.mwf {
		n.UseFunc(mw)
	}

	router := x.NewRouterPublic()
	csrf := x.NewCSRFHandler(router, r)

	n.UseFunc(x.CleanPath) // Prevent double slashes from breaking CSRF.
	r.WithCSRFHandler(csrf)
	n.UseHandler(r.CSRFHandler())

	r.RegisterPublicRoutes(ctx, router)
	n.Use(reqlog.NewMiddlewareFromLogger(l, "public#"+c.SelfPublicURL(nil).String()))
	n.Use(sqa(cmd, r))
	n.Use(r.PrometheusManager())

	if tracer := r.Tracer(ctx); tracer.IsLoaded() {
		n.Use(tracer)
	}

	var handler http.Handler = n
	options, enabled := r.Config(ctx).CORS("public")
	if enabled {
		handler = cors.New(options).Handler(handler)
	}

	server := graceful.WithDefaults(&http.Server{
		Addr:    c.PublicListenOn(),
		Handler: context.ClearHandler(handler),
	})

	l.Printf("Starting the public httpd on: %s", server.Addr)
	if err := graceful.Graceful(server.ListenAndServe, server.Shutdown); err != nil {
		l.Fatalln("Failed to gracefully shutdown public httpd")
	}
	l.Println("Public httpd was shutdown gracefully")
}

func ServeAdmin(r driver.Registry, wg *sync.WaitGroup, cmd *cobra.Command, args []string, opts ...Option) {
	defer wg.Done()
	modifiers := newOptions(opts)
	ctx := cmd.Context()

	c := r.Config(cmd.Context())
	l := r.Logger()
	n := negroni.New()
	for _, mw := range modifiers.mwf {
		n.UseFunc(mw)
	}

	router := x.NewRouterAdmin()
	r.RegisterAdminRoutes(ctx, router)
	n.Use(reqlog.NewMiddlewareFromLogger(l, "admin#"+c.SelfPublicURL(nil).String()))
	n.Use(sqa(cmd, r))
	n.Use(r.PrometheusManager())

	if tracer := r.Tracer(ctx); tracer.IsLoaded() {
		n.Use(tracer)
	}

	n.UseHandler(router)
	server := graceful.WithDefaults(&http.Server{
		Addr:    c.AdminListenOn(),
		Handler: context.ClearHandler(n),
	})

	l.Printf("Starting the admin httpd on: %s", server.Addr)
	if err := graceful.Graceful(server.ListenAndServe, server.Shutdown); err != nil {
		l.Fatalln("Failed to gracefully shutdown admin httpd")
	}
	l.Println("Admin httpd was shutdown gracefully")
}

func sqa(cmd *cobra.Command, d driver.Registry) *metricsx.Service {
	// Creates only ones
	// instance
	return metricsx.New(
		cmd,
		d.Logger(),
		d.Config(cmd.Context()).Source(),
		&metricsx.Options{
			Service: "ory-kratos",
			ClusterID: metricsx.Hash(
				strings.Join([]string{
					d.Config(cmd.Context()).DSN(),
					d.Config(cmd.Context()).SelfPublicURL(nil).String(),
					d.Config(cmd.Context()).SelfAdminURL().String(),
				}, "|"),
			),
			IsDevelopment: d.Config(cmd.Context()).IsInsecureDevMode(),
			WriteKey:      "qQlI6q8Q4WvkzTjKQSor4sHYOikHIvvi",
			WhitelistedPaths: []string{
				"/",
				healthx.AliveCheckPath,
				healthx.ReadyCheckPath,
				healthx.VersionPath,

				password.RouteRegistration,
				password.RouteSettings,

				oidc.RouteBase,

				login.RouteInitBrowserFlow,
				login.RouteInitAPIFlow,
				login.RouteGetFlow,
				login.RouteSubmitFlow,

				logout.RouteBrowser,

				registration.RouteInitBrowserFlow,
				registration.RouteInitAPIFlow,
				registration.RouteGetFlow,

				session.RouteWhoami,
				identity.RouteBase,

				settings.RouteInitBrowserFlow,
				settings.RouteInitAPIFlow,
				settings.RouteGetFlow,

				verification.RouteInitAPIFlow,
				verification.RouteInitBrowserFlow,
				verification.RouteGetFlow,

				profile.RouteSettings,

				link.RouteAdminCreateRecoveryLink,
				link.RouteRecovery,
				link.RouteVerification,

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

func bgTasks(d driver.Registry, wg *sync.WaitGroup, cmd *cobra.Command, args []string) {
	defer wg.Done()

	if d.Config(cmd.Context()).IsBackgroundCourierEnabled() {
		go courier.Watch(cmd.Context(), d)
	}
}

func ServeAll(d driver.Registry, opts ...Option) func(cmd *cobra.Command, args []string) {
	return func(cmd *cobra.Command, args []string) {
		var wg sync.WaitGroup
		wg.Add(3)
		go ServePublic(d, &wg, cmd, args, opts...)
		go ServeAdmin(d, &wg, cmd, args, opts...)
		go bgTasks(d, &wg, cmd, args)
		wg.Wait()
	}
}
