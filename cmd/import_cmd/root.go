package import_cmd

import (
	"github.com/spf13/cobra"

	"github.com/ory/x/cmdx"

	"github.com/ory/kratos/cmd/cliclient"
)

func NewImportCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "import_cmd",
		Short: "Import resources",
	}
	cliclient.RegisterClientFlags(cmd.PersistentFlags())
	cmdx.RegisterFormatFlags(cmd.PersistentFlags())
	return cmd
}

func RegisterCommandRecursive(parent *cobra.Command) {
	c := NewImportCmd()
	parent.AddCommand(c)
	c.AddCommand(NewImportIdentityCmd(c))
}
