// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package courier

import (
	"context"
	"net/http"

	"github.com/spf13/cobra"
	"github.com/urfave/negroni"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"golang.org/x/sync/errgroup"

	"github.com/ory/graceful"
	"github.com/ory/kratos/driver"
	"github.com/ory/kratos/x"
	"github.com/ory/x/configx"
	"github.com/ory/x/otelx"
	"github.com/ory/x/reqlog"
	"github.com/ory/x/servicelocatorx"
)

func NewWatchCmd(slOpts []servicelocatorx.Option, dOpts []driver.RegistryOption) *cobra.Command {
	c := &cobra.Command{
		Use:   "watch",
		Short: "Starts the Ory Kratos message courier",
		RunE: func(cmd *cobra.Command, args []string) error {
			r, err := driver.New(cmd.Context(), cmd.ErrOrStderr(), servicelocatorx.NewOptions(slOpts...), dOpts, []configx.OptionModifier{configx.WithFlags(cmd.Flags())})
			if err != nil {
				return err
			}

			return StartCourier(cmd.Context(), r)
		},
	}
	c.PersistentFlags().Int("expose-metrics-port", 0, "The port to expose the metrics endpoint on (not exposed by default)")
	return c
}

func StartCourier(ctx context.Context, r driver.Registry) error {
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

func ServeMetrics(ctx context.Context, r driver.Registry) error {
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
		tp := tracer.Provider()
		handler = otelx.NewHandler(handler, "cmd.courier.ServeMetrics", otelhttp.WithTracerProvider(tp))
	}

	//#nosec G112 -- the correct settings are set by graceful.WithDefaults
	server := graceful.WithDefaults(&http.Server{
		Addr:    c.MetricsListenOn(ctx),
		Handler: handler,
	})

	l.Printf("Starting the metrics httpd on: %s", server.Addr)
	if err := graceful.GracefulContext(ctx, server.ListenAndServe, server.Shutdown); err != nil {
		l.Errorln("Failed to gracefully shutdown metrics httpd")
		return err
	}
	l.Println("Metrics httpd was shutdown gracefully")
	return nil
}

func Watch(ctx context.Context, r driver.Registry) error {
	ctx, cancel := context.WithCancel(ctx)

	r.Logger().Println("Courier worker started.")
	if err := graceful.Graceful(func() error {
		c, err := r.Courier(ctx)
		if err != nil {
			return err
		}

		return c.Work(ctx)
	}, func(_ context.Context) error {
		cancel()
		return nil
	}); err != nil {
		r.Logger().WithError(err).Error("Failed to run courier worker.")
		return err
	}

	r.Logger().Println("Courier worker was shutdown gracefully.")
	return nil
}
