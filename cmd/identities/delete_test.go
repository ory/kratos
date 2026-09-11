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
	"github.com/ory/x/cmdx"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/x"
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
		id := x.NewUUID().String()
		stdErr := cmd.ExecExpectedErr(t, id)

		assert.Equal(t, id+": Unable to locate the resource\n", stdErr)
	})

	t.Run("case=continues deleting after failures", func(t *testing.T) {
		is, ids := makeIdentities(t, reg, 2)
		missing := []string{x.NewUUID().String(), x.NewUUID().String()}

		stdOut, stdErr, err := cmd.Exec(nil, missing[0], ids[0], missing[1], ids[1])

		require.ErrorIs(t, err, cmdx.ErrNoPrintButFail)
		assert.JSONEq(t, `["`+strings.Join(ids, `","`)+`"]`, stdOut)
		assert.Equal(t, missing[0]+": Unable to locate the resource\n"+missing[1]+": Unable to locate the resource\n", stdErr)
		for _, i := range is {
			_, err := reg.Persister().GetIdentity(context.Background(), i.ID, identity.ExpandNothing)
			assert.ErrorIs(t, err, sqlcon.ErrNoRows())
		}
	})
}

type deleteRoundTripper func(*http.Request) (*http.Response, error)

func (f deleteRoundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

func TestDeleteCmdTransportError(t *testing.T) {
	failedID, deletedID := x.NewUUID().String(), x.NewUUID().String()
	var requested []string
	ctx := context.WithValue(t.Context(), cliclient.ClientContextKey, func(*cobra.Command) (*cliclient.ClientContext, error) {
		return &cliclient.ClientContext{
			Endpoint: "http://kratos.test",
			HTTPClient: &http.Client{Transport: deleteRoundTripper(func(r *http.Request) (*http.Response, error) {
				assert.Equal(t, http.MethodDelete, r.Method)
				requested = append(requested, r.URL.Path)
				if strings.HasSuffix(r.URL.Path, failedID) {
					return nil, errors.New("connection failed")
				}
				return &http.Response{StatusCode: http.StatusNoContent, Header: make(http.Header), Body: http.NoBody, Request: r}, nil
			})},
		}, nil
	})
	cmd := identities.NewDeleteIdentityCmd()
	cmdx.RegisterFormatFlags(cmd.Flags())
	stdOut, stdErr, err := cmdx.ExecCtx(ctx, cmd, nil, "--format", "json", failedID, deletedID)

	require.ErrorIs(t, err, cmdx.ErrNoPrintButFail)
	assert.Equal(t, deletedID, gjson.Parse(stdOut).String())
	assert.Equal(t, []string{"/admin/identities/" + failedID, "/admin/identities/" + deletedID}, requested)
	assert.True(t, strings.HasPrefix(stdErr, failedID+": "), stdErr)
	assert.Contains(t, stdErr, "connection failed")
	assert.NotContains(t, stdErr, cmdx.ErrNoPrintButFail.Error())
	assert.True(t, cmd.SilenceErrors)
	assert.True(t, cmd.SilenceUsage)
}
