package lint

import (
	"github.com/spf13/cobra"
)

func NewLintCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "lint",
		Short: "Helpers for linting code",
	}
	return c
}

func RegisterCommandRecursive(parent *cobra.Command) {
	c := NewLintCmd()
	parent.AddCommand(c)
	c.AddCommand(NewLintJsonnetCmd())
}
