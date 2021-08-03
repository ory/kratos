package hashers

import (
	"github.com/spf13/cobra"

	"github.com/ory/kratos/cmd/hashers/argon2"
)

func NewRootCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "hashers",
		Short: "This command contains helpers around hashing",
	}
	return c
}

func RegisterCommandRecursive(parent *cobra.Command) {
	rootCmd := NewRootCmd()
	parent.AddCommand(rootCmd)

	argon2.RegisterCommandRecursive(rootCmd)
}
