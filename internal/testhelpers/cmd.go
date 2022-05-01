package testhelpers

import (
	"bytes"
	"context"
	_ "embed"
	"io"
	"testing"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/cmd/cliclient"
	"github.com/ory/kratos/driver"
	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/internal"
	"github.com/ory/x/cmdx"
)

//go:embed stub/identity.schema.json
var identitySchema []byte

func CmdSetup(t *testing.T, cmd *cobra.Command) driver.Registry {
	conf, reg := internal.NewFastRegistryWithMocks(t)
	_, admin := NewKratosServerWithCSRF(t, reg)
	SetDefaultIdentitySchemaFromRaw(conf, identitySchema)
	// CmdSetup command
	cliclient.RegisterClientFlags(cmd.Flags())
	cmdx.RegisterFormatFlags(cmd.Flags())
	require.NoError(t, cmd.Flags().Set(cliclient.FlagEndpoint, admin.URL))
	require.NoError(t, cmd.Flags().Set(cmdx.FlagFormat, string(cmdx.FormatJSON)))
	return reg
}

func CmdExec(cmd *cobra.Command, stdIn io.Reader, args ...string) (string, string, error) {
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

func CmdExecNoErr(t *testing.T, cmd *cobra.Command, args ...string) string {
	stdOut, stdErr, err := CmdExec(cmd, nil, args...)
	require.NoError(t, err, "stdout: %s\nstderr: %s", stdOut, stdErr)
	require.Len(t, stdErr, 0, stdOut)
	return stdOut
}

func CmdExecErr(t *testing.T, cmd *cobra.Command, args ...string) string {
	stdOut, stdErr, err := CmdExec(cmd, nil, args...)
	require.True(t, errors.Is(err, cmdx.ErrNoPrintButFail))
	require.Len(t, stdOut, 0, stdErr)
	return stdErr
}

func CmdMakeIdentities(t *testing.T, reg driver.Registry, n int) (is []*identity.Identity, ids []string) {
	for j := 0; j < n; j++ {
		i := identity.NewIdentity(config.DefaultIdentityTraitsSchemaID)
		i.MetadataPublic = []byte(`{"foo":"bar"}`)
		require.NoError(t, reg.Persister().CreateIdentity(context.Background(), i))
		is = append(is, i)
		ids = append(ids, i.ID.String())
	}
	return
}
