package identities

import (
	"github.com/spf13/cobra"

	"github.com/ory/kratos/cmd/cliclient"
)

// identitiesCmd represents the identity command
var identitiesCmd = &cobra.Command{
	Use: "identities",
}

func RegisterCommandRecursive(parent *cobra.Command) {
	parent.AddCommand(identitiesCmd)

	identitiesCmd.AddCommand(importCmd)
	identitiesCmd.AddCommand(validateCmd)
	identitiesCmd.AddCommand(listCmd)
}

func init() {
	cliclient.RegisterClientFlags(identitiesCmd.PersistentFlags())
}
