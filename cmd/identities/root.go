package identities

import (
	"github.com/spf13/cobra"

	"github.com/ory/x/cmdx"

	"github.com/ory/kratos/cmd/cliclient"
)

// NewIdentitiesCmd represents the identity command
func NewIdentitiesCmd() *cobra.Command {
	var identitiesCmd = &cobra.Command{
		Use:   "identities",
		Short: "Tools to interact with remote identities",
	}

	cliclient.RegisterClientFlags(identitiesCmd.PersistentFlags())
	cmdx.RegisterFormatFlags(identitiesCmd.PersistentFlags())
	return identitiesCmd
}

func RegisterCommandRecursive(parent *cobra.Command) {
	c := NewIdentitiesCmd()
	parent.AddCommand(c)

	c.AddCommand(NewImportCmd())
	c.AddCommand(NewValidateCmd())
	c.AddCommand(NewListCmd())
	c.AddCommand(NewGetCmd())
	c.AddCommand(NewDeleteCmd())
	c.AddCommand(NewPatchCmd())
}
