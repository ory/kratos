package client

import (
	"encoding/json"
	"os"

	"github.com/spf13/cobra"

	"github.com/ory/hive/sdk/go/hive/client/admin"
	"github.com/ory/hive/sdk/go/hive/models"
	"github.com/ory/x/cmdx"
)

type IdentityClient struct{}

func NewIdentityClient() *IdentityClient {
	return new(IdentityClient)
}

func (ic *IdentityClient) Import(cmd *cobra.Command, args []string) {
	cmdx.MinArgs(cmd, args, 1)

	for _, p := range args {
		var is []models.Identity
		f, err := os.Open(p)
		cmdx.Must(err, "Unable to open file %s: %s", p, err)
		d := json.NewDecoder(f)
		d.DisallowUnknownFields()
		err = d.Decode(&is)
		cmdx.Must(err, "Unable to decode file %s to JSON: %s", p, err)

		for _, i := range is {
			_, err := c(cmd, "endpoint").Admin.UpsertIdentity(
				admin.NewUpsertIdentityParams().WithBody(&i),
			)
			cmdx.Must(err, "Unable to import identity: %s", err)
		}
	}
}

func (ic *IdentityClient) List(cmd *cobra.Command, args []string) {
	cmdx.ExactArgs(cmd, args, 2)
	//ic.c.Admin

}

func (ic *IdentityClient) Get(cmd *cobra.Command, args []string) {

}
