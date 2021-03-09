package identities

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ory/kratos/cmd/cliclient"
	"github.com/ory/kratos/internal/clihelpers"
	"github.com/ory/x/cmdx"
)

var DeleteCmd = &cobra.Command{
	Use:   "delete <id-0 [id-1 ...]>",
	Short: "Delete identities by ID",
	Long: fmt.Sprintf(`This command deletes one or more identities by ID. To delete an identity by some selector, e.g. the recovery email address, use the list command in combination with jq.

%s
`, clihelpers.WarningJQIsComplicated),
	Example: `To delete the identity with the recovery email address "foo@bar.com", run:

	$ kratos identities delete $(kratos identities list --format json | jq -r 'map(select(.recovery_addresses[].value == "foo@bar.com")) | .[].id')`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c := cliclient.NewClient(cmd)

		var (
			deleted = make([]string, 0, len(args))
			errs    []error
		)

		for _, a := range args {
			// TODO merge
			// _, err := c.Admin.DeleteIdentity(admin.NewDeleteIdentityParams().WithID(a).WithTimeout(time.Second).WithHTTPClient(cliclient.NewHTTPClient(cmd)))
			_, err := c.AdminApi.DeleteIdentity(cmd.Context(), a).Execute()
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
