package migrate

import (
	"github.com/spf13/cobra"
)

func NewMigrateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "migrate",
		Short: "Various migration helpers",
	}
}

func RegisterCommandRecursive(parent *cobra.Command) {
	c := NewMigrateCmd()
	parent.AddCommand(c)
	c.AddCommand(NewMigrateSQLCmd())
}
