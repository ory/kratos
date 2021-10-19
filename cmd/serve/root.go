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
	"github.com/ory/kratos/driver/config"
	"github.com/ory/x/configx"

	"github.com/spf13/cobra"

	"github.com/ory/kratos/cmd/daemon"
	"github.com/ory/kratos/driver"
)

// serveCmd represents the serve command
func NewServeCmd() (serveCmd *cobra.Command) {
	serveCmd = &cobra.Command{
		Use:   "serve",
		Short: "Run the Ory Kratos server",
		Run: func(cmd *cobra.Command, args []string) {
			d := driver.New(cmd.Context(), cmd.ErrOrStderr(), configx.WithFlags(cmd.Flags()))

			if d.Config(cmd.Context()).IsInsecureDevMode() {
				d.Logger().Warn(`

YOU ARE RUNNING Ory KRATOS IN DEV MODE.
SECURITY IS DISABLED.
DON'T DO THIS IN PRODUCTION!

`)
			}

			configVersion := d.Config(cmd.Context()).ConfigVersion()
			if configVersion == config.UnknownVersion {
				d.Logger().Warn("The config has no version specified. Add the version to improve your development experience.")
			} else if config.Version != "" &&
				configVersion != config.Version {
				d.Logger().Warnf("Config version is '%s' but kratos runs on version '%s'", configVersion, config.Version)
			}

			daemon.ServeAll(d)(cmd, args)
		},
	}
	configx.RegisterFlags(serveCmd.PersistentFlags())

	serveCmd.PersistentFlags().Bool("sqa-opt-out", false, "Disable anonymized telemetry reports - for more information please visit https://www.ory.sh/docs/ecosystem/sqa")
	serveCmd.PersistentFlags().Bool("dev", false, "Disables critical security features to make development easier")
	serveCmd.PersistentFlags().Bool("watch-courier", false, "Run the message courier as a background task, to simplify single-instance setup")
	return serveCmd
}

func RegisterCommandRecursive(parent *cobra.Command) {
	parent.AddCommand(NewServeCmd())
}
