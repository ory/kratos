// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package identities

import (
	"encoding/json"
	"fmt"

	kratos "github.com/ory/kratos/internal/httpclient"

	"github.com/ory/x/cmdx"

	"github.com/spf13/cobra"

	"github.com/ory/kratos/cmd/cliclient"
)

func NewImportCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "import",
		Short: "Import resources",
	}
	cmd.AddCommand(NewImportIdentitiesCmd())
	cliclient.RegisterClientFlags(cmd.PersistentFlags())
	cmdx.RegisterFormatFlags(cmd.PersistentFlags())
	return cmd
}

// NewImportIdentitiesCmd represents the import command
func NewImportIdentitiesCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "identities file-1.json [file-2.json] [file-3.json] [file-n.json]",
		Short: "Import one or more identities from files or STD_IN",
		Example: `Create an example identity:

	cat > ./file.json <<EOF
	{
	    "schema_id": "default",
	    "traits": {
	        "email": "foo@example.com"
	    }
	}
	EOF

	{{ .CommandPath }} file.json

Alternatively:

	cat file.json | {{ .CommandPath }}`,
		Long: `Import identities from files or STD_IN.

Files can contain only a single or an array of identities. The validity of files can be tested beforehand using "... identities validate".`,
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := cliclient.NewClient(cmd)
			if err != nil {
				return err
			}

			imported := make([]kratos.Identity, 0, len(args))
			failed := make(map[string]error)

			is, err := readIdentities(cmd, args)
			if err != nil {
				return err
			}

			for src, i := range is {
				var params kratos.CreateIdentityBody
				err = json.Unmarshal([]byte(i), &params)
				if err != nil {
					_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "STD_IN: Could not parse identity")
					return cmdx.FailSilently(cmd)
				}

				ident, _, err := c.IdentityApi.CreateIdentity(cmd.Context()).CreateIdentityBody(params).Execute()
				if err != nil {
					failed[src] = cmdx.PrintOpenAPIError(cmd, err)
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
