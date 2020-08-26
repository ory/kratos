package migrate

import (
	"github.com/spf13/cobra"

	"github.com/ory/kratos/cmd/migrate/sql"
)

// migrateCmd represents the migrate command
var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Various migration helpers",
}

func RegisterCommandRecursive(parent *cobra.Command) {
	parent.AddCommand(migrateCmd)

	sql.RegisterCommandRecursive(migrateCmd)
}
