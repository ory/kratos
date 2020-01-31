package cmd

import (
	"fmt"
	"os"

	gbl "github.com/gobuffalo/logger"
	"github.com/gobuffalo/packr/v2/plog"
	"github.com/sirupsen/logrus"

	"github.com/ory/x/logrusx"

	"github.com/ory/x/viperx"

	"github.com/spf13/cobra"
)

var logger *logrus.Logger

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
	cobra.OnInitialize(func() {
		logger = logrusx.New()
		viperx.InitializeConfig("kratos", "", logger)
		plog.Logger = gbl.Logrus{FieldLogger: logger}
	})

	viperx.RegisterConfigFlag(rootCmd, "kratos")
}
