// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package identities

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ory/kratos/cmd/cliclient"
	"github.com/ory/kratos/pkg/clihelpers"
	"github.com/ory/x/cmdx"
)

func NewDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete resources",
	}
	cmd.AddCommand(NewDeleteIdentityCmd())
	cliclient.RegisterClientFlags(cmd.PersistentFlags())
	cmdx.RegisterFormatFlags(cmd.PersistentFlags())
	return cmd
}

func NewDeleteIdentityCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "identity <id> [<id> ...]",
		Short: "Delete one or more identities by their ID(s)",
		Long: fmt.Sprintf(`This command deletes one or more identities by ID. To delete an identity by some selector, e.g. the recovery email address, use the list command in combination with jq.

%s`, clihelpers.WarningJQIsComplicated),
		Example: `To delete the identity with the recovery email address "foo@bar.com", run:

	{{ .CommandPath }} $({{ .Root.Name }} list identities --format json | jq -r 'map(select(.recovery_addresses[].value == "foo@bar.com")) | .[].id')`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := cliclient.NewClient(cmd)
			if err != nil {
				return err
			}

			var (
				deleted = make([]cmdx.OutputIder, 0, len(args))
				failed  []string
			)

			for _, a := range args {
				_, err := c.IdentityAPI.DeleteIdentity(cmd.Context(), a).Execute()
				if err != nil {
					_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Failed to delete identity %s:\n", a)
					// PrintOpenAPIError prints the message itself and returns the
					// ErrNoPrintButFail sentinel when it could parse an API error body.
					// For transport-level errors (no body, e.g. timeouts) it prints
					// nothing and returns the original error instead, which we must
					// still surface here.
					if printErr := cmdx.PrintOpenAPIError(cmd, err); !errors.Is(printErr, cmdx.ErrNoPrintButFail) {
						_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "%s\n", printErr)
					}
					failed = append(failed, a)
					continue
				}
				deleted = append(deleted, cmdx.OutputIder(a))
			}

			if len(deleted) == 1 {
				cmdx.PrintRow(cmd, &deleted[0])
			} else if len(deleted) > 1 {
				cmdx.PrintTable(cmd, &cmdx.OutputIderCollection{Items: deleted})
			}

			if len(failed) != 0 {
				return cmdx.FailSilently(cmd)
			}

			return nil
		},
	}
}
