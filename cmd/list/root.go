package list

import (
	"github.com/spf13/cobra"

	"github.com/ory/kratos/cmd/cliclient"
	"github.com/ory/x/cmdx"
)

func NewListCmd() *cobra.Command {
	c := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List resources",
	}
	cliclient.RegisterClientFlags(c.PersistentFlags())
	cmdx.RegisterFormatFlags(c.PersistentFlags())
	return c
}

func RegisterCommandRecursive(parent *cobra.Command) {
	c := NewListCmd()
	parent.AddCommand(c)
	c.AddCommand(NewListIdentityCmd(c))
}
