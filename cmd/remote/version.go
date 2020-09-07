package remote

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/ory/kratos/cmd/cliclient"
	"github.com/ory/kratos/internal/clihelpers"
	"github.com/ory/kratos/internal/httpclient/client/version"
	"github.com/ory/x/cmdx"
)

type versionValue struct {
	Version string `json:"version"`
}

func (v *versionValue) Header() []string {
	return []string{"VERSION"}
}

func (v *versionValue) Fields() []string {
	return []string{v.Version}
}

func (v *versionValue) Interface() interface{} {
	return v
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version of the remote Kratos instance.",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		c := cliclient.NewClient(cmd)

		resp, err := c.Version.GetVersion(&version.GetVersionParams{Context: context.Background()})
		cmdx.Must(err, "Could not get version: %s", err)

		clihelpers.PrintRow(cmd, (*versionValue)(resp.Payload))
	},
}
