package identities

import (
	"context"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/ory/kratos/cmd/cliclient"
	"github.com/ory/kratos/internal/httpclient/client/admin"
	"github.com/ory/x/cmdx"
)

var listCmd = &cobra.Command{
	Use: "list [<page> <per-page>]",
	Args: func(cmd *cobra.Command, args []string) error {
		// zero or exactly two args
		if len(args) != 0 && len(args) != 2 {
			return fmt.Errorf("expected zero or two args, got %d", len(args))
		}
		return nil
	},
	Aliases: []string{"ls"},
	Run: func(cmd *cobra.Command, args []string) {
		c := cliclient.NewClient()

		params := &admin.ListIdentitiesParams{
			Context: context.Background(),
		}

		if len(args) == 2 {
			page, err := strconv.ParseInt(args[0], 0, 64)
			cmdx.Must(err, "Could not parse page argument\"%s\": %s", args[0], err)
			params.Page = &page

			perPage, err := strconv.ParseInt(args[1], 0, 64)
			cmdx.Must(err, "Could not parse per-page argument\"%s\": %s", args[1], err)
			params.PerPage = &perPage
		}

		resp, err := c.Admin.ListIdentities(params)
		cmdx.Must(err, "Could not get the identities: %s", err)

		for _, i := range resp.Payload {
			fmt.Println(i.ID)
		}
	},
}
