/*
Copyright © 2019 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package migrate

import (
	"github.com/spf13/cobra"

	"github.com/ory/kratos/cmd/cliclient"
	"github.com/ory/x/configx"
)

// migrateSqlCmd represents the sql command
func NewMigrateSQLCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "sql <database-url>",
		Short: "Create SQL schemas and apply migration plans",
		Long: `Run this command on a fresh SQL installation and when you upgrade Ory Kratos to a new minor version.

It is recommended to run this command close to the SQL instance (e.g. same subnet) instead of over the public internet.
This decreases risk of failure and decreases time required.

You can read in the database URL using the -e flag, for example:
	export DSN=...
	kratos migrate sql -e

### WARNING ###

Before running this command on an existing database, create a back up!
`,
		Run: func(cmd *cobra.Command, args []string) {
			cliclient.NewMigrateHandler().MigrateSQL(cmd, args)
		},
	}

	configx.RegisterFlags(c.PersistentFlags())
	c.Flags().BoolP("read-from-env", "e", false, "If set, reads the database connection string from the environment variable DSN or config file key dsn.")
	c.Flags().BoolP("yes", "y", false, "If set all confirmation requests are accepted without user interaction.")
	return c
}
