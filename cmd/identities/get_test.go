package identities_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/ory/kratos/cmd/identities"

	"github.com/ory/x/assertx"

	"github.com/ory/kratos/x"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/identity"
)

func TestGetCmd(t *testing.T) {
	reg := setup(t, identities.GetCmd)

	t.Run("case=gets a single identity", func(t *testing.T) {
		i := identity.NewIdentity(config.DefaultIdentityTraitsSchemaID)
		require.NoError(t, reg.Persister().CreateIdentity(context.Background(), i))

		stdOut := execNoErr(t, identities.GetCmd, i.ID.String())

		ij, err := json.Marshal(i)
		require.NoError(t, err)

		assertx.EqualAsJSONExcept(t, json.RawMessage(ij), json.RawMessage(stdOut), []string{"created_at", "updated_at"})
	})

	t.Run("case=gets three identities", func(t *testing.T) {
		is, ids := makeIdentities(t, reg, 3)

		stdOut := execNoErr(t, identities.GetCmd, ids...)

		isj, err := json.Marshal(is)
		require.NoError(t, err)

		assertx.EqualAsJSONExcept(t, json.RawMessage(isj), json.RawMessage(stdOut), []string{"created_at", "updated_at"})
	})

	t.Run("case=fails with unknown ID", func(t *testing.T) {
		stdErr := execErr(t, identities.GetCmd, x.NewUUID().String())

		assert.Contains(t, stdErr, "404 Not Found", stdErr)
	})
}
