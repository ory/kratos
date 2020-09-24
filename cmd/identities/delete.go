package identities

import (
	"context"
	"fmt"
	"os"

	"github.com/ory/kratos/internal/clihelpers"

	"github.com/spf13/cobra"

	"github.com/ory/kratos/cmd/cliclient"
	"github.com/ory/kratos/internal/httpclient/client/admin"
)

var deleteCmd = &cobra.Command{
	Use:   "delete <id-0 [id-1 ...]>",
	Short: "Delete identities by ID",
	Long: fmt.Sprintf(`This command deletes one or more identities by ID. To delete an identity by some selector, e.g. the recovery email address, use the list command in combination with jq.
Example: delete the identity with the recovery email address "foo@bar.com":

kratos identities delete $(kratos identities list --format json | jq -r 'map(select(.recovery_addresses[].value == "foo@bar.com")) | .[].id')

%s
`, clihelpers.WarningJQIsComplicated),
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		c := cliclient.NewClient(cmd)

		var (
			deleted = make([]string, 0, len(args))
			errs    []error
		)

		for _, a := range args {
			_, err := c.Admin.DeleteIdentity(&admin.DeleteIdentityParams{
				ID:      a,
				Context: context.Background(),
			})
			if err != nil {
				errs = append(errs, err)
				continue
			}
			deleted = append(deleted, a)
		}

		for _, d := range deleted {
			fmt.Println(d)
		}

		for _, err := range errs {
			fmt.Fprintln(os.Stderr, err)
		}

		if len(errs) != 0 {
			os.Exit(1)
		}
	},
}
