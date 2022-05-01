package delete_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/ory/x/sqlcon"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/cmd/delete"
	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/internal/testhelpers"
	"github.com/ory/kratos/x"
)

func TestDeleteIdentityCmd(t *testing.T) {
	c := delete.NewDeleteIdentityCmd(new(cobra.Command))
	reg := testhelpers.CmdSetup(t, c)

	t.Run("case=deletes successfully", func(t *testing.T) {
		// create identity to delete
		i := identity.NewIdentity(config.DefaultIdentityTraitsSchemaID)
		require.NoError(t, reg.Persister().CreateIdentity(context.Background(), i))

		stdOut := testhelpers.CmdExecNoErr(t, c, i.ID.String())

		// expect ID and no error
		assert.Equal(t, i.ID.String()+"\n", stdOut)

		// expect identity to be deleted
		_, err := reg.Persister().GetIdentity(context.Background(), i.ID)
		assert.True(t, errors.Is(err, sqlcon.ErrNoRows))
	})

	t.Run("case=deletes three identities", func(t *testing.T) {
		is, ids := testhelpers.CmdMakeIdentities(t, reg, 3)

		stdOut := testhelpers.CmdExecNoErr(t, c, ids...)

		assert.Equal(t, strings.Join(ids, "\n")+"\n", stdOut)

		for _, i := range is {
			_, err := reg.Persister().GetIdentity(context.Background(), i.ID)
			assert.Error(t, err)
		}
	})

	t.Run("case=fails with unknown ID", func(t *testing.T) {
		stdErr := testhelpers.CmdExecErr(t, c, x.NewUUID().String())

		assert.Contains(t, stdErr, "404 Not Found", stdErr)
	})
}
