package identities

import (
	"bytes"
	"context"
	"encoding/json"

	"github.com/spf13/cobra"

	"github.com/ory/kratos/cmd/cliclient"
	"github.com/ory/kratos/internal/httpclient/client/admin"
	"github.com/ory/kratos/internal/httpclient/models"
	"github.com/ory/x/cmdx"
)

// importCmd represents the import command
var importCmd = &cobra.Command{
	Use:  "import <file.json [file-2.json [file-3.json] ...]>",
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		c := cliclient.NewClient()

		for _, fn := range args {
			fc := validateIdentityFile(fn, c)

			// because the model.Identity does not come with the Marshal/Unmarshal
			// methods of identity.Identity, we have to work around them
			//traits := gjson.GetBytes(fc, "traits")
			//schemaID := gjson.GetBytes(fc, "schema_id")
			//fc, err := sjson.DeleteBytes(fc, "traits")
			//cmdx.Must(err, "%s: Unexpected error removing traits: %s", fn, err)
			//fc, err = sjson.DeleteBytes(fc, "schema_id")
			//cmdx.Must(err, "%s: Unexpected error removing schema_id: %s", fn, err)

			var i models.Identity
			err := json.NewDecoder(bytes.NewBuffer(fc)).Decode(&i)
			cmdx.Must(err, "%s: Could not parse identity file: %s", fn, fc)

			//i.Traits = traits
			//i.SchemaID = pointerx.String(schemaID.String())

			_, err = c.Admin.CreateIdentity(&admin.CreateIdentityParams{
				Body:    &i,
				Context: context.Background(),
			})
			cmdx.Must(err, "%s: Could not create identity: %s", fn, err)
		}
	},
}
