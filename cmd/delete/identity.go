package delete

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ory/x/cmdx"

	"github.com/ory/kratos/cmd/cliclient"
	"github.com/ory/kratos/internal/clihelpers"
)

func NewDeleteIdentityCmd(parent *cobra.Command) *cobra.Command {
	return &cobra.Command{
		Use:     "identity id-0 [id-1] [id-2] [id-n]",
		Aliases: []string{"identities"},
		Short:   "Delete one or more identities by their ID(s)",
		Long: fmt.Sprintf(`This command deletes one or more identities by ID. To delete an identity by some selector, e.g. the recovery email address, use the list command in combination with jq.

%s`, clihelpers.WarningJQIsComplicated),
		Example: fmt.Sprintf(`To delete the identity with the recovery email address "foo@bar.com", run:

	%s identity $(%s list identities --format json | jq -r 'map(select(.recovery_addresses[].value == "foo@bar.com")) | .[].id')`, parent.Use, parent.Root().Use),
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c := cliclient.NewClient(cmd)

			var (
				deleted = make([]string, 0, len(args))
				errs    []error
			)

			for _, a := range args {
				_, err := c.V0alpha2Api.AdminDeleteIdentity(cmd.Context(), a).Execute()
				if err != nil {
					errs = append(errs, err)
					continue
				}
				deleted = append(deleted, a)
			}

			for _, d := range deleted {
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), d)
			}

			for _, err := range errs {
				_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "%+v\n", err)
			}

			if len(errs) != 0 {
				return cmdx.FailSilently(cmd)
			}
			return nil
		},
	}
}
