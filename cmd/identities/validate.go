package identities

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/markbates/pkger"
	"github.com/spf13/cobra"
	"github.com/tidwall/gjson"

	"github.com/ory/jsonschema/v3"
	"github.com/ory/kratos/cmd/cliclient"
	"github.com/ory/kratos/internal/httpclient/client"
	"github.com/ory/kratos/internal/httpclient/client/common"
	"github.com/ory/x/cmdx"
	"github.com/ory/x/viperx"
)

var validateCmd = &cobra.Command{
	Use:  "validate  <file.json [file-2.json [file-3.json] ...]>",
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		c := cliclient.NewClient()

		for _, fn := range args {
			validateIdentityFile(fn, c)
		}

		fmt.Println("All identity files are valid.")
	},
}

func validateIdentityFile(fn string, c *client.OryKratos) []byte {
	fc, err := ioutil.ReadFile(fn)
	cmdx.Must(err, `%s: Could not open identity file: %s`, fn, err)

	sid := gjson.GetBytes(fc, "schema_id")
	if !sid.Exists() {
		_, _ = fmt.Fprintf(os.Stderr, "%s: Expected key \"schema_id\" to be defined.\n", fn)
		os.Exit(1)
	}

	f, err := pkger.Open("/.schema/identity.schema.json")
	cmdx.Must(err, "%s: Could not open the identity schema: %s", fn, err)

	identitySchema, err := ioutil.ReadAll(f)
	cmdx.Must(err, "%s: Could not read the identity schema: %s", fn, err)

	s, err := jsonschema.CompileString("identity.schema.json", string(identitySchema))
	cmdx.Must(err, "%s: Could not compile the identity schema: %s", fn, err)

	var foundValidationErrors bool
	err = s.Validate(bytes.NewBuffer(fc))
	if err != nil {
		fmt.Printf("%s: not valid\n", fn)
		viperx.PrintHumanReadableValidationErrors(os.Stderr, err)
		foundValidationErrors = true
	}

	ts, err := c.Common.GetSchema(&common.GetSchemaParams{ID: sid.String(), Context: context.Background()})
	cmdx.Must(err, `%s: Could not fetch schema with ID "%s": %s`, fn, sid.String(), err)

	traitsSchema, err := json.Marshal(ts.Payload)
	cmdx.Must(err, "%s: Could not marshal the traits schema: %s", fn, err)

	s, err = jsonschema.CompileString("identity_traits.schema.json", string(traitsSchema))
	cmdx.Must(err, "%s: Could not compile the traits schema: %s", fn, err)

	err = s.Validate(bytes.NewBuffer(fc))
	if err != nil {
		fmt.Printf("%s: not valid\n", fn)
		viperx.PrintHumanReadableValidationErrors(os.Stderr, err)
		foundValidationErrors = true
	}

	if foundValidationErrors {
		os.Exit(1)
	}

	return fc
}
