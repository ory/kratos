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

package serve

import (
	"os"
	"strconv"

	"github.com/ory/kratos/internal/clihelpers"
	"github.com/ory/x/logrusx"

	"github.com/ory/kratos/driver/configuration"

	"github.com/ory/x/viperx"

	"github.com/spf13/cobra"

	"github.com/ory/x/flagx"

	"github.com/ory/kratos/cmd/daemon"
	"github.com/ory/kratos/driver"
	"github.com/ory/kratos/x"
)

var logger *logrusx.Logger

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use: "serve",
	Short: "Run the ORY Kratos server",
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

		x.WatchAndValidateViper(logger)
		d := driver.MustNewDefaultDriver(logger, clihelpers.BuildVersion, clihelpers.BuildTime, clihelpers.BuildGitHash, dev)

		configVersion := d.Configuration().ConfigVersion()
		if configVersion == configuration.UnknownVersion {
			d.Logger().Warn("The config has no version specified. Add the version to improve your development experience.")
		} else if clihelpers.BuildVersion != "" &&
			configVersion != clihelpers.BuildVersion {
			d.Logger().Warnf("Config version is '%s' but kratos runs on version '%s'", configVersion, clihelpers.BuildVersion)
		}

		daemon.ServeAll(d)(cmd, args)
	},
}

func RegisterCommandRecursive(parent *cobra.Command) {
	parent.AddCommand(serveCmd)
}

func init() {
	disableTelemetryEnv, _ := strconv.ParseBool(os.Getenv("DISABLE_TELEMETRY"))
	serveCmd.PersistentFlags().Bool("disable-telemetry", disableTelemetryEnv, "Disable anonymized telemetry reports - for more information please visit https://www.ory.sh/docs/ecosystem/sqa")
	serveCmd.PersistentFlags().Bool("dev", false, "Disables critical security features to make development easier")
}
