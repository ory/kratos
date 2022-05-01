package get

import (
	"github.com/spf13/cobra"

	"github.com/ory/kratos/cmd/cliclient"
	"github.com/ory/x/cmdx"
)

func NewGetCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "get",
		Short: "Get resources",
	}
	cliclient.RegisterClientFlags(cmd.PersistentFlags())
	cmdx.RegisterFormatFlags(cmd.PersistentFlags())
	return cmd
}

func RegisterCommandRecursive(parent *cobra.Command) {
	c := NewGetCmd()
	parent.AddCommand(c)
	c.AddCommand(NewGetIdentityCmd(parent))
}
