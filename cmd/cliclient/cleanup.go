// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package cliclient

import (
	"github.com/pkg/errors"

	"github.com/ory/x/servicelocatorx"

	"github.com/ory/x/contextx"

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

	d, err := driver.NewWithoutInit(
		cmd.Context(),
		cmd.ErrOrStderr(),
		servicelocatorx.NewOptions(),
		nil,
		opts,
	)
	if len(d.Config().DSN(cmd.Context())) == 0 {
		return errors.New(`required config value "dsn" was not set`)
	} else if err != nil {
		return errors.Wrap(err, "An error occurred initializing cleanup")
	}

	err = d.Init(cmd.Context(), &contextx.Default{})
	if err != nil {
		return errors.Wrap(err, "An error occurred initializing cleanup")
	}

	keepLast := flagx.MustGetDuration(cmd, "keep-last")

	err = d.Persister().CleanupDatabase(
		cmd.Context(),
		d.Config().DatabaseCleanupSleepTables(cmd.Context()),
		keepLast,
		d.Config().DatabaseCleanupBatchSize(cmd.Context()))
	if err != nil {
		return errors.Wrap(err, "An error occurred while cleaning up expired data")
	}

	return nil
}
