package identities

import (
	"fmt"
	"strconv"

	"github.com/ory/x/cmdx"

	"github.com/spf13/cobra"

	"github.com/ory/kratos-client-go/client/admin"
	"github.com/ory/kratos/cmd/cliclient"
)

var ListCmd = &cobra.Command{
	Use:   "list [<page> <per-page>]",
	Short: "List identities",
	Long:  "List identities (paginated)",
	Args: func(cmd *cobra.Command, args []string) error {
		// zero or exactly two args
		if len(args) != 0 && len(args) != 2 {
			return fmt.Errorf("expected zero or two args, got %d: %+v", len(args), args)
		}
		return nil
	},
	Aliases: []string{"ls"},
	RunE: func(cmd *cobra.Command, args []string) error {
		c := cliclient.NewClient(cmd)

		params := &admin.ListIdentitiesParams{
			Context: cmd.Context(),
		}

		if len(args) == 2 {
			page, err := strconv.ParseInt(args[0], 0, 64)
			if err != nil {
				_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Could not parse page argument\"%s\": %s", args[0], err)
				return cmdx.FailSilently(cmd)
			}
			params.Page = &page

			perPage, err := strconv.ParseInt(args[1], 0, 64)
			if err != nil {
				_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Could not parse per-page argument\"%s\": %s", args[1], err)
				return cmdx.FailSilently(cmd)
			}
			params.PerPage = &perPage
		}

		resp, err := c.Admin.ListIdentities(params)
		if err != nil {
			_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Could not get the identities: %+v\n", err)
			return cmdx.FailSilently(cmd)
		}

		cmdx.PrintTable(cmd, &outputIdentityCollection{
			identities: resp.Payload,
		})

		return nil
	},
}
