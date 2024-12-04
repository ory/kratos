// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package cliclient

import (
	"fmt"

	"github.com/ory/x/popx"
	"github.com/ory/x/servicelocatorx"

	"github.com/pkg/errors"

	"github.com/ory/x/contextx"

	"github.com/ory/x/configx"

	"github.com/spf13/cobra"

	"github.com/ory/kratos/driver"
	"github.com/ory/kratos/driver/config"
	"github.com/ory/x/cmdx"
	"github.com/ory/x/flagx"
)

type MigrateHandler struct{}

func NewMigrateHandler() *MigrateHandler {
	return &MigrateHandler{}
}

func (h *MigrateHandler) getPersister(cmd *cobra.Command, args []string, opts []driver.RegistryOption) (d driver.Registry, err error) {
	if flagx.MustGetBool(cmd, "read-from-env") {
		d, err = driver.NewWithoutInit(
			cmd.Context(),
			cmd.ErrOrStderr(),
			servicelocatorx.NewOptions(),
			nil,
			[]configx.OptionModifier{
				configx.WithFlags(cmd.Flags()),
				configx.SkipValidation(),
			})
		if err != nil {
			return nil, err
		}
		if len(d.Config().DSN(cmd.Context())) == 0 {
			fmt.Println(cmd.UsageString())
			fmt.Println("")
			fmt.Println("When using flag -e, environment variable DSN must be set")
			return nil, cmdx.FailSilently(cmd)
		}
	} else {
		if len(args) != 1 {
			fmt.Println(cmd.UsageString())
			return nil, cmdx.FailSilently(cmd)
		}
		d, err = driver.NewWithoutInit(
			cmd.Context(),
			cmd.ErrOrStderr(),
			servicelocatorx.NewOptions(),
			nil,
			[]configx.OptionModifier{
				configx.WithFlags(cmd.Flags()),
				configx.SkipValidation(),
				configx.WithValue(config.ViperKeyDSN, args[0]),
			})
		if err != nil {
			return nil, err
		}
	}

	err = d.Init(cmd.Context(), &contextx.Default{}, append(opts, driver.SkipNetworkInit)...)
	if err != nil {
		return nil, errors.Wrap(err, "an error occurred initializing migrations")
	}

	return d, nil
}

func (h *MigrateHandler) MigrateSQLDown(cmd *cobra.Command, args []string, opts ...driver.RegistryOption) error {
	p, err := h.getPersister(cmd, args, opts)
	if err != nil {
		return err
	}
	return popx.MigrateSQLDown(cmd, p.Persister())
}

func (h *MigrateHandler) MigrateSQLStatus(cmd *cobra.Command, args []string, opts ...driver.RegistryOption) error {
	p, err := h.getPersister(cmd, args, opts)
	if err != nil {
		return err
	}
	return popx.MigrateStatus(cmd, p.Persister())
}

func (h *MigrateHandler) MigrateSQLUp(cmd *cobra.Command, args []string, opts ...driver.RegistryOption) error {
	p, err := h.getPersister(cmd, args, opts)
	if err != nil {
		return err
	}
	return popx.MigrateSQLUp(cmd, p.Persister())
}
