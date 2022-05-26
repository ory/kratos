package cliclient

import (
	"github.com/pkg/errors"

	"github.com/ory/x/configx"

	"github.com/spf13/cobra"

	"github.com/ory/kratos/driver"
	"github.com/ory/kratos/driver/config"
	"github.com/ory/x/flagx"
)

type CleanupHandler struct{}

func NewCleanupHandler() *CleanupHandler {
	return &CleanupHandler{}
}

func (h *CleanupHandler) CleanupSQL(cmd *cobra.Command, args []string) error {
	opts := []configx.OptionModifier{
		configx.WithFlags(cmd.Flags()),
		configx.SkipValidation(),
	}

	if !flagx.MustGetBool(cmd, "read-from-env") {
		if len(args) != 1 {
			return errors.New(`expected to get the DSN as an argument, or the "read-from-env" flag`)
		}
		opts = append(opts, configx.WithValue(config.ViperKeyDSN, args[0]))
	}

	d := driver.NewWithoutInit(
		cmd.Context(),
		cmd.ErrOrStderr(),
		opts...,
	)
	if len(d.Config(cmd.Context()).DSN()) == 0 {
		return errors.New(`required config value "dsn" was not set`)
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
