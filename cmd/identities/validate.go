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
		c := cliclient.NewClient(cmd)

		for _, fn := range args {
			validateIdentityFile(fn, c)
		}

		fmt.Println("All identity files are valid.")
	},
}

func validateIdentityFile(fn string, c *client.OryKratos) []byte {
	fc, err := ioutil.ReadFile(fn)
	cmdx.Must(err, `%s: Could not open identity file: %s`, fn, err)

	validateIdentity(fn, fc, c)
	return fc
}

func validateIdentity(src string, fc []byte, c *client.OryKratos) {
	swaggerSchema, err := pkger.Open("/.schema/api.swagger.json")
	schemaCompiler := jsonschema.NewCompiler()
	err = schemaCompiler.AddResource("api.swagger.json", swaggerSchema)
	cmdx.Must(err, "Could not get swagger schema: %s", err)

	sid := gjson.GetBytes(fc, "schema_id")
	if !sid.Exists() {
		_, _ = fmt.Fprintf(os.Stderr, "%s: Expected key \"schema_id\" to be defined.\n", src)
		os.Exit(1)
	}

	cmdx.Must(err, "%s: Could not open the identity schema: %s", src, err)

	s, err := schemaCompiler.Compile("api.swagger.json#/definitions/CreateIdentityRequestPayload")
	cmdx.Must(err, "%s: Could not compile the identity schema: %s", src, err)
	// force additional properties to false because swagger does not render this key
	s.AdditionalProperties = false

	var foundValidationErrors bool
	err = s.Validate(bytes.NewBuffer(fc))
	if err != nil {
		fmt.Printf("%s: not valid\n", src)
		viperx.PrintHumanReadableValidationErrors(os.Stderr, err)
		foundValidationErrors = true
	}

	ts, err := c.Common.GetSchema(&common.GetSchemaParams{ID: sid.String(), Context: context.Background()})
	cmdx.Must(err, `%s: Could not fetch schema with ID "%s": %s`, src, sid.String(), err)

	traitsSchema, err := json.Marshal(ts.Payload)
	cmdx.Must(err, "%s: Could not marshal the traits schema: %s", src, err)

	s, err = jsonschema.CompileString("identity_traits.schema.json", string(traitsSchema))
	cmdx.Must(err, "%s: Could not compile the traits schema: %s", src, err)

	err = s.Validate(bytes.NewBuffer(fc))
	if err != nil {
		fmt.Printf("%s: not valid\n", src)
		viperx.PrintHumanReadableValidationErrors(os.Stderr, err)
		foundValidationErrors = true
	}

	if foundValidationErrors {
		os.Exit(1)
	}
}
