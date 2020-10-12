package identities

import (
	"github.com/ory/x/cmdx"
	"github.com/spf13/cobra"

	"github.com/ory/kratos/cmd/cliclient"
)

// identitiesCmd represents the identity command
var identitiesCmd = &cobra.Command{
	Use:   "identities",
	Short: "Tools to interact with remote identities",
}

func RegisterCommandRecursive(parent *cobra.Command) {
	parent.AddCommand(identitiesCmd)

	identitiesCmd.AddCommand(importCmd)
	identitiesCmd.AddCommand(validateCmd)
	identitiesCmd.AddCommand(listCmd)
	identitiesCmd.AddCommand(getCmd)
	identitiesCmd.AddCommand(deleteCmd)
	identitiesCmd.AddCommand(patchCmd)
}

func init() {
	cliclient.RegisterClientFlags(identitiesCmd.PersistentFlags())
	cmdx.RegisterFormatFlags(identitiesCmd.PersistentFlags())
}
