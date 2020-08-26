package jsonnet

import (
	"github.com/spf13/cobra"
)

// jsonnetCmd represents the jsonnet command
var jsonnetCmd = &cobra.Command{
	Use:   "jsonnet",
	Short: "Helpers for linting and formatting JSONNet code",
}

func RegisterCommandRecursive(parent *cobra.Command) {
	parent.AddCommand(jsonnetCmd)

	jsonnetCmd.AddCommand(jsonnetFormatCmd)
	jsonnetCmd.AddCommand(jsonnetLintCmd)
}
