// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package identities_test

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/tidwall/gjson"

	"github.com/ory/kratos/cmd/cliclient"
	"github.com/ory/kratos/cmd/identities"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/x"
	"github.com/ory/x/cmdx"
	"github.com/ory/x/sqlcon"
)

func TestDeleteCmd(t *testing.T) {
	reg, cmd := setup(t, identities.NewDeleteIdentityCmd)

	t.Run("case=deletes successfully", func(t *testing.T) {
		// create identity to delete
		i := identity.NewIdentity(config.DefaultIdentityTraitsSchemaID)
		require.NoError(t, reg.Persister().CreateIdentity(context.Background(), i))

		stdOut := cmd.ExecNoErr(t, i.ID.String())

		// expect ID and no error
		assert.Equal(t, i.ID.String(), gjson.Parse(stdOut).String())

		// expect identity to be deleted
		_, err := reg.Persister().GetIdentity(context.Background(), i.ID, identity.ExpandNothing)
		assert.True(t, errors.Is(err, sqlcon.ErrNoRows()))
	})

	t.Run("case=deletes three identities", func(t *testing.T) {
		is, ids := makeIdentities(t, reg, 3)

		stdOut := cmd.ExecNoErr(t, ids...)

		assert.Equal(t, `["`+strings.Join(ids, "\",\"")+"\"]\n", stdOut)

		for _, i := range is {
			_, err := reg.Persister().GetIdentity(context.Background(), i.ID, identity.ExpandNothing)
			assert.Error(t, err)
		}
	})

	t.Run("case=fails with unknown ID", func(t *testing.T) {
		unknownID := x.NewUUID().String()
		stdErr := cmd.ExecExpectedErr(t, unknownID)

		assert.Contains(t, stdErr, "Unable to locate the resource", stdErr)
		assert.Contains(t, stdErr, unknownID, stdErr)
		assert.NotContains(t, stdErr, "this error should never be printed", stdErr)
	})

	t.Run("case=fails with mix of known and unknown IDs", func(t *testing.T) {
		is, ids := makeIdentities(t, reg, 1)
		unknownID := x.NewUUID().String()

		// the unknown ID goes first: a regression that returns on the first
		// error instead of continuing the loop would still delete ids[0] and
		// pass if it were listed first, so this ordering is what actually
		// proves deletion continues past a failure.
		stdOut, stdErr, err := cmd.Exec(nil, unknownID, ids[0])
		require.Error(t, err)

		assert.Contains(t, stdOut, ids[0], stdOut)
		assert.NotContains(t, stdOut, unknownID, stdOut)
		assert.Contains(t, stdErr, unknownID, stdErr)
		assert.NotContains(t, stdErr, "this error should never be printed", stdErr)

		// the known identity should still have been deleted despite the other ID failing
		_, err = reg.Persister().GetIdentity(context.Background(), is[0].ID, identity.ExpandNothing)
		assert.True(t, errors.Is(err, sqlcon.ErrNoRows()))
	})

	t.Run("case=surfaces transport errors instead of the sentinel", func(t *testing.T) {
		// PrintOpenAPIError only pretty-prints and returns the ErrNoPrintButFail
		// sentinel for errors that carry a parsed API response body. A transport
		// failure has no body, so PrintOpenAPIError returns the original error
		// unprinted; the command must still surface it instead of discarding it.
		//
		// The failure is injected via a RoundTripper returning a known error,
		// rather than dialing a real port, so the test is hermetic and asserts
		// on a value we control instead of an OS-dependent dial error string.
		transportErr := errors.New("simulated transport failure")
		ctx := context.WithValue(context.Background(), cliclient.ClientContextKey,
			func(cmd *cobra.Command) (*cliclient.ClientContext, error) {
				return &cliclient.ClientContext{
					Endpoint:   "http://localhost",
					HTTPClient: &http.Client{Transport: failingRoundTripper{err: transportErr}},
				}, nil
			})

		unreachable := &cmdx.CommandExecuter{
			Ctx: ctx,
			New: func() *cobra.Command {
				c := identities.NewDeleteIdentityCmd()
				cliclient.RegisterClientFlags(c.Flags())
				cmdx.RegisterFormatFlags(c.Flags())
				return c
			},
			PersistentArgs: []string{"--" + cmdx.FlagFormat, string(cmdx.FormatJSON)},
		}

		stdErr := unreachable.ExecExpectedErr(t, x.NewUUID().String())

		assert.NotContains(t, stdErr, "this error should never be printed", stdErr)
		// asserts on the actual injected error, not just "something got
		// printed" - the bug this guards against prints only the "Failed to
		// delete identity <id>:" header and silently drops the real error,
		// which would still make a weaker "output is non-empty" check pass.
		assert.Contains(t, stdErr, transportErr.Error(), stdErr)
	})
}

// failingRoundTripper always fails with a known error, standing in for a real
// transport failure (e.g. a timeout) without depending on real network I/O.
type failingRoundTripper struct{ err error }

func (rt failingRoundTripper) RoundTrip(*http.Request) (*http.Response, error) {
	return nil, rt.err
}
