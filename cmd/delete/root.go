package delete

import (
	"github.com/spf13/cobra"

	"github.com/ory/kratos/cmd/cliclient"
	"github.com/ory/x/cmdx"
)

func NewDeleteCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "delete",
		Short: "Delete resources",
	}
	cliclient.RegisterClientFlags(c.PersistentFlags())
	cmdx.RegisterFormatFlags(c.PersistentFlags())
	return c
}

func RegisterCommandRecursive(parent *cobra.Command) {
	c := NewDeleteCmd()
	parent.AddCommand(c)
	c.AddCommand(NewDeleteIdentityCmd(c))
}
