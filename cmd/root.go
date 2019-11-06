package cmd

import (
	"fmt"
	"os"

	"github.com/fsnotify/fsnotify"
	"github.com/gobuffalo/packr/v2"
	"github.com/sirupsen/logrus"

	"github.com/ory/gojsonschema"
	"github.com/ory/viper"

	"github.com/ory/x/viperx"

	"github.com/ory/x/logrusx"

	"github.com/spf13/cobra"
)

var logger *logrus.Logger

var schemas = packr.New("schemas", "../docs")

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
	schema, err := schemas.Find("config.schema.json")
	if err != nil {
		panic(err)
	}

	cobra.OnInitialize(func() {
		viperx.InitializeConfig("kratos", "", nil)
		logger = logrusx.New()

		if err := viperx.Validate(gojsonschema.NewBytesLoader(schema)); err != nil {
			viperx.LoggerWithValidationErrorFields(logger, err).
				WithError(err).
				Fatal("The configuration is invalid and could not be loaded.")
		}

		viperx.AddWatcher(func(event fsnotify.Event) error {
			if err := viperx.Validate(gojsonschema.NewBytesLoader(schema)); err != nil {
				viperx.LoggerWithValidationErrorFields(logger, err).
					WithError(err).
					Error("The changed configuration is invalid and could not be loaded. Rolling back to the last working configuration revision. Please address the validation errors before restarting ORY Kratos.")
				return viperx.ErrRollbackConfigurationChanges
			}
			return nil
		})

		viperx.WatchConfig(logger, &viperx.WatchOptions{
			Immutables: []string{"serve", "profiling", "log"},
			OnImmutableChange: func(key string) {
				logger.
					WithField("key", key).
					WithField("reset_to", fmt.Sprintf("%v", viper.Get(key))).
					Error("A configuration value marked as immutable has changed. Rolling back to the last working configuration revision. To reload the values please restart ORY Kratos.")
			},
		})
	})

	viperx.RegisterConfigFlag(rootCmd, "kratos")
}
