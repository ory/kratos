package jsonnet

import (
	"github.com/spf13/cobra"

	"github.com/ory/kratos/cmd/jsonnet/format"
	"github.com/ory/kratos/cmd/jsonnet/lint"
)

// jsonnetCmd represents the jsonnet command
var jsonnetCmd = &cobra.Command{
	Use:   "jsonnet",
	Short: "Helpers for linting and formatting JSONNet code",
}

func RegisterCommandRecursive(parent *cobra.Command) {
	parent.AddCommand(jsonnetCmd)

	format.RegisterCommandRecursive(jsonnetCmd)
	lint.RegisterCommandRecursive(jsonnetCmd)
}
