package identities

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/ory/kratos/cmd/cliclient"
	"github.com/ory/kratos/internal/httpclient/client/admin"
)

var deleteCmd = &cobra.Command{
	Use:  "delete",
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
