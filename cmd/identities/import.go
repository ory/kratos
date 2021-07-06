package identities

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	kratos "github.com/ory/kratos-client-go"

	"github.com/ory/x/cmdx"

	"github.com/spf13/cobra"

	"github.com/ory/kratos/cmd/cliclient"
)

// NewImportCmd represents the import command
func NewImportCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "import <file.json [file-2.json [file-3.json] ...]>",
		Short: "Import identities from files or STD_IN",
		Example: `$ cat > ./file.json <<EOF
{
    "schema_id": "default",
    "traits": {
        "email": "foo@example.com"
    }
}
EOF

$ kratos identities import file.json
# Alternatively:
$ cat file.json | kratos identities import`,
		Long: `Import identities from files or STD_IN.

Files can contain only a single or an array of identities. The validity of files can be tested beforehand using "... identities validate".

WARNING: Importing credentials is not yet supported.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			c := cliclient.NewClient(cmd)

			imported := make([]kratos.Identity, 0, len(args))
			failed := make(map[string]error)

			is, err := readIdentities(cmd, args)
			if err != nil {
				return err
			}

			for src, i := range is {
				err = ValidateIdentity(cmd, src, i, func(ctx context.Context, id string) (map[string]interface{}, *http.Response, error) {
					return c.PublicApi.GetSchema(ctx, id).Execute()
				})
				if err != nil {
					return err
				}

				var params kratos.AdminCreateIdentityBody
				err = json.Unmarshal([]byte(i), &params)
				if err != nil {
					_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "STD_IN: Could not parse identity")
					return cmdx.FailSilently(cmd)
				}

				ident, _, err := c.V0alpha1Api.AdminCreateIdentity(cmd.Context()).AdminCreateIdentityBody(params).Execute()
				if err != nil {
					failed[src] = err
				} else {
					imported = append(imported, *ident)
				}
			}
			if len(imported) == 1 {
				cmdx.PrintRow(cmd, (*outputIdentity)(&imported[0]))
			} else {
				cmdx.PrintTable(cmd, &outputIdentityCollection{identities: imported})
			}
			cmdx.PrintErrors(cmd, failed)

			if len(failed) != 0 {
				return cmdx.FailSilently(cmd)
			}

			return nil
		},
	}
}
