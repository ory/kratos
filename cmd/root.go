package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/ory/kratos/driver/config"

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

// RootCmd represents the base command when called without any subcommands
func NewRootCmd() (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use: "kratos",
	}
	identities.RegisterCommandRecursive(cmd)

	jsonnet.RegisterCommandRecursive(cmd)
	serve.RegisterCommandRecursive(cmd)
	migrate.RegisterCommandRecursive(cmd)
	remote.RegisterCommandRecursive(cmd)
	hashers.RegisterCommandRecursive(cmd)
	courier.RegisterCommandRecursive(cmd)

	cmd.AddCommand(cmdx.Version(&config.Version, &config.Commit, &config.Date))

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
