package sql_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/gobuffalo/pop"
	"github.com/gobuffalo/pop/logging"
	"github.com/google/uuid"

	"github.com/ory/x/sqlcon/dockertest"

	// "github.com/ory/x/sqlcon/dockertest"
	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/courier"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/selfservice/flow/login"
	"github.com/ory/kratos/selfservice/flow/profile"
	"github.com/ory/kratos/selfservice/flow/registration"
	"github.com/ory/kratos/session"
)

// Workaround for https://github.com/gobuffalo/pop/pull/481
var sqlite = fmt.Sprintf("sqlite3://%s.sqlite?_fk=true&mode=rwc", filepath.Join(os.TempDir(), uuid.New().String()))

func init() {
	internal.RegisterFakes()
	// op.Debug = true
}

// nolint:staticcheck
func TestMain(m *testing.M) {
	atexit := dockertest.NewOnExit()
	atexit.Add(func() {
		_ = os.Remove(strings.TrimPrefix(sqlite, "sqlite://"))
		dockertest.KillAllTestDatabases()
	})
	atexit.Exit(m.Run())
}

func pl(t *testing.T) func(lvl logging.Level, s string, args ...interface{}) {
	return func(lvl logging.Level, s string, args ...interface{}) {
		if pop.Debug == false {
			return
		}

		if lvl == logging.SQL {
			if len(args) > 0 {
				xargs := make([]string, len(args))
				for i, a := range args {
					switch a.(type) {
					case string:
						xargs[i] = fmt.Sprintf("%q", a)
					default:
						xargs[i] = fmt.Sprintf("%v", a)
					}
				}
				s = fmt.Sprintf("%s - %s | %s", lvl, s, xargs)
			} else {
				s = fmt.Sprintf("%s - %s", lvl, s)
			}
		} else {
			s = fmt.Sprintf(s, args...)
			s = fmt.Sprintf("%s - %s", lvl, s)
		}
		t.Log(s)
	}
}

func TestPersister(t *testing.T) {
	conns := map[string]string{
		"sqlite": sqlite,
	}

	var l sync.Mutex
	if !testing.Short() {
		funcs := map[string]func(t *testing.T) string{
			"postgres": dockertest.RunTestPostgreSQL,
			// "mysql":    dockertest.RunTestMySQL,
			// "cockroach": dockertest.RunTestCockroachDB, // pending: https://github.com/gobuffalo/fizz/pull/69
		}

		var wg sync.WaitGroup
		wg.Add(len(funcs))

		for k, f := range funcs {
			go func(s string, f func(t *testing.T) string) {
				defer wg.Done()
				db := f(t)
				l.Lock()
				conns[s] = db
				l.Unlock()
			}(k, f)
		}

		wg.Wait()
	}

	for name, dsn := range conns {
		t.Run(fmt.Sprintf("database=%s", name), func(t *testing.T) {
			_, reg := internal.NewRegistryDefaultWithDSN(t, dsn)
			p := reg.Persister()

			pop.SetLogger(pl(t))
			require.NoError(t, p.MigrationStatus(context.Background()))
			require.NoError(t, p.MigrateUp(context.Background()))

			t.Run("contract=identity.TestPool", func(t *testing.T) {
				pop.SetLogger(pl(t))
				identity.TestPool(p)(t)
			})
			t.Run("contract=registration.TestRequestPersister", func(t *testing.T) {
				pop.SetLogger(pl(t))
				registration.TestRequestPersister(p)(t)
			})
			t.Run("contract=login.TestRequestPersister", func(t *testing.T) {
				pop.SetLogger(pl(t))
				login.TestRequestPersister(p)(t)
			})
			t.Run("contract=profile.TestRequestPersister", func(t *testing.T) {
				pop.SetLogger(pl(t))
				profile.TestRequestPersister(p)(t)
			})
			t.Run("contract=session.TestRequestPersister", func(t *testing.T) {
				pop.SetLogger(pl(t))
				session.TestPersister(p)(t)
			})
			t.Run("contract=courier.TestPersister", func(t *testing.T) {
				pop.SetLogger(pl(t))
				courier.TestPersister(p)(t)
			})
		})

		t.Logf("DSN: %s", dsn)
	}
}
