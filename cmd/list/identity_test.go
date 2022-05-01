package list_test

import (
	"context"
	"github.com/ory/kratos/cmd/list"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/internal/testhelpers"
	"github.com/ory/x/cmdx"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
)

func TestListIdentityCmd(t *testing.T) {
	c := list.NewListIdentityCmd(new(cobra.Command))
	reg := testhelpers.CmdSetup(t, c)
	require.NoError(t, c.Flags().Set(cmdx.FlagQuiet, "true"))

	var deleteIdentities = func(t *testing.T, is []*identity.Identity) {
		for _, i := range is {
			require.NoError(t, reg.Persister().DeleteIdentity(context.Background(), i.ID))
		}
	}

	t.Run("case=lists all identities with default pagination", func(t *testing.T) {
		is, ids := testhelpers.CmdMakeIdentities(t, reg, 5)
		defer deleteIdentities(t, is)

		stdOut := testhelpers.CmdExecNoErr(t, c)

		for _, i := range ids {
			assert.Contains(t, stdOut, i)
		}
	})

	t.Run("case=lists all identities with pagination", func(t *testing.T) {
		is, ids := testhelpers.CmdMakeIdentities(t, reg, 6)
		defer deleteIdentities(t, is)

		stdoutP1 := testhelpers.CmdExecNoErr(t, c, "1", "3")
		stdoutP2 := testhelpers.CmdExecNoErr(t, c, "2", "3")

		for _, id := range ids {
			// exactly one of page 1 and 2 should contain the id
			assert.True(t, strings.Contains(stdoutP1, id) != strings.Contains(stdoutP2, id), "%s \n %s", stdoutP1, stdoutP2)
		}
	})
}
