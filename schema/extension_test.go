package schema

import (
	"bytes"
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/jsonschema/v3"
	_ "github.com/ory/jsonschema/v3/fileloader"
)

type extensionStub struct {
	identifiers  []string
	accountNames []string
}

func (r *extensionStub) Run(ctx jsonschema.ValidationContext, config ExtensionConfig, value interface{}) error {
	if config.Credentials.Password.Identifier {
		r.identifiers = append(r.identifiers, fmt.Sprintf("%s", value))
	}
	if config.Credentials.TOTP.AccountName {
		r.accountNames = append(r.accountNames, fmt.Sprintf("%s", value))
	}
	return nil
}

func (r *extensionStub) Finish() error {
	return nil
}

var ctx = context.Background()

func TestExtensionRunner(t *testing.T) {
	t.Run("method=get identifier", func(t *testing.T) {
		for k, tc := range []struct {
			expectErr error
			schema    string
			doc       string
			expect    []string
		}{
			{
				doc:    `{"email":"foo@ory.sh"}`,
				schema: "file://./stub/extension/schema.json",
				expect: []string{"foo@ory.sh"},
			},
			{
				doc:    `{"emails":["foo@ory.sh","bar@ory.sh"]}`,
				schema: "file://./stub/extension/schema.nested.json",
				expect: []string{"foo@ory.sh", "bar@ory.sh"},
			},
		} {
			t.Run(fmt.Sprintf("case=%d", k), func(t *testing.T) {
				c := jsonschema.NewCompiler()
				runner, err := NewExtensionRunner(ctx)
				require.NoError(t, err)

				r := new(extensionStub)
				runner.AddRunner(r).Register(c)

				err = c.MustCompile(ctx, tc.schema).Validate(bytes.NewBufferString(tc.doc))
				if tc.expectErr != nil {
					require.EqualError(t, err, tc.expectErr.Error())
				}

				assert.EqualValues(t, tc.expect, r.identifiers)
				assert.EqualValues(t, tc.expect, r.accountNames)
			})
		}
	})

	t.Run("method=applies meta schema", func(t *testing.T) {
		c := jsonschema.NewCompiler()
		runner, err := NewExtensionRunner(ctx)
		require.NoError(t, err)

		runner.Register(c)

		_, err = c.Compile(ctx, "file://./stub/extension/invalid.schema.json")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "expected boolean, but got number")
	})
}
