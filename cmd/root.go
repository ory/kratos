package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/ory/kratos/driver/config"

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
var RootCmd = &cobra.Command{
	Use: "kratos",
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the RootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		if !errors.Is(err, cmdx.ErrNoPrintButFail) {
			fmt.Fprintln(RootCmd.ErrOrStderr(), err)
		}
		os.Exit(1)
	}
}

func init() {
	identities.RegisterCommandRecursive(RootCmd)
	identities.RegisterFlags()

	jsonnet.RegisterCommandRecursive(RootCmd)
	serve.RegisterCommandRecursive(RootCmd)
	migrate.RegisterCommandRecursive(RootCmd)
	remote.RegisterCommandRecursive(RootCmd)
	hashers.RegisterCommandRecursive(RootCmd)

	RootCmd.AddCommand(cmdx.Version(&config.Version, &config.Commit, &config.Date))
}
