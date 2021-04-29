package remote

import (
	"github.com/spf13/cobra"

	"github.com/ory/x/cmdx"

	"github.com/ory/kratos/cmd/cliclient"
)

var remoteCmd = &cobra.Command{
	Use:   "remote",
	Short: "Helpers and management for remote Ory Kratos instances",
}

func RegisterCommandRecursive(parent *cobra.Command) {
	parent.AddCommand(remoteCmd)

	remoteCmd.AddCommand(versionCmd)
	remoteCmd.AddCommand(statusCmd)
}

func init() {
	cliclient.RegisterClientFlags(remoteCmd.PersistentFlags())
	cmdx.RegisterFormatFlags(remoteCmd.PersistentFlags())
}
