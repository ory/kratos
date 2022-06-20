package cleanup

import (
	"github.com/spf13/cobra"

	"github.com/ory/x/configx"
)

func NewCleanupCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "cleanup",
		Short: "Various cleanup helpers",
	}
	configx.RegisterFlags(c.PersistentFlags())
	return c
}

func RegisterCommandRecursive(parent *cobra.Command) {
	c := NewCleanupCmd()
	parent.AddCommand(c)
	c.AddCommand(NewCleanupSQLCmd())
}
