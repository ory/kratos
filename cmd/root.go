// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"errors"
	"fmt"
	"runtime"

	"github.com/spf13/cobra"

	"github.com/ory/kratos/cmd/cleanup"
	"github.com/ory/kratos/cmd/courier"
	"github.com/ory/kratos/cmd/hashers"
	"github.com/ory/kratos/cmd/identities"
	"github.com/ory/kratos/cmd/jsonnet"
	"github.com/ory/kratos/cmd/migrate"
	"github.com/ory/kratos/cmd/remote"
	"github.com/ory/kratos/cmd/serve"
	"github.com/ory/kratos/driver"
	"github.com/ory/kratos/driver/config"
	"github.com/ory/x/cmdx"
	"github.com/ory/x/dbal"
	"github.com/ory/x/jsonnetsecure"
	"github.com/ory/x/profilex"
)

func NewRootCmd(driverOpts ...driver.RegistryOption) (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use: "kratos",
	}
	cmdx.EnableUsageTemplating(cmd)

	courier.RegisterCommandRecursive(cmd, nil, driverOpts)
	cmd.AddCommand(identities.NewGetCmd())
	cmd.AddCommand(identities.NewDeleteCmd())
	cmd.AddCommand(jsonnet.NewFormatCmd())
	hashers.RegisterCommandRecursive(cmd)
	cmd.AddCommand(identities.NewImportCmd())
	cmd.AddCommand(jsonnet.NewLintCmd())
	cmd.AddCommand(identities.NewListCmd())
	migrate.RegisterCommandRecursive(cmd)
	serve.RegisterCommandRecursive(cmd, nil, driverOpts, nil)
	cleanup.RegisterCommandRecursive(cmd)
	remote.RegisterCommandRecursive(cmd)
	cmd.AddCommand(identities.NewValidateCmd())
	cmd.AddCommand(cmdx.Version(&config.Version, &config.Commit, &config.Date))

	// Registers a hidden "jsonnet" subcommand for process-isolated Jsonnet VMs.
	cmd.AddCommand(jsonnetsecure.NewJsonnetCmd())

	return cmd
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the RootCmd.
func Execute() int {
	defer profilex.Profile().Stop()

	dbal.RegisterDriver(func() dbal.Driver {
		return driver.NewRegistryDefault()
	})

	jsonnetPool := jsonnetsecure.NewProcessPool(runtime.GOMAXPROCS(0))
	defer jsonnetPool.Close()

	c := NewRootCmd(driver.WithJsonnetPool(jsonnetPool))

	if err := c.Execute(); err != nil {
		if !errors.Is(err, cmdx.ErrNoPrintButFail) {
			_, _ = fmt.Fprintln(c.ErrOrStderr(), err)
		}
		return 1
	}
	return 0
}
