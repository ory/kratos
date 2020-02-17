// Copyright © 2018 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"
	"os"
	"strconv"

	"github.com/fsnotify/fsnotify"
	"github.com/gobuffalo/packr/v2"
	"github.com/ory/viper"
	"github.com/ory/x/flagx"
	"github.com/ory/x/viperx"

	"github.com/spf13/cobra"

	"github.com/ory/kratos/cmd/daemon"
	"github.com/ory/kratos/driver"
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use: "serve",
	Run: func(cmd *cobra.Command, args []string) {
		logger = viperx.InitializeConfig("kratos", "", logger)
		// plog.Logger = gbl.Logrus{FieldLogger: logger}

		dev := flagx.MustGetBool(cmd, "dev")
		if dev {
			logger.Warn(`

YOU ARE RUNNING ORY KRATOS IN DEV MODE.
SECURITY IS DISABLED.
DON'T DO THIS IN PRODUCTION!

`)
		}

		watchAndValidateViper()
		daemon.ServeAll(driver.MustNewDefaultDriver(logger, BuildVersion, BuildTime, BuildGitHash, dev))(cmd, args)
	},
}

var schemas = packr.New("schemas", "../docs")

func init() {
	rootCmd.AddCommand(serveCmd)

	disableTelemetryEnv, _ := strconv.ParseBool(os.Getenv("DISABLE_TELEMETRY"))
	serveCmd.PersistentFlags().Bool("disable-telemetry", disableTelemetryEnv, "Disable anonymized telemetry reports - for more information please visit https://www.ory.sh/docs/ecosystem/sqa")
	serveCmd.PersistentFlags().Bool("dev", false, "Disables critical security features to make development easier")
}

func watchAndValidateViper() {
	schema, err := schemas.Find("config.schema.json")
	if err != nil {
		logger.WithError(err).Fatal("Unable to open configuration JSON Schema.")
	}

	if err := viperx.Validate("config.schema.json", schema); err != nil {
		viperx.LoggerWithValidationErrorFields(logger, err).
			Fatal("The configuration is invalid and could not be loaded.")
	}

	viperx.AddWatcher(func(event fsnotify.Event) error {
		if err := viperx.Validate("config.schema.json", schema); err != nil {
			viperx.LoggerWithValidationErrorFields(logger, err).
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
}
