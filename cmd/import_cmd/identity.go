package import_cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/spf13/cobra"

	kratos "github.com/ory/kratos-client-go"
	"github.com/ory/kratos/cmd/cliclient"
	"github.com/ory/kratos/cmd/definitions"
	"github.com/ory/kratos/cmd/validate"
	"github.com/ory/kratos/internal/clihelpers"
	"github.com/ory/x/cmdx"
)

func NewImportIdentityCmd(parent *cobra.Command) *cobra.Command {
	return &cobra.Command{
		Use:     "identity file-1.json [file-2.json] [file-3.json] [file-n.json]",
		Aliases: []string{"identities"},
		Short:   "Import one or more identities from files or STD_IN",
		Example: fmt.Sprintf(`Create an example identity:

	cat > ./file.json <<EOF
	{
	    "schema_id": "default",
	    "traits": {
	        "email": "foo@example.com"
	    }
	}
	EOF

	%[1]s identity file.json

Alternatively:

	cat file.json | %[1]s identity`, parent.Use),
		Long: `Import identities from files or STD_IN.

Files can contain only a single or an array of identities. The validity of files can be tested beforehand using "... identities validate".

WARNING: Importing credentials is not yet supported.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			c := cliclient.NewClient(cmd)

			imported := make([]kratos.Identity, 0, len(args))
			failed := make(map[string]error)

			is, err := clihelpers.ReadIdentities(cmd, args)
			if err != nil {
				return err
			}

			for src, i := range is {
				err = validate.ValidateIdentity(cmd, src, i, func(ctx context.Context, id string) (map[string]interface{}, *http.Response, error) {
					return c.V0alpha2Api.GetJsonSchema(ctx, id).Execute()
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

				ident, _, err := c.V0alpha2Api.AdminCreateIdentity(cmd.Context()).AdminCreateIdentityBody(params).Execute()
				if err != nil {
					failed[src] = err
				} else {
					imported = append(imported, *ident)
				}
			}
			if len(imported) == 1 {
				cmdx.PrintRow(cmd, (*definitions.OutputIdentity)(&imported[0]))
			} else {
				cmdx.PrintTable(cmd, &definitions.OutputIdentityCollection{Identities: imported})
			}
			cmdx.PrintErrors(cmd, failed)

			if len(failed) != 0 {
				return cmdx.FailSilently(cmd)
			}

			return nil
		},
	}
}
