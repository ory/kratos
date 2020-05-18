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
	"os"
	"strconv"

	"github.com/ory/x/viperx"

	"github.com/gobuffalo/packr/v2"

	"github.com/spf13/cobra"

	"github.com/ory/x/flagx"

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

var schemas = packr.New("schemas", "../.schema")

func init() {
	rootCmd.AddCommand(serveCmd)

	disableTelemetryEnv, _ := strconv.ParseBool(os.Getenv("DISABLE_TELEMETRY"))
	serveCmd.PersistentFlags().Bool("disable-telemetry", disableTelemetryEnv, "Disable anonymized telemetry reports - for more information please visit https://www.ory.sh/docs/ecosystem/sqa")
	serveCmd.PersistentFlags().Bool("dev", false, "Disables critical security features to make development easier")
}
