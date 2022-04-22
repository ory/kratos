package cleanup

import (
	"github.com/ory/x/configx"
	"github.com/spf13/cobra"
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
