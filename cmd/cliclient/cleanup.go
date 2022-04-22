package cliclient

import (
	"fmt"
	"os"

	"github.com/ory/x/configx"

	"github.com/spf13/cobra"

	"github.com/ory/kratos/driver"
	"github.com/ory/kratos/driver/config"
	"github.com/ory/x/cmdx"
	"github.com/ory/x/flagx"
)

type CleanupHandler struct{}

func NewCleanupHandler() *CleanupHandler {
	return &CleanupHandler{}
}

func (h *CleanupHandler) CleanupSQL(cmd *cobra.Command, args []string) {
	var d driver.Registry

	if flagx.MustGetBool(cmd, "read-from-env") {
		d = driver.NewWithoutInit(
			cmd.Context(),
			cmd.ErrOrStderr(),
			configx.WithFlags(cmd.Flags()),
			configx.SkipValidation())
		if len(d.Config(cmd.Context()).DSN()) == 0 {
			fmt.Println(cmd.UsageString())
			fmt.Println("")
			fmt.Println("When using flag -e, environment variable DSN must be set")
			os.Exit(1)
			return
		}
	} else {
		if len(args) != 1 {
			fmt.Println(cmd.UsageString())
			os.Exit(1)
			return
		}
		d = driver.NewWithoutInit(
			cmd.Context(),
			cmd.ErrOrStderr(),
			configx.WithFlags(cmd.Flags()),
			configx.SkipValidation(),
			configx.WithValue(config.ViperKeyDSN, args[0]))
	}

	err := d.Init(cmd.Context(), driver.SkipNetworkInit)
	cmdx.Must(err, "An error occurred initializing migrations: %s", err)

	keepLast := flagx.MustGetDuration(cmd, "keep-last")

	err = d.Persister().CleanupDatabase(cmd.Context(), d.Config(cmd.Context()).DatabaseCleanupSleepTables(), keepLast)
	cmdx.Must(err, "An error occurred while cleaning up expired data: %s", err)

}
