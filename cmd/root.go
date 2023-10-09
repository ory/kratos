// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/ory/kratos/cmd/cleanup"
	"github.com/ory/kratos/driver/config"
	"github.com/ory/x/jsonnetsecure"

	"github.com/ory/kratos/cmd/courier"
	"github.com/ory/kratos/cmd/hashers"

	"github.com/ory/kratos/cmd/remote"

	"github.com/ory/kratos/cmd/identities"
	"github.com/ory/kratos/cmd/jsonnet"
	"github.com/ory/kratos/cmd/migrate"
	"github.com/ory/kratos/cmd/serve"
	"github.com/ory/x/cmdx"

	"github.com/spf13/cobra"
)

func NewRootCmd() (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use: "kratos",
	}
	cmdx.EnableUsageTemplating(cmd)

	courier.RegisterCommandRecursive(cmd, nil, nil)
	cmd.AddCommand(identities.NewGetCmd())
	cmd.AddCommand(identities.NewDeleteCmd())
	cmd.AddCommand(jsonnet.NewFormatCmd())
	hashers.RegisterCommandRecursive(cmd)
	cmd.AddCommand(identities.NewImportCmd())
	cmd.AddCommand(jsonnet.NewLintCmd())
	cmd.AddCommand(identities.NewListCmd())
	migrate.RegisterCommandRecursive(cmd)
	serve.RegisterCommandRecursive(cmd, nil, nil)
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
func Execute() {
	c := NewRootCmd()

	if err := c.Execute(); err != nil {
		if !errors.Is(err, cmdx.ErrNoPrintButFail) {
			_, _ = fmt.Fprintln(c.ErrOrStderr(), err)
		}
		os.Exit(1)
	}
}
