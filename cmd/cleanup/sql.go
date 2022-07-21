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
package cleanup

import (
	"fmt"
	"time"

	"github.com/ory/x/cmdx"

	"github.com/spf13/cobra"

	"github.com/ory/kratos/driver/config"

	"github.com/ory/kratos/cmd/cliclient"
	"github.com/ory/x/configx"
)

// cleanupSqlCmd represents the sql command
func NewCleanupSQLCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "sql <database-url>",
		Short: "Cleanup sql database from expired flows and sessions",
		Long: `Run this command as frequently as you need.
It is recommended to run this command close to the SQL instance (e.g. same subnet) instead of over the public internet.
This decreases risk of failure and decreases time required.
You can read in the database URL using the -e flag, for example:
	export DSN=...
	kratos cleanup sql -e
### WARNING ###
Before running this command on an existing database, create a back up!
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			err := cliclient.NewCleanupHandler().CleanupSQL(cmd, args)
			if err != nil {
				fmt.Fprintln(cmd.OutOrStdout(), err)
				return cmdx.FailSilently(cmd)
			}
			return nil
		},
	}

	configx.RegisterFlags(c.PersistentFlags())
	c.Flags().BoolP("read-from-env", "e", true, "If set, reads the database connection string from the environment variable DSN or config file key dsn.")
	c.Flags().Duration(config.ViperKeyDatabaseCleanupSleepTables, time.Minute, "How long to wait between each table cleanup")
	c.Flags().IntP(config.ViperKeyDatabaseCleanupBatchSize, "b", 100, "Set the number of records to be cleaned per run")
	c.Flags().Duration("keep-last", 0, "Don't remove records younger than")
	return c
}
