// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package identities

import (
	"fmt"

	kratos "github.com/ory/kratos/internal/httpclient"
	"github.com/ory/kratos/x"
	"github.com/ory/x/cmdx"
	"github.com/ory/x/stringsx"

	"github.com/ory/kratos/internal/clihelpers"

	"github.com/spf13/cobra"

	"github.com/ory/kratos/cmd/cliclient"
)

const (
	FlagIncludeCreds = "include-credentials"
)

func NewGetCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "get",
		Short: "Get resources",
	}
	cmd.AddCommand(NewGetIdentityCmd())
	cliclient.RegisterClientFlags(cmd.PersistentFlags())
	cmdx.RegisterFormatFlags(cmd.PersistentFlags())
	return cmd
}

func NewGetIdentityCmd() *cobra.Command {
	var (
		includeCreds []string
	)

	cmd := &cobra.Command{
		Use:   "identity [id-1] [id-2] [id-n]",
		Short: "Get one or more identities by their ID(s)",
		Long: fmt.Sprintf(`This command gets all the details about an identity. To get an identity by some selector, e.g. the recovery email address, use the list command in combination with jq.

%s`, clihelpers.WarningJQIsComplicated),
		Example: `To get the identities with the recovery email address at the domain "ory.sh", run:

	{{ .CommandPath }} $({{ .Root.Name }} ls identities --format json | jq -r 'map(select(.recovery_addresses[].value | endswith("@ory.sh"))) | .[].id')`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := cliclient.NewClient(cmd)
			if err != nil {
				return err
			}

			// we check includeCreds argument is valid
			for _, opt := range includeCreds {
				e := stringsx.SwitchExact(opt)
				if !e.AddCase("oidc") {
					cmd.PrintErrln(`You have to put a valid value of credentials type to be included, try --help for details.`)
					return cmdx.FailSilently(cmd)
				}
			}

			identities := make([]kratos.Identity, 0, len(args))
			failed := make(map[string]error)
			for _, id := range args {
				identity, _, err := c.IdentityApi.
					GetIdentity(cmd.Context(), id).
					IncludeCredential(includeCreds).
					Execute()

				if x.SDKError(err) != nil {
					failed[id] = cmdx.PrintOpenAPIError(cmd, err)
					continue
				}

				identities = append(identities, *identity)
			}

			if len(identities) == 1 {
				cmdx.PrintRow(cmd, (*outputIdentity)(&identities[0]))
			} else if len(identities) > 1 {
				cmdx.PrintTable(cmd, &outputIdentityCollection{identities})
			}
			cmdx.PrintErrors(cmd, failed)

			if len(failed) != 0 {
				return cmdx.FailSilently(cmd)
			}
			return nil
		},
	}

	flags := cmd.Flags()
	// include credential flag to add third party tokens in returned data
	flags.StringArrayVarP(&includeCreds, FlagIncludeCreds, "i", []string{}, `Include third party tokens (only "oidc" supported) `)
	return cmd
}
