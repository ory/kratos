package cmd

import (
	"fmt"
	"os"

	"github.com/ory/x/logrusx"
	"github.com/ory/x/viperx"

	"github.com/spf13/cobra"
)

var logger *logrusx.Logger

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
}
