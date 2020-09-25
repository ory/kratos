package identities

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/spf13/cobra"

	"github.com/ory/kratos/internal/clihelpers"

	"github.com/ory/kratos/cmd/cliclient"
	"github.com/ory/kratos/internal/httpclient/client/admin"
	"github.com/ory/kratos/internal/httpclient/models"
)

// importCmd represents the import command
var importCmd = &cobra.Command{
	Use:   "import <file.json [file-2.json [file-3.json] ...]>",
	Short: "import identities from files or STD_IN",
	Long:  "Import identities from files or STD_IN. Files are expected to each contain a single identity. The validity of files can be tested beforehand using `... identities validate`. Importing credentials is not yet supported.",
	RunE: func(cmd *cobra.Command, args []string) error {
		c := cliclient.NewClient(cmd)

		imported := make([]*models.Identity, 0, len(args))
		failed := make(map[string]error)

		if len(args) == 0 {
			fc, err := ioutil.ReadAll(cmd.InOrStdin())
			if err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "STD_IN: Could not read: %s\n", err)
				return clihelpers.FailSilently(cmd)
			}

			err = validateIdentity(cmd, "STD_IN", fc, c.Common.GetSchema)
			if err != nil {
				return err
			}

			var params models.CreateIdentity
			err = json.NewDecoder(bytes.NewBuffer(fc)).Decode(&params)
			if err != nil {
				fmt.Fprintln(cmd.ErrOrStderr(), "STD_IN: Could not parse identity")
				return clihelpers.FailSilently(cmd)
			}

			resp, err := c.Admin.CreateIdentity(&admin.CreateIdentityParams{
				Body:    &params,
				Context: context.Background(),
			})

			if err != nil {
				failed["STD_IN"] = err
			} else {
				imported = append(imported, resp.Payload)
			}
		} else {
			for _, fn := range args {
				fc, err := validateIdentityFile(cmd, fn, c)
				if err != nil {
					return err
				}

				var params models.CreateIdentity
				err = json.Unmarshal(fc, &params)
				if err != nil {
					fmt.Fprintf(cmd.ErrOrStderr(), "%s: Could not parse identity file\n", fn)
					return clihelpers.FailSilently(cmd)
				}

				resp, err := c.Admin.CreateIdentity(
					admin.NewCreateIdentityParams().
						WithBody(&params).
						WithTimeout(time.Second))

				if err != nil {
					failed[fn] = err
					continue
				}

				imported = append(imported, resp.Payload)
			}
		}
		if len(imported) == 1 {
			clihelpers.PrintRow(cmd, (*outputIdentity)(imported[0]))
		} else {
			clihelpers.PrintCollection(cmd, &outputIdentityCollection{identities: imported})
		}
		clihelpers.PrintErrors(cmd, failed)

		if len(failed) != 0 {
			return clihelpers.FailSilently(cmd)
		}

		return nil
	},
}
