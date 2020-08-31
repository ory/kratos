package cmd

import (
	"fmt"
	"os"

	"github.com/ory/kratos/cmd/identities"
	"github.com/ory/kratos/cmd/jsonnet"
	"github.com/ory/kratos/cmd/migrate"
	"github.com/ory/kratos/cmd/serve"
	"github.com/ory/kratos/internal/clihelpers"
	"github.com/ory/x/cmdx"

	"github.com/ory/x/viperx"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use: "kratos",
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	viperx.RegisterConfigFlag(rootCmd, "kratos")

	identities.RegisterCommandRecursive(rootCmd)
	jsonnet.RegisterCommandRecursive(rootCmd)
	serve.RegisterCommandRecursive(rootCmd)
	migrate.RegisterCommandRecursive(rootCmd)

	rootCmd.AddCommand(cmdx.Version(&clihelpers.BuildVersion, &clihelpers.BuildGitHash, &clihelpers.BuildTime))
}
