package courier

import (
	cx "context"
	"net/http"

	"github.com/gorilla/context"
	"github.com/rs/cors"
	"github.com/spf13/cobra"
	"github.com/urfave/negroni"

	"github.com/ory/graceful"
	"github.com/ory/kratos/driver"
	"github.com/ory/kratos/x"
	"github.com/ory/x/configx"
	"github.com/ory/x/reqlog"
	"github.com/ory/x/stringsx"
)

var watchCmd = &cobra.Command{
	Use:   "watch",
	Short: "Starts the ORY Kratos message courier",
	Run: func(cmd *cobra.Command, args []string) {
		d := driver.New(cmd.Context(), configx.WithFlags(cmd.Flags()))

		go ServeHealth(d, cmd, args)
		Watch(d, cmd, args)
	},
}

func ServeHealth(r driver.Registry, cmd *cobra.Command, args []string) {
	c := r.Config(cmd.Context())
	l := r.Logger()
	n := negroni.New()

	router := x.NewRouterPublic()
	csrf := x.NewCSRFHandler(
		router,
		r.Writer(),
		l,
		stringsx.Coalesce(c.SelfPublicURL().Path, "/"),
		c.SelfPublicURL().Hostname(),
		!c.IsInsecureDevMode(),
	)

	n.UseFunc(x.CleanPath) // Prevent double slashes from breaking CSRF.
	r.WithCSRFHandler(csrf)
	n.UseHandler(r.CSRFHandler())

	r.HealthHandler().SetRoutes(router.Router, false)
	n.Use(reqlog.NewMiddlewareFromLogger(l, "public#"+c.SelfPublicURL().String()))

	var handler http.Handler = n
	options, enabled := r.Config(cmd.Context()).CORS("public")
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

func Watch(d driver.Registry, cmd *cobra.Command, args []string) {
	ctx, cancel := cx.WithCancel(cmd.Context())

	d.Logger().Println("Courier worker started.")
	if err := graceful.Graceful(func() error {
		return d.Courier().Work(ctx)
	}, func(_ cx.Context) error {
		cancel()
		return nil
	}); err != nil {
		d.Logger().WithError(err).Fatalf("Failed to run courier worker.")
	}

	d.Logger().Println("Courier worker was shutdown gracefully.")

}
