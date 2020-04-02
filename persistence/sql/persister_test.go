package sql_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/go-errors/errors"
	"github.com/stretchr/testify/assert"

	"github.com/ory/x/sqlcon"

	"github.com/ory/kratos/continuity"
	"github.com/ory/kratos/persistence/sql"
	"github.com/ory/kratos/selfservice/errorx"
	"github.com/ory/kratos/x"

	"github.com/gobuffalo/pop/v5"
	"github.com/gobuffalo/pop/v5/logging"
	"github.com/google/uuid"

	"github.com/ory/x/sqlcon/dockertest"

	// "github.com/ory/x/sqlcon/dockertest"
	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/courier"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/selfservice/flow/login"
	"github.com/ory/kratos/selfservice/flow/registration"
	"github.com/ory/kratos/selfservice/flow/settings"
	"github.com/ory/kratos/selfservice/flow/verify"
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
			"postgres":  dockertest.RunTestPostgreSQL,
			"mysql":     dockertest.RunTestMySQL,
			"cockroach": dockertest.RunTestCockroachDB,
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
			require.NoError(t, p.MigrationStatus(context.Background(), os.Stderr))
			require.NoError(t, p.MigrateUp(context.Background()))

			t.Run("contract=identity.TestPool", func(t *testing.T) {
				pop.SetLogger(pl(t))
				identity.TestPool(p.(identity.PrivilegedPool))(t)
			})
			t.Run("contract=registration.TestRequestPersister", func(t *testing.T) {
				pop.SetLogger(pl(t))
				registration.TestRequestPersister(p)(t)
			})
			t.Run("contract=errorx.TestPersister", func(t *testing.T) {
				pop.SetLogger(pl(t))
				errorx.TestPersister(p)(t)
			})
			t.Run("contract=login.TestRequestPersister", func(t *testing.T) {
				pop.SetLogger(pl(t))
				login.TestRequestPersister(p)(t)
			})
			t.Run("contract=settings.TestRequestPersister", func(t *testing.T) {
				pop.SetLogger(pl(t))
				settings.TestRequestPersister(p)(t)
			})
			t.Run("contract=session.TestRequestPersister", func(t *testing.T) {
				pop.SetLogger(pl(t))
				session.TestPersister(p)(t)
			})
			t.Run("contract=courier.TestPersister", func(t *testing.T) {
				pop.SetLogger(pl(t))
				courier.TestPersister(p)(t)
			})
			t.Run("contract=verify.TestPersister", func(t *testing.T) {
				pop.SetLogger(pl(t))
				verify.TestPersister(p)(t)
			})
			t.Run("contract=continuity.TestPersister", func(t *testing.T) {
				pop.SetLogger(pl(t))
				continuity.TestPersister(p)(t)
			})
		})

		t.Logf("DSN: %s", dsn)
	}
}

func getErr(args ...interface{}) error {
	if len(args) == 0 {
		return nil
	}
	lastArg := args[len(args)-1]
	if e, ok := lastArg.(error); ok {
		return e
	}
	return nil
}

func TestPersister_Transaction(t *testing.T) {
	_, reg := internal.NewRegistryDefault(t)
	p := reg.Persister()

	t.Run("case=should not create identity because callback returned error", func(t *testing.T) {
		i := &identity.Identity{
			ID:     x.NewUUID(),
			Traits: identity.Traits(""),
		}
		errMessage := "failing because why not"
		err := p.Transaction(context.Background(), func(connection *pop.Connection) error {
			require.NoError(t, connection.Create(i))
			return errors.Errorf(errMessage)
		})
		require.Error(t, err)
		assert.Equal(t, errMessage, err.Error())
		_, err = p.GetIdentity(context.Background(), i.ID)
		require.Error(t, err)
		assert.Equal(t, sqlcon.ErrNoRows.Error(), err.Error())
	})

	t.Run("case=functions should use the context connection", func(t *testing.T) {
		c := p.GetConnection(context.Background())
		errMessage := "some stupid error you can't debug"
		lr := &login.Request{
			ID: x.NewUUID(),
		}
		err := c.Transaction(func(tx *pop.Connection) error {
			ctx := sql.WithTransaction(context.Background(), tx)
			require.NoError(t, p.CreateLoginRequest(ctx, lr), "%+v", lr)
			require.NoError(t, p.UpdateLoginRequestMethod(ctx, lr.ID, identity.CredentialsTypePassword, &login.RequestMethod{}))
			require.NoError(t, getErr(p.GetLoginRequest(ctx, lr.ID)), "%+v", lr)
			return errors.Errorf(errMessage)
		})
		require.Error(t, err)
		assert.Equal(t, errMessage, err.Error())
		_, err = p.GetLoginRequest(context.Background(), lr.ID)
		require.Error(t, err)
		assert.Equal(t, sqlcon.ErrNoRows.Error(), err.Error())
	})
}
