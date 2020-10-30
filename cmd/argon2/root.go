package argon2

import "github.com/spf13/cobra"

var rootCmd = &cobra.Command{
	Use: "argon2",
}

func RegisterCommandRecursive(parent *cobra.Command) {
	parent.AddCommand(rootCmd)

	rootCmd.AddCommand(NewValuesCmd())
}
