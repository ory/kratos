package verify

import (
	"bytes"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/ory/jsonschema/v3"
	_ "github.com/ory/jsonschema/v3/fileloader"

	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/schema"
	"github.com/ory/kratos/x"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExtensionRunnerIdentityTraits(t *testing.T) {
	id := &identity.Identity{ID: x.NewUUID()}
	for k, tc := range []struct {
		expectErr error
		schema    string
		doc       string
		expect    []Address
	}{
		{
			doc:    `{"username":"foo@ory.sh"}`,
			schema: "file://./stub/extension/schema.json",
			expect: []Address{
				{
					Value:      "foo@ory.sh",
					Verified:   false,
					Status:     StatusPending,
					Via:        ViaEmail,
					IdentityID: id.ID,
				},
			},
		},
		{
			doc:       `{"emails":["foo@ory.sh","bar@ory.sh"], "username": "foobar"}`,
			schema:    "file://./stub/extension/schema.json",
			expectErr: errors.New("I[#/username] S[#/properties/username/format] \"foobar\" is not valid \"email\""),
		},
		{
			doc:    `{"emails":["foo@ory.sh","bar@ory.sh"], "username": "foobar@ory.sh"}`,
			schema: "file://./stub/extension/schema.json",
			expect: []Address{
				{
					Value:      "foo@ory.sh",
					Verified:   false,
					Status:     StatusPending,
					Via:        ViaEmail,
					IdentityID: id.ID,
				},
				{
					Value:      "bar@ory.sh",
					Verified:   false,
					Status:     StatusPending,
					Via:        ViaEmail,
					IdentityID: id.ID,
				},
				{
					Value:      "foobar@ory.sh",
					Verified:   false,
					Status:     StatusPending,
					Via:        ViaEmail,
					IdentityID: id.ID,
				},
			},
		},
	} {
		t.Run(fmt.Sprintf("case=%d", k), func(t *testing.T) {
			c := jsonschema.NewCompiler()
			runner, err := schema.NewExtensionRunner(schema.ExtensionRunnerIdentityMetaSchema)
			require.NoError(t, err)
			e := NewValidationExtensionRunner(id,time.Minute)
			runner.AddRunner(e.Runner).Register(c)

			err = c.MustCompile(tc.schema).Validate(bytes.NewBufferString(tc.doc))
			if tc.expectErr != nil {
				require.EqualError(t, err, tc.expectErr.Error())
				return
			}

			addresses := e.Addresses()
			for k := range addresses {
				assert.NotEmpty(t, addresses[k].Code)
				addresses[k].Code = ""
			}
			assert.EqualValues(t, tc.expect, e.Addresses())
		})
	}
}
