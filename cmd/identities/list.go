package identities

import (
	"fmt"
	"strconv"

	"github.com/ory/x/cmdx"

	"github.com/spf13/cobra"

	"github.com/ory/kratos/cmd/cliclient"
)

func NewListCmd(root *cobra.Command) *cobra.Command {
	c := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List resources",
	}
	c.AddCommand(NewListIdentitiesCmd(root))
	cliclient.RegisterClientFlags(c.PersistentFlags())
	cmdx.RegisterFormatFlags(c.PersistentFlags())
	return c
}

func NewListIdentitiesCmd(root *cobra.Command) *cobra.Command {
	return &cobra.Command{
		Use:     "identities [<page> <per-page>]",
		Short:   "List identities",
		Long:    "List identities (paginated)",
		Example: fmt.Sprintf("%[1]s ls identities 100 1", root.Use),
		Args: func(cmd *cobra.Command, args []string) error {
			// zero or exactly two args
			if len(args) != 0 && len(args) != 2 {
				return fmt.Errorf("expected zero or two args, got %d: %+v", len(args), args)
			}
			return nil
		},
		Aliases: []string{"ls"},
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := cliclient.NewClient(cmd)
			if err != nil {
				return err
			}

			req := c.V0alpha2Api.AdminListIdentities(cmd.Context())

			if len(args) == 2 {
				page, err := strconv.ParseInt(args[0], 0, 64)
				if err != nil {
					_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Could not parse page argument\"%s\": %s", args[0], err)
					return cmdx.FailSilently(cmd)
				}
				req = req.Page(page)

				perPage, err := strconv.ParseInt(args[1], 0, 64)
				if err != nil {
					_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Could not parse per-page argument\"%s\": %s", args[1], err)
					return cmdx.FailSilently(cmd)
				}
				req = req.PerPage(perPage)
			}

			identities, _, err := req.Execute()
			if err != nil {
				return cmdx.PrintOpenAPIError(cmd, err)
			}

			cmdx.PrintTable(cmd, &outputIdentityCollection{
				identities: identities,
			})

			return nil
		},
	}
}
