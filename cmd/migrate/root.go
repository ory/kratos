// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package migrate

import (
	"github.com/spf13/cobra"

	"github.com/ory/kratos/cmd/cliclient"
	"github.com/ory/kratos/driver"
	"github.com/ory/x/configx"
	"github.com/ory/x/popx"
)

func NewMigrateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "migrate",
		Short: "Various migration helpers",
	}
}

func RegisterCommandRecursive(parent *cobra.Command) {
	c := NewMigrateCmd()

	configx.RegisterFlags(c.PersistentFlags())
	c.AddCommand(NewMigrateSQLCmd())

	parent.AddCommand(c)
}

func NewMigrateSQLDownCmd(opts ...driver.RegistryOption) *cobra.Command {
	return popx.NewMigrateSQLDownCmd("kratos", func(cmd *cobra.Command, args []string) error {
		return cliclient.NewMigrateHandler().MigrateSQLDown(cmd, args, opts...)
	})
}

func NewMigrateSQLUpCmd(opts ...driver.RegistryOption) *cobra.Command {
	return popx.NewMigrateSQLUpCmd("kratos", func(cmd *cobra.Command, args []string) error {
		return cliclient.NewMigrateHandler().MigrateSQLUp(cmd, args, opts...)
	})
}

func NewMigrateSQLStatusCmd(opts ...driver.RegistryOption) *cobra.Command {
	return popx.NewMigrateSQLStatusCmd("kratos", func(cmd *cobra.Command, args []string) error {
		return cliclient.NewMigrateHandler().MigrateSQLStatus(cmd, args, opts...)
	})
}
