package format

import (
	"github.com/spf13/cobra"
)

func NewFormatCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "format",
		Short: "Helpers for formatting code",
	}
	return c
}

func RegisterCommandRecursive(parent *cobra.Command) {
	c := NewFormatCmd()
	parent.AddCommand(c)
	c.AddCommand(NewFormatJsonnetCmd())
}
