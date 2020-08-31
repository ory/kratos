package migrate

import (
	"github.com/spf13/cobra"
)

// migrateCmd represents the migrate command
var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Various migration helpers",
}

func RegisterCommandRecursive(parent *cobra.Command) {
	parent.AddCommand(migrateCmd)

	migrateCmd.AddCommand(migrateSqlCmd)
}
