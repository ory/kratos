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
		r := driver.New(cmd.Context(), configx.WithFlags(cmd.Flags()))
		StartCourier(cmd.Context(), r)
	},
}

func init() {
	watchCmd.PersistentFlags().Int("expose-metrics-port", 0, "The port to expose the metrics endpoint on (not exposed by default)")
}

func StartCourier(ctx cx.Context, r driver.Registry) {
	c := r.Config(ctx)

	if c.CourierExposeMetricsPort() != 0 {
		go ServeMetrics(ctx, r)
	}
	Watch(ctx, r)
}

func ServeMetrics(ctx cx.Context, r driver.Registry) {

	c := r.Config(ctx)
	l := r.Logger()
	n := negroni.New()

	router := x.NewRouterAdmin()

	r.MetricsHandler().SetRoutes(router.Router)
	n.Use(reqlog.NewMiddlewareFromLogger(l, "admin#"+c.SelfPublicURL().String()))
	n.Use(r.PrometheusManager())

	if tracer := r.Tracer(ctx); tracer.IsLoaded() {
		n.Use(tracer)
	}

	n.UseHandler(router)

	server := graceful.WithDefaults(&http.Server{
		Addr:    c.MetricsListenOn(),
		Handler: context.ClearHandler(n),
	})

	l.Printf("Starting the metrics httpd on: %s", server.Addr)
	if err := graceful.Graceful(func() error {
		go func() {
			if err := server.ListenAndServe(); err != nil {
				l.Fatalf("Failed to start the metrics httpd: %s\n", err)
				return
			}
		}()
		<-ctx.Done()
		return server.Shutdown(ctx)
	}, server.Shutdown); err != nil {
		l.Errorln("Failed to gracefully shutdown metrics httpd")
	} else {
		l.Println("Metrics httpd was shutdown gracefully")
	}
}

func Watch(ctx cx.Context, r driver.Registry) {

	ctx, cancel := cx.WithCancel(ctx)

	r.Logger().Println("Courier worker started.")
	if err := graceful.Graceful(func() error {
		return r.Courier(ctx).Work(ctx)
	}, func(_ cx.Context) error {
		cancel()
		return nil
	}); err != nil {
		r.Logger().WithError(err).Fatalf("Failed to run courier worker.")
	}

	r.Logger().Println("Courier worker was shutdown gracefully.")

}
