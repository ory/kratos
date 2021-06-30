package courier

import (
	"github.com/spf13/cobra"

	"github.com/ory/x/configx"
)

// NewCourierCmd creates a new courier command
func NewCourierCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "courier",
		Short: "Commands related to the Ory Kratos message courier",
	}
	configx.RegisterFlags(c.PersistentFlags())
	return c
}

func RegisterCommandRecursive(parent *cobra.Command) {
	c := NewCourierCmd()
	parent.AddCommand(c)
	c.AddCommand(NewWatchCmd())
}
