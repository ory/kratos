package identities

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ory/kratos/cmd/cliclient"
	"github.com/ory/kratos/internal/clihelpers"
	"github.com/ory/x/cmdx"
)

func NewDeleteCmd(root *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete resources",
	}
	cmd.AddCommand(NewDeleteIdentityCmd(root))
	cliclient.RegisterClientFlags(cmd.PersistentFlags())
	cmdx.RegisterFormatFlags(cmd.PersistentFlags())
	return cmd
}

func NewDeleteIdentityCmd(root *cobra.Command) *cobra.Command {
	return &cobra.Command{
		Use:   "identity id-0 [id-1] [id-2] [id-n]",
		Short: "Delete one or more identities by their ID(s)",
		Long: fmt.Sprintf(`This command deletes one or more identities by ID. To delete an identity by some selector, e.g. the recovery email address, use the list command in combination with jq.

%s`, clihelpers.WarningJQIsComplicated),
		Example: fmt.Sprintf(`To delete the identity with the recovery email address "foo@bar.com", run:

	%[1]s delete identity $(%[1]s list identities --format json | jq -r 'map(select(.recovery_addresses[].value == "foo@bar.com")) | .[].id')`, root.Use),
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
