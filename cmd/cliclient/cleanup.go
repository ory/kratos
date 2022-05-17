package cliclient

import (
	"fmt"

	"github.com/pkg/errors"

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

func (h *CleanupHandler) CleanupSQL(cmd *cobra.Command, args []string) error {
	var d driver.Registry

	if flagx.MustGetBool(cmd, "read-from-env") {
		d = driver.NewWithoutInit(
			cmd.Context(),
			cmd.ErrOrStderr(),
			configx.WithFlags(cmd.Flags()),
			configx.SkipValidation())
		if len(d.Config(cmd.Context()).DSN()) == 0 {
			fmt.Fprintln(cmd.OutOrStdout(), cmd.UsageString())
			fmt.Fprintln(cmd.OutOrStdout(), "")
			fmt.Fprintln(cmd.OutOrStdout(), "When using flag -e, environment variable DSN must be set")
			return cmdx.FailSilently(cmd)
		}
	} else {
		if len(args) != 1 {
			fmt.Println(cmd.UsageString())
			return cmdx.FailSilently(cmd)
		}
		d = driver.NewWithoutInit(
			cmd.Context(),
			cmd.ErrOrStderr(),
			configx.WithFlags(cmd.Flags()),
			configx.SkipValidation(),
			configx.WithValue(config.ViperKeyDSN, args[0]))
	}

	err := d.Init(cmd.Context(), driver.SkipNetworkInit)
	if err != nil {
		return errors.Wrap(err, "An error occurred initializing cleanup")
	}

	keepLast := flagx.MustGetDuration(cmd, "keep-last")

	err = d.Persister().CleanupDatabase(
		cmd.Context(),
		d.Config(cmd.Context()).DatabaseCleanupSleepTables(),
		keepLast,
		d.Config(cmd.Context()).DatabaseCleanupBatchSize())
	if err != nil {
		return errors.Wrap(err, "An error occurred while cleaning up expired data")
	}

	return nil
}
