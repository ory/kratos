package courier

import (
	cx "context"

	"github.com/spf13/cobra"

	"github.com/ory/graceful"
	"github.com/ory/kratos/driver"
	"github.com/ory/x/configx"
)

var watchCmd = &cobra.Command{
	Use:   "watch",
	Short: "Starts the ORY Kratos message courier",
	Run: func(cmd *cobra.Command, args []string) {
		d := driver.New(cmd.Context(), configx.WithFlags(cmd.Flags()))
		Watch(d, cmd, args)
	},
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
