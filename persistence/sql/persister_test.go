package sql_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/gobuffalo/pop"
	"github.com/gobuffalo/pop/logging"
	"github.com/google/uuid"
	"github.com/pkg/errors"

	"github.com/ory/x/sqlcon/dockertest"

	// "github.com/ory/x/sqlcon/dockertest"
	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/persistence/sql"
	"github.com/ory/kratos/selfservice/flow/login"
	"github.com/ory/kratos/selfservice/flow/profile"
	"github.com/ory/kratos/selfservice/flow/registration"
	"github.com/ory/kratos/session"
)

// Workaround for https://github.com/gobuffalo/pop/pull/481
var sqlite = fmt.Sprintf("sqlite3://%s.sqlite?_fk=true&mode=rwc", filepath.Join(os.TempDir(), uuid.New().String()))

func init() {
	internal.RegisterFakes()
	pop.Debug = true
}

func TestMain(m *testing.M) {
	atexit := dockertest.NewOnExit()
	atexit.Add(func() {
		// _ = os.Remove(strings.TrimPrefix(sqlite, "sqlite://"))
		dockertest.KillAllTestDatabases()
	})
	atexit.Exit(m.Run())
}

func pl(t *testing.T) func(lvl logging.Level, s string, args ...interface{}) {
	return func(lvl logging.Level, s string, args ...interface{}) {
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
	t.Logf("%s", sqlite)

	conf, reg := internal.NewMemoryRegistry(t)

	for name, dsn := range map[string]string{
		"sqlite": sqlite,
		// "postgres": dockertest.RunTestPostgreSQL(t),
	} {
		t.Run("database="+name, func(t *testing.T) {
			var c *pop.Connection
			var err error

			bc := backoff.NewExponentialBackOff()
			bc.MaxElapsedTime = time.Minute / 2
			bc.Reset()
			require.NoError(t, backoff.Retry(func() (err error) {
				c, err = pop.NewConnection(&pop.ConnectionDetails{
					URL: dsn,
				})
				if err != nil {
					t.Logf("Unable to connect to database: %+v", err)
					return errors.WithStack(err)
				}
				return c.Open()
			}, bc))

			p, err := sql.NewPersister(reg, conf, c)
			require.NoError(t, err)
			defer c.Close()

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
		})
	}

	t.Logf("sqlite location: %s", sqlite)
}
