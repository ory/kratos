package identities

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/ory/kratos/internal/httpclient/client/public"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestValidateIdentity(t *testing.T) {
	var testCmd = func() (*cobra.Command, *bytes.Buffer, *bytes.Buffer) {
		out, err := &bytes.Buffer{}, &bytes.Buffer{}
		cmd := &cobra.Command{}
		cmd.SetOut(out)
		cmd.SetErr(err)
		return cmd, out, err
	}

	var testSchemaGetter = func(resp interface{}) schemaGetter {
		return func(_ *public.GetSchemaParams) (*public.GetSchemaOK, error) {
			return &public.GetSchemaOK{Payload: resp}, nil
		}
	}

	for i, tc := range []struct {
		description           string
		payload               string
		expectedRequiredError string
		identitySchema        map[string]interface{}
	}{
		{
			description:           "requires schema_id from swagger schema",
			payload:               "{}",
			expectedRequiredError: "schema_id",
		},
		{
			description:           "requires traits from swagger schema",
			payload:               `{"schema_id": "case2"}`,
			expectedRequiredError: "traits",
		},
		{
			description:           "requires custom trait key",
			payload:               `{"schema_id": "case3", "traits": {}}`,
			expectedRequiredError: "random-property",
			identitySchema: map[string]interface{}{
				"type":     "object",
				"required": []string{"random-property"},
				"properties": map[string]interface{}{
					"random-property": map[string]interface{}{
						"type": "string",
					},
				},
			},
		},
	} {
		t.Run(fmt.Sprintf("case=%d/description=%s", i, tc.description), func(t *testing.T) {
			cmd, stdOut, stdErr := testCmd()

			if tc.identitySchema == nil {
				tc.identitySchema = map[string]interface{}{}
			}
			err := validateIdentity(cmd, "test identity", tc.payload, testSchemaGetter(tc.identitySchema))
			assert.Error(t, err, stdOut.String(), stdErr.String())
			assert.Len(t, stdOut.String(), 0, stdErr.String())
			assert.Contains(t, stdErr.String(), "required", stdOut.String())
			assert.Contains(t, stdErr.String(), tc.expectedRequiredError, stdOut.String())
		})
	}
}
