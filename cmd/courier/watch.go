package courier

import (
	cx "context"
	"net/http"

	"github.com/gorilla/context"
	"github.com/spf13/cobra"
	"github.com/urfave/negroni"

	"github.com/ory/graceful"
	"github.com/ory/kratos/driver"
	"github.com/ory/kratos/x"
	"github.com/ory/x/configx"
	"github.com/ory/x/reqlog"
)

var watchCmd = &cobra.Command{
	Use:   "watch",
	Short: "Starts the ORY Kratos message courier",
	Run: func(cmd *cobra.Command, args []string) {
		d := driver.New(cmd.Context(), configx.WithFlags(cmd.Flags()))

		metricsPort := d.Config(cmd.Context()).CourierExposeMetricsPort()

		if d.Config(cmd.Context()).CourierExposeMetricsPort() != 0 {
			go ServeMetrics(d, cmd, metricsPort, args)
		}
		Watch(d, cmd, args)
	},
}

func ServeMetrics(r driver.Registry, cmd *cobra.Command, port int, args []string) {

	c := r.Config(cmd.Context())
	l := r.Logger()
	n := negroni.New()

	router := x.NewRouterAdmin()

	r.MetricsHandler().SetRoutes(router.Router)
	n.Use(reqlog.NewMiddlewareFromLogger(l, "admin#"+c.SelfPublicURL().String()))
	n.Use(r.PrometheusManager())

	if tracer := r.Tracer(cmd.Context()); tracer.IsLoaded() {
		n.Use(tracer)
	}

	n.UseHandler(router)

	server := graceful.WithDefaults(&http.Server{
		Addr:    c.MetricsListenOn(),
		Handler: context.ClearHandler(n),
	})

	l.Printf("Starting the metrics httpd on: %s", server.Addr)
	if err := graceful.Graceful(server.ListenAndServe, server.Shutdown); err != nil {
		l.Fatalln("Failed to gracefully shutdown metrics httpd")
	}
	l.Println("Metrics httpd was shutdown gracefully")
}

func Watch(d driver.Registry, cmd *cobra.Command, args []string) {
	ctx, cancel := cx.WithCancel(cmd.Context())

	d.Logger().Println("Courier worker started.")
	if err := graceful.Graceful(func() error {
		return d.Courier(ctx).Work(ctx)
	}, func(_ cx.Context) error {
		cancel()
		return nil
	}); err != nil {
		d.Logger().WithError(err).Fatalf("Failed to run courier worker.")
	}

	d.Logger().Println("Courier worker was shutdown gracefully.")

}
