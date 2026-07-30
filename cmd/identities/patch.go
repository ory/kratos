// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package identities

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ory/kratos/cmd/cliclient"
	"github.com/ory/kratos/internal/clihelpers"
	kratos "github.com/ory/kratos/internal/httpclient"
	"github.com/ory/x/cmdx"
	"github.com/ory/x/flagx"
)

func NewPatchCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "patch",
		Short: "Patch resources",
	}
	cmd.AddCommand(NewPatchIdentityCmd())
	cliclient.RegisterClientFlags(cmd.PersistentFlags())
	cmdx.RegisterFormatFlags(cmd.PersistentFlags())
	return cmd
}

func IsValidOp(op string) bool {
	switch op {
	case
		"add",
		"remove",
		"replace",
		"move",
		"copy",
		"test":
		return true
	}
	return false
}

func NewPatchIdentityCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "identity id-0 [id-1] [id-2] [id-n]",
		Short: "Patch one or more identities by their ID(s)",
		Long: fmt.Sprintf(`This command patches one or more identities by ID. To patch an identity by some selector, e.g. the recovery email address, use the list command in combination with jq.

%s`, clihelpers.WarningJQIsComplicated),
		Example: `To patch the identity with the recovery email address "foo@bar.com", run:

	{{ .CommandPath }} $({{ .Root.Name }} list identities --format json | jq -r 'map(select(.recovery_addresses[].value == "foo@bar.com")) | .[].id')`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			op := flagx.MustGetString(cmd, "op")
			if !IsValidOp(op) {
				cmdx.Fatalf("Not a valid JSON Patch operation. Valid values are: 'add', 'remove', 'replace', 'move', 'copy' and 'test'")
			}
			from := flagx.MustGetString(cmd, "from")
			path := flagx.MustGetString(cmd, "path")
			value := flagx.MustGetString(cmd, "value")
			if value != "" && op != "add" && op != "replace" {
				cmdx.Fatalf("Value is only used in 'add' and 'replace' operations.")
			}

			c, err := cliclient.NewClient(cmd)
			if err != nil {
				return err
			}

			jsonPatch := kratos.JsonPatch{
				Op:   op,
				Path: path,
			}
			if from != "" {
				if op != "move" {
					cmdx.Fatalf("From is only used in 'move' operations.")
				}
				jsonPatch.From = &from
			}
			if value != "" {
				var parsedValue interface{}
				err = json.Unmarshal([]byte(value), &parsedValue)
				if err != nil {
					jsonPatch.Value = value
				} else {
					jsonPatch.Value = parsedValue
				}
			}
			var (
				patched = make([]kratos.Identity, 0, len(args))
				failed  = make(map[string]error)
			)

			for _, a := range args {
				ident, _, err := c.IdentityApi.PatchIdentity(cmd.Context(), a).JsonPatch([]kratos.JsonPatch{jsonPatch}).Execute()
				if err != nil {
					failed[a] = cmdx.PrintOpenAPIError(cmd, err)
					continue
				}
				patched = append(patched, *ident)
			}

			if len(patched) == 1 {
				cmdx.PrintRow(cmd, (*outputIdentity)(&patched[0]))
			} else if len(patched) > 1 {
				cmdx.PrintTable(cmd, &outputIdentityCollection{identities: patched})
			}

			cmdx.PrintErrors(cmd, failed)
			if len(failed) != 0 {
				return cmdx.FailSilently(cmd)
			}

			return nil
		},
	}

	c.Flags().StringP("from", "f", "", "This field is used together with operation 'move' and uses JSON Pointer notation.")
	c.Flags().StringP("op", "o", "", "The JSON Patch operation. Valid values are: 'add', 'remove', 'replace', 'move', 'copy' and 'test'")
	c.Flags().StringP("path", "p", "", "The JSON Pointer to the target key.")
	c.Flags().StringP("value", "v", "", "The value to be used. Only available for 'add' and 'replace' operations..")

	cmdx.Must(c.MarkFlagRequired("op"), "")
	cmdx.Must(c.MarkFlagRequired("path"), "")

	return c
}
