package courier

import (
	cx "context"
	"net/http"

	"golang.org/x/sync/errgroup"

	"github.com/spf13/cobra"
	"github.com/urfave/negroni"

	"github.com/ory/graceful"
	"github.com/ory/kratos/driver"
	"github.com/ory/kratos/x"
	"github.com/ory/x/configx"
	"github.com/ory/x/otelx"
	"github.com/ory/x/reqlog"
)

func NewWatchCmd() *cobra.Command {
	var c = &cobra.Command{
		Use:   "watch",
		Short: "Starts the Ory Kratos message courier",
		RunE: func(cmd *cobra.Command, args []string) error {
			r, err := driver.New(cmd.Context(), cmd.ErrOrStderr(), configx.WithFlags(cmd.Flags()))
			if err != nil {
				return err
			}

			return StartCourier(cmd.Context(), r)
		},
	}
	c.PersistentFlags().Int("expose-metrics-port", 0, "The port to expose the metrics endpoint on (not exposed by default)")
	return c
}

func StartCourier(ctx cx.Context, r driver.Registry) error {
	eg, ctx := errgroup.WithContext(ctx)

	if r.Config().CourierExposeMetricsPort(ctx) != 0 {
		eg.Go(func() error {
			return ServeMetrics(ctx, r)
		})
	}

	eg.Go(func() error {
		return Watch(ctx, r)
	})

	return eg.Wait()
}

func ServeMetrics(ctx cx.Context, r driver.Registry) error {
	c := r.Config()
	l := r.Logger()
	n := negroni.New()

	router := x.NewRouterAdmin()

	r.MetricsHandler().SetRoutes(router.Router)
	n.Use(reqlog.NewMiddlewareFromLogger(l, "admin#"+c.SelfPublicURL(ctx).String()))
	n.Use(r.PrometheusManager())

	n.UseHandler(router)

	var handler http.Handler = n
	if tracer := r.Tracer(ctx); tracer.IsLoaded() {
		handler = otelx.NewHandler(n, "cmd.courier.ServeMetrics")
	}

	// #nosec G112 - the correct settings are set by graceful.WithDefaults
	server := graceful.WithDefaults(&http.Server{
		Addr:    c.MetricsListenOn(ctx),
		Handler: handler,
	})

	l.Printf("Starting the metrics httpd on: %s", server.Addr)
	if err := graceful.Graceful(func() error {
		errChan := make(chan error, 1)
		go func(errChan chan error) {
			if err := server.ListenAndServe(); err != nil {
				errChan <- err
			}
		}(errChan)
		select {
		case err := <-errChan:
			l.Errorf("Failed to start the metrics httpd: %s\n", err)
			return err
		case <-ctx.Done():
			l.Printf("Shutting down the metrics httpd: context closed: %s\n", ctx.Err())
			return server.Shutdown(ctx)
		}
	}, server.Shutdown); err != nil {
		l.Errorln("Failed to gracefully shutdown metrics httpd")
		return err
	} else {
		l.Println("Metrics httpd was shutdown gracefully")
	}
	return nil
}

func Watch(ctx cx.Context, r driver.Registry) error {
	ctx, cancel := cx.WithCancel(ctx)

	r.Logger().Println("Courier worker started.")
	if err := graceful.Graceful(func() error {
		return r.Courier(ctx).Work(ctx)
	}, func(_ cx.Context) error {
		cancel()
		return nil
	}); err != nil {
		r.Logger().WithError(err).Error("Failed to run courier worker.")
		return err
	}

	r.Logger().Println("Courier worker was shutdown gracefully.")
	return nil
}
