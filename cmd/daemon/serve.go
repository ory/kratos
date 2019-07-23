package daemon

import (
	"net/http"
	"sync"

	"github.com/ory/x/healthx"

	"github.com/sirupsen/logrus"

	"github.com/gorilla/context"
	"github.com/spf13/cobra"
	"github.com/urfave/negroni"

	"github.com/ory/graceful"
	"github.com/ory/x/metricsx"

	"github.com/ory/hive/driver"
	"github.com/ory/hive/selfservice"
	"github.com/ory/hive/selfservice/password"
	"github.com/ory/hive/x"
)

func servePublic(d driver.Driver, wg *sync.WaitGroup, cmd *cobra.Command, args []string) {
	defer wg.Done()

	c := d.Configuration()
	l := d.Logger()
	n := negroni.New()
	r := d.Registry()

	telemetry(cmd, n, d)

	router := x.NewRouterPublic()
	r.StrategyHandler().RegisterPublicRoutes(router)
	r.SessionHandler().RegisterPublicRoutes(router)

	n.Use(NewNegroniLoggerMiddleware(l.(*logrus.Logger), "public#"+c.SelfPublicURL().String()))
	n.UseHandler(
		r.NoSurf(router),
	)
	server := graceful.WithDefaults(&http.Server{
		Addr:    c.PublicListenOn(),
		Handler: context.ClearHandler(n),
	})

	l.Printf("Starting the public httpd on: %s", server.Addr)
	if err := graceful.Graceful(server.ListenAndServe, server.Shutdown); err != nil {
		l.Fatalln("Failed to gracefully shutdown public httpd")
	}
	l.Println("Public httpd was shutdown gracefully")
}

func serveAdmin(d driver.Driver, wg *sync.WaitGroup, cmd *cobra.Command, args []string) {
	defer wg.Done()

	c := d.Configuration()
	l := d.Logger()
	n := negroni.New()
	r := d.Registry()

	router := x.NewRouterAdmin()
	r.IdentityHandler().RegisterAdminRoutes(router)
	r.SessionHandler().RegisterAdminRoutes(router)
	r.ErrorHandler().RegisterAdminRoutes(router)

	n.Use(NewNegroniLoggerMiddleware(l.(*logrus.Logger), "admin#"+c.SelfAdminURL().String()))
	telemetry(cmd, n, d)

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

func telemetry(cmd *cobra.Command, n *negroni.Negroni, d driver.Driver) {
	m := metricsx.New(
		cmd,
		d.Logger(),
		&metricsx.Options{
			Service:       "ory-hive",
			ClusterID:     metricsx.Hash(d.Configuration().DSN()),
			IsDevelopment: d.Configuration().DSN() == "memory",
			WriteKey:      "qQlI6q8Q4WvkzTjKQSor4sHYOikHIvvi",
			WhitelistedPaths: []string{
				"/",
				healthx.AliveCheckPath,
				healthx.ReadyCheckPath,
				healthx.VersionPath,
				"/auth/methods/oidc/",
				password.RegistrationPath,
				password.LoginPath,
				selfservice.BrowserLoginPath,
				selfservice.BrowserLoginRequestsPath,
				selfservice.BrowserLogoutPath,
				selfservice.BrowserRegistrationPath,
				selfservice.BrowserRegistrationRequestsPath,
			},
			BuildVersion: d.BuildInfo().Version,
			BuildHash:    d.BuildInfo().Hash,
			BuildTime:    d.BuildInfo().Time,
		},
	)
	n.Use(m)
}

func ServeAll(d driver.Driver) func(cmd *cobra.Command, args []string) {
	return func(cmd *cobra.Command, args []string) {
		var wg sync.WaitGroup
		wg.Add(2)
		go servePublic(d, &wg, cmd, args)
		go serveAdmin(d, &wg, cmd, args)
		wg.Wait()
	}
}
