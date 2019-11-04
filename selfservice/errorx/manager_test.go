package errorx_test

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/x/sqlcon/dockertest"

	"github.com/ory/herodot"

	"github.com/ory/hive/internal"
	. "github.com/ory/hive/selfservice/errorx"
)

// nolint: staticcheck
func TestMain(m *testing.M) {
	flag.Parse()
	runner := dockertest.Register()
	runner.Exit(m.Run())
}

func TestManager(t *testing.T) {
	conf, reg := internal.NewMemoryRegistry(t)
	var managers = map[string]Manager{
		"memory": NewManagerMemory(reg, conf),
	}

	if !testing.Short() {
		var l sync.Mutex
		dockertest.Parallel([]func(){
			func() {
				db, err := dockertest.ConnectToTestPostgreSQL()
				require.NoError(t, err)

				_, reg := internal.NewRegistrySQL(t, db)

				l.Lock()
				managers["postgres"] = reg.ErrorManager()
				l.Unlock()
			},
		})
	}

	for n, m := range managers {
		t.Run(fmt.Sprintf("manager=%s", n), func(t *testing.T) {
			sent := herodot.ErrNotFound.WithReason("foobar")
			id, err := m.Add(context.Background(), sent)
			require.NoError(t, err)

			var expected bytes.Buffer
			require.NoError(t, json.NewEncoder(&expected).Encode([]error{sent}))

			actual, err := m.Read(context.Background(), id)
			require.NoError(t, err)

			actualS, _ := json.Marshal(actual)
			assert.JSONEq(t, expected.String(), string(actualS))

			require.NoError(t, m.Clear(context.Background(), time.Second, false))
			actual, err = m.Read(context.Background(), id)
			require.NoError(t, err)

			gotu, _ := json.Marshal(actual)
			sentu, _ := json.Marshal([]error{sent})

			assert.JSONEq(t, string(sentu), string(gotu))

			time.Sleep(time.Millisecond * 100)

			require.NoError(t, m.Clear(context.Background(), time.Millisecond, false))
			_, err = m.Read(context.Background(), id)
			require.Error(t, err)
		})
	}
}
