package testhelpers

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/avast/retry-go/v3"
	"github.com/phayes/freeport"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/cmd"
	"github.com/ory/kratos/x"
	"github.com/ory/x/cmdx"
	"github.com/ory/x/configx"
)

func StartE2EServer(t *testing.T, configFile string) (publicUrl, adminUrl string) {
	adminPort, err := freeport.GetFreePort()
	require.NoError(t, err)

	publicPort, err := freeport.GetFreePort()
	require.NoError(t, err)

	ctx := configx.ContextWithConfigOptions(context.Background(),
		configx.WithValue("dsn", "memory"),
		configx.WithValue("dev", true),
		configx.WithValue("log.level", "trace"),
		configx.WithValue("log.leak_sensitive_values", true),
		configx.WithValue("serve.public.port", publicPort),
		configx.WithValue("serve.admin.port", adminPort),
	)
	//nolint:staticcheck SA1029
	ctx = context.WithValue(ctx, "dsn", "memory")
	ctx, cancel := context.WithCancel(ctx)
	t.Cleanup(cancel)

	executor := &cmdx.CommandExecuter{New: func() *cobra.Command {
		return cmd.NewRootCmd()
	}, Ctx: ctx}

	go func() {
		t.Log("Starting server...")
		_ = executor.ExecNoErr(t, "serve", "--config", configFile, "--watch-courier")
	}()

	publicUrl = fmt.Sprintf("http://127.0.0.1:%d", publicPort)
	adminUrl = fmt.Sprintf("http://127.0.0.1:%d", adminPort)

	require.NoError(t, retry.Do(func() error {
		res, err := http.Get(publicUrl + "/health/alive")
		if err != nil {
			return err
		}
		defer res.Body.Close()
		body := x.MustReadAll(res.Body)
		if res.StatusCode != http.StatusOK {
			t.Logf("%s", body)
			return fmt.Errorf("expected status code 200 but got: %d", res.StatusCode)
		}
		return nil
	}), err)

	return
}
