// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package identities

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
	c.AddCommand(NewListIdentitiesCmd())
	cliclient.RegisterClientFlags(c.PersistentFlags())
	cmdx.RegisterFormatFlags(c.PersistentFlags())
	return c
}

func NewListIdentitiesCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "identities [<page> <per-page>]",
		Short:   "List identities",
		Long:    "List identities (paginated)",
		Example: "{{ .CommandPath }} 100 1",
		Args:    cmdx.ZeroOrTwoArgs,
		Aliases: []string{"ls"},
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := cliclient.NewClient(cmd)
			if err != nil {
				return err
			}

			req := c.IdentityApi.ListIdentities(cmd.Context())
			if len(args) == 2 {
				page, perPage, err := cmdx.ParsePaginationArgs(cmd, args[0], args[1])
				if err != nil {
					return err
				}

				req = req.Page(page)
				req = req.PerPage(perPage)
			}

			identities, _, err := req.Execute()
			if err != nil {
				return cmdx.PrintOpenAPIError(cmd, err)
			}

			cmdx.PrintTable(cmd, &outputIdentityCollection{identities: identities})
			return nil
		},
	}
}
