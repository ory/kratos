package pkg

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/phayes/freeport"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/cmd"
	"github.com/ory/x/cmdx"
	"github.com/ory/x/configx"
)

func NewKratosServer(t testing.TB) (publicURL, adminURL string) {
	dsn := "sqlite://file::memory:?_fk=true&cache=shared"

	publicPort, err := freeport.GetFreePort()
	require.NoError(t, err)

	adminPort, err := freeport.GetFreePort()
	require.NoError(t, err)

	ctx := configx.ContextWithConfigOptions(context.Background(),
		configx.WithValue("dsn", dsn),
		configx.WithValue("dev", true),
		configx.WithValue("log.level", "trace"),
		configx.WithValue("serve.public.port", publicPort),
		configx.WithValue("serve.admin.port", adminPort),
	)
	ctx, cancel := context.WithCancel(ctx)
	t.Cleanup(cancel)

	executor := &cmdx.CommandExecuter{
		New: func() *cobra.Command {
			return cmd.RootCmd
		},
		Ctx: ctx,
	}
	_ = executor.ExecNoErr(t, "")

	go func() {
		t.Log("Starting server...")
		_ = executor.ExecNoErr(t, "serve", "--config", "../../pkg/stub/kratos.yaml")
	}()

	publicURL = fmt.Sprintf("http://127.0.0.1:%d", publicPort)
	adminURL = fmt.Sprintf("http://127.0.0.1:%d", adminPort)

	startedAt := time.Now()
	for {
		_, err := http.Get(fmt.Sprintf("%s/health/alive", publicURL))
		if err == nil {
			break
		}

		time.Sleep(50 * time.Millisecond)
		if time.Now().Sub(startedAt).Seconds() > 5 {
			t.Fatalf("Unable to reach server.")
		}
	}

	return
}
