// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package identities

import (
	"github.com/spf13/cobra"

	"github.com/ory/x/pagination/keysetpagination"

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
	c := &cobra.Command{
		Use:     "identities",
		Short:   "List identities",
		Long:    "Return a list of identities.",
		Example: "{{ .CommandPath }} --page-size 100",
		Args:    cmdx.ZeroOrTwoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := cliclient.NewClient(cmd)
			if err != nil {
				return err
			}

			req := c.IdentityApi.ListIdentities(cmd.Context())
			page, perPage, err := cmdx.ParseTokenPaginationArgs(cmd)
			if err != nil {
				return err
			}

			req = req.PageToken(page)
			req = req.PageSize(int64(perPage))

			identities, res, err := req.Execute()
			if err != nil {
				return cmdx.PrintOpenAPIError(cmd, err)
			}

			pages := keysetpagination.ParseHeader(res)
			cmdx.PrintTable(cmd, &outputIdentityCollection{
				Identities:       identities,
				NextPageToken:    pages.NextToken,
				includePageToken: true,
			})
			return nil
		},
	}
	cmdx.RegisterTokenPaginationFlags(c)
	return c
}
