// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package migrate

import (
	"github.com/spf13/cobra"

	"github.com/ory/kratos/cmd/cliclient"
	"github.com/ory/kratos/driver"
	"github.com/ory/x/configx"
)

// migrateSqlCmd represents the sql command
func NewMigrateSQLCmd(opts ...driver.RegistryOption) *cobra.Command {
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
		RunE: func(cmd *cobra.Command, args []string) error {
			return cliclient.NewMigrateHandler().MigrateSQL(cmd, args, opts...)
		},
	}

	configx.RegisterFlags(c.PersistentFlags())
	c.Flags().BoolP("read-from-env", "e", false, "If set, reads the database connection string from the environment variable DSN or config file key dsn.")
	c.Flags().BoolP("yes", "y", false, "If set all confirmation requests are accepted without user interaction.")
	return c
}
