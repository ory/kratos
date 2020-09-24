package identities

import (
	"github.com/spf13/cobra"

	"github.com/ory/kratos/internal/clihelpers"

	"github.com/ory/kratos/cmd/cliclient"
)

// identitiesCmd represents the identity command
var identitiesCmd = &cobra.Command{
	Use:   "identities",
	Short: "Tools to interact with remote identities",
}

func RegisterCommandRecursive(parent *cobra.Command) {
	parent.AddCommand(identitiesCmd)

	identitiesCmd.AddCommand(putCmd)
	identitiesCmd.AddCommand(validateCmd)
	identitiesCmd.AddCommand(listCmd)
	identitiesCmd.AddCommand(getCmd)
	identitiesCmd.AddCommand(deleteCmd)
	identitiesCmd.AddCommand(patchCmd)
}

func init() {
	cliclient.RegisterClientFlags(identitiesCmd.PersistentFlags())
	clihelpers.RegisterFormatFlags(identitiesCmd.PersistentFlags())
}
