package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/ory/x/cmdx"

	"github.com/ory/kratos/cmd/courier"
	"github.com/ory/kratos/cmd/delete"
	"github.com/ory/kratos/cmd/format"
	"github.com/ory/kratos/cmd/get"
	"github.com/ory/kratos/cmd/hashers"
	"github.com/ory/kratos/cmd/import"
	"github.com/ory/kratos/cmd/lint"
	"github.com/ory/kratos/cmd/list"
	"github.com/ory/kratos/cmd/migrate"
	"github.com/ory/kratos/cmd/remote"
	"github.com/ory/kratos/cmd/serve"
	"github.com/ory/kratos/cmd/validate"
	"github.com/ory/kratos/driver/config"
)

// RootCmd represents the base command when called without any subcommands
func NewRootCmd() (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use: "kratos",
	}
	courier.RegisterCommandRecursive(cmd)
	delete.RegisterCommandRecursive(cmd)
	format.RegisterCommandRecursive(cmd)
	get.RegisterCommandRecursive(cmd)
	hashers.RegisterCommandRecursive(cmd)
	import_cmd.RegisterCommandRecursive(cmd)
	lint.RegisterCommandRecursive(cmd)
	list.RegisterCommandRecursive(cmd)
	migrate.RegisterCommandRecursive(cmd)
	remote.RegisterCommandRecursive(cmd)
	serve.RegisterCommandRecursive(cmd)
	validate.RegisterCommandRecursive(cmd)
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
