package identities

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/ory/kratos/x"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/identity"
)

func TestGetCmd(t *testing.T) {
	reg := setup(t, GetCmd)

	t.Run("case=gets a single identity", func(t *testing.T) {
		i := identity.NewIdentity(config.DefaultIdentityTraitsSchemaID)
		require.NoError(t, reg.Persister().CreateIdentity(context.Background(), i))

		stdOut := execNoErr(t, GetCmd, i.ID.String())

		ij, err := json.Marshal(i)
		require.NoError(t, err)

		assert.Equal(t, string(ij)+"\n", stdOut)
	})

	t.Run("case=gets three identities", func(t *testing.T) {
		is, ids := makeIdentities(t, reg, 3)

		stdOut := execNoErr(t, GetCmd, ids...)

		isj, err := json.Marshal(is)
		require.NoError(t, err)

		assert.Equal(t, string(isj)+"\n", stdOut)
	})

	t.Run("case=fails with unknown ID", func(t *testing.T) {
		stdErr := execErr(t, GetCmd, x.NewUUID().String())

		assert.Contains(t, stdErr, "getIdentityNotFound", stdErr)
	})
}
