package identities_test

import (
	"bytes"
	"context"
	"io"
	"testing"

	"github.com/ory/x/cmdx"

	"github.com/pkg/errors"

	"github.com/ory/kratos/identity"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/cmd/cliclient"
	"github.com/ory/kratos/driver"
	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/internal/testhelpers"
)

func setup(t *testing.T, cmd *cobra.Command) driver.Registry {
	conf, reg := internal.NewFastRegistryWithMocks(t)
	_, admin := testhelpers.NewKratosServerWithCSRF(t, reg)
	conf.MustSet(config.ViperKeyDefaultIdentitySchemaURL, "file://./stubs/identity.schema.json")
	// setup command
	cliclient.RegisterClientFlags(cmd.Flags())
	cmdx.RegisterFormatFlags(cmd.Flags())
	require.NoError(t, cmd.Flags().Set(cliclient.FlagEndpoint, admin.URL))
	require.NoError(t, cmd.Flags().Set(cmdx.FlagFormat, string(cmdx.FormatJSON)))
	return reg
}

func exec(cmd *cobra.Command, stdIn io.Reader, args ...string) (string, string, error) {
	stdOut, stdErr := &bytes.Buffer{}, &bytes.Buffer{}
	cmd.SetErr(stdErr)
	cmd.SetOut(stdOut)
	cmd.SetIn(stdIn)
	defer cmd.SetIn(nil)
	if args == nil {
		args = []string{}
	}
	cmd.SetArgs(args)
	err := cmd.Execute()
	return stdOut.String(), stdErr.String(), err
}

func execNoErr(t *testing.T, cmd *cobra.Command, args ...string) string {
	stdOut, stdErr, err := exec(cmd, nil, args...)
	require.NoError(t, err)
	require.Len(t, stdErr, 0, stdOut)
	return stdOut
}

func execErr(t *testing.T, cmd *cobra.Command, args ...string) string {
	stdOut, stdErr, err := exec(cmd, nil, args...)
	require.True(t, errors.Is(err, cmdx.ErrNoPrintButFail))
	require.Len(t, stdOut, 0, stdErr)
	return stdErr
}

func makeIdentities(t *testing.T, reg driver.Registry, n int) (is []*identity.Identity, ids []string) {
	for j := 0; j < n; j++ {
		i := identity.NewIdentity(config.DefaultIdentityTraitsSchemaID)
		require.NoError(t, reg.Persister().CreateIdentity(context.Background(), i))
		is = append(is, i)
		ids = append(ids, i.ID.String())
	}
	return
}
