package identities

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/spf13/cobra"

	"github.com/ory/kratos/internal/clihelpers"

	"github.com/ory/kratos/cmd/cliclient"
	"github.com/ory/kratos/internal/httpclient/client/admin"
	"github.com/ory/kratos/internal/httpclient/models"
	"github.com/ory/x/cmdx"
)

// putCmd represents the import command
var putCmd = &cobra.Command{
	Use: "import <file.json [file-2.json [file-3.json] ...]>",
	Short: "Import identities from files or STD_IN",
	Long: "Import identities from files or STD_IN. Files are expected to each contain a single identity. The validity of files can be tested beforehand using the `validate` command.",
	Run: importIdentities,
}

func importIdentities(cmd *cobra.Command, args []string) {
	c := cliclient.NewClient(cmd)

	imported := &outputIdentityCollection{
		identities: make([]*models.Identity, 0, len(args)),
	}
	failed := make(map[string]error)

	if len(args) == 0 {
		fc, err := ioutil.ReadAll(os.Stdin)
		cmdx.Must(err, "Could not read from STD_IN: %s", err)

		validateIdentity("STD_IN", fc, c)

		var params models.CreateIdentity
		err = json.NewDecoder(bytes.NewBuffer(fc)).Decode(&params)
		cmdx.Must(err, "STD_IN: Could not parse identity: %s", err)

		resp, err := c.Admin.CreateIdentity(&admin.CreateIdentityParams{
			Body:    &params,
			Context: context.Background(),
		})

		if err != nil {
			failed["STD_IN"] = err
		} else {
			imported.identities = append(imported.identities, resp.Payload)
		}
	} else {
		for _, fn := range args {
			fc := validateIdentityFile(fn, c)

			var params models.CreateIdentity
			err := json.NewDecoder(bytes.NewBuffer(fc)).Decode(&params)
			cmdx.Must(err, "%s: Could not parse identity file: %s", fn, err)

			resp, err := c.Admin.CreateIdentity(&admin.CreateIdentityParams{
				Body:    &params,
				Context: context.Background(),
			})

			if err != nil {
				failed[fn] = err
				continue
			}

			imported.identities = append(imported.identities, resp.Payload)
		}
	}

	clihelpers.PrintCollection(cmd, imported)
	clihelpers.PrintErrors(cmd, failed)

	if len(failed) != 0 {
		os.Exit(1)
	}
}
