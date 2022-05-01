package validate

import (
	"github.com/spf13/cobra"

	"github.com/ory/x/cmdx"

	"github.com/ory/kratos/cmd/cliclient"
)

func NewValidateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate resources",
	}
	cliclient.RegisterClientFlags(cmd.PersistentFlags())
	cmdx.RegisterFormatFlags(cmd.PersistentFlags())
	return cmd
}

func RegisterCommandRecursive(parent *cobra.Command) {
	c := NewValidateCmd()
	parent.AddCommand(c)
	c.AddCommand(NewValidateIdentityCmd())
}
