package hashers

import (
	"github.com/ory/kratos/cmd/hashers/argon2"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use: "hashers",
	Short: "This command contains helpers around hashing.",
}

func RegisterCommandRecursive(parent *cobra.Command) {
	parent.AddCommand(rootCmd)

	argon2.RegisterCommandRecursive(rootCmd)
}
