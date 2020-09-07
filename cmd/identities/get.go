package identities

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/ory/kratos/cmd/cliclient"
	"github.com/ory/kratos/internal/clihelpers"
	"github.com/ory/kratos/internal/httpclient/client/admin"
	"github.com/ory/x/cmdx"
)

var getCmd = &cobra.Command{
	Use:  "get <id>",
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		c := cliclient.NewClient(cmd)

		resp, err := c.Admin.GetIdentity(&admin.GetIdentityParams{
			ID:      args[0],
			Context: context.Background(),
		})
		cmdx.Must(err, "Could not get identity \"%s\": %s", args[0], err)

		clihelpers.PrintRow(cmd, (*outputIdentity)(resp.Payload))
	},
}
