package identities

import (
	"github.com/spf13/cobra"
)

// identitiesCmd represents the identity command
var identitiesCmd = &cobra.Command{
	Use: "identities",
}

func RegisterCommandRecursive(parent *cobra.Command) {
	parent.AddCommand(identitiesCmd)

	identitiesCmd.AddCommand(importCmd)
}

func init() {
	identitiesCmd.PersistentFlags().String("endpoint", "", "Specifies the Ory Kratos Admin URL. Defaults to KRATOS_URLS_ADMIN")
}
