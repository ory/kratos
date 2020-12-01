package sql_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/ory/kratos/driver/configuration"
	"github.com/ory/kratos/schema"
	"github.com/ory/viper"
	"github.com/ory/x/sqlxx"
	"github.com/ory/x/urlx"

	"github.com/go-errors/errors"
	"github.com/stretchr/testify/assert"

	"github.com/ory/x/sqlcon"

	"github.com/ory/kratos/continuity"
	"github.com/ory/kratos/internal/testhelpers"
	"github.com/ory/kratos/persistence/sql"
	"github.com/ory/kratos/selfservice/errorx"
	"github.com/ory/kratos/selfservice/flow/recovery"
	"github.com/ory/kratos/selfservice/strategy/link"
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
	"github.com/ory/kratos/selfservice/flow/verification"
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
		// _ = os.Remove(strings.TrimPrefix(sqlite, "sqlite://"))
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

func createCleanDatabases(t *testing.T) map[string]*sql.Persister {
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

	t.Logf("sqlite: %s", sqlite)

	ps := make(map[string]*sql.Persister, len(conns))
	for name, dsn := range conns {
		_, reg := internal.NewRegistryDefaultWithDSN(t, dsn)
		p := reg.Persister().(*sql.Persister)

		_ = os.Remove("migrations/schema.sql")
		testhelpers.CleanSQL(t, p.Connection())
		t.Cleanup(func() {
			testhelpers.CleanSQL(t, p.Connection())
			_ = os.Remove("migrations/schema.sql")
		})

		pop.SetLogger(pl(t))
		require.NoError(t, p.MigrationStatus(context.Background(), os.Stderr))
		require.NoError(t, p.MigrateUp(context.Background()))

		ps[name] = p
	}

	return ps
}

func TestPersister(t *testing.T) {
	conns := createCleanDatabases(t)

	for name, p := range conns {
		t.Run(fmt.Sprintf("database=%s", name), func(t *testing.T) {
			t.Run("contract=identity.TestPool", func(t *testing.T) {
				pop.SetLogger(pl(t))
				identity.TestPool(p)(t)
			})
			t.Run("contract=registration.TestFlowPersister", func(t *testing.T) {
				pop.SetLogger(pl(t))
				registration.TestFlowPersister(p)(t)
			})
			t.Run("contract=errorx.TestPersister", func(t *testing.T) {
				pop.SetLogger(pl(t))
				errorx.TestPersister(p)(t)
			})
			t.Run("contract=login.TestFlowPersister", func(t *testing.T) {
				pop.SetLogger(pl(t))
				login.TestFlowPersister(p)(t)
			})
			t.Run("contract=settings.TestFlowPersister", func(t *testing.T) {
				pop.SetLogger(pl(t))
				settings.TestRequestPersister(p)(t)
			})
			t.Run("contract=session.TestFlowPersister", func(t *testing.T) {
				pop.SetLogger(pl(t))
				session.TestPersister(p)(t)
			})
			t.Run("contract=courier.TestPersister", func(t *testing.T) {
				pop.SetLogger(pl(t))
				courier.TestPersister(p)(t)
			})
			t.Run("contract=verification.TestPersister", func(t *testing.T) {
				pop.SetLogger(pl(t))
				verification.TestFlowPersister(p)(t)
			})
			t.Run("contract=recovery.TestFlowPersister", func(t *testing.T) {
				pop.SetLogger(pl(t))
				recovery.TestFlowPersister(p)(t)
			})
			t.Run("contract=link.TestPersister", func(t *testing.T) {
				pop.SetLogger(pl(t))
				link.TestPersister(p)(t)
			})
			t.Run("contract=continuity.TestPersister", func(t *testing.T) {
				pop.SetLogger(pl(t))
				continuity.TestPersister(p)(t)
			})
		})

		t.Logf("DSN: %s", p.Connection().URL())
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
	_, reg := internal.NewFastRegistryWithMocks(t)
	p := reg.Persister()

	t.Run("case=should not create identity because callback returned error", func(t *testing.T) {
		i := &identity.Identity{
			ID:     x.NewUUID(),
			Traits: identity.Traits(`{}`),
		}
		errMessage := "failing because why not"
		err := p.Transaction(context.Background(), func(ctx context.Context, connection *pop.Connection) error {
			require.NoError(t, connection.Create(i))
			return errors.Errorf(errMessage)
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), errMessage)
		_, err = p.GetIdentity(context.Background(), i.ID)
		require.Error(t, err)
		assert.Equal(t, sqlcon.ErrNoRows.Error(), err.Error())
	})

	t.Run("case=functions should use the context connection", func(t *testing.T) {
		c := p.GetConnection(context.Background())
		errMessage := "some stupid error you can't debug"
		lr := &login.Flow{
			ID: x.NewUUID(),
		}
		err := c.Transaction(func(tx *pop.Connection) error {
			ctx := sql.WithTransaction(context.Background(), tx)
			require.NoError(t, p.CreateLoginFlow(ctx, lr), "%+v", lr)
			require.NoError(t, p.UpdateLoginFlowMethod(ctx, lr.ID, identity.CredentialsTypePassword, &login.FlowMethod{}))
			require.NoError(t, getErr(p.GetLoginFlow(ctx, lr.ID)), "%+v", lr)
			return errors.Errorf(errMessage)
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), errMessage)
		_, err = p.GetLoginFlow(context.Background(), lr.ID)
		require.Error(t, err)
		assert.Equal(t, sqlcon.ErrNoRows.Error(), err.Error())
	})
}

// note: it is important that this test runs on clean databases, as the racy behaviour only happens there
func TestPersister_CreateIdentityRacy(t *testing.T) {
	defaultSchema := schema.Schema{
		ID:     configuration.DefaultIdentityTraitsSchemaID,
		URL:    urlx.ParseOrPanic("file://./stub/identity.schema.json"),
		RawURL: "file://./stub/identity.schema.json",
	}

	for name, p := range createCleanDatabases(t) {
		viper.Set(configuration.ViperKeyDefaultIdentitySchemaURL, defaultSchema.RawURL)

		t.Run(fmt.Sprintf("db=%s", name), func(t *testing.T) {
			wg := sync.WaitGroup{}

			for i := 0; i < 10; i++ {
				wg.Add(1)
				// capture i
				ii := i
				go func() {
					defer wg.Done()

					id := identity.NewIdentity("")
					id.SetCredentials(identity.CredentialsTypePassword, identity.Credentials{
						Type:        identity.CredentialsTypePassword,
						Identifiers: []string{fmt.Sprintf("racy identity %d", ii)},
						Config:      sqlxx.JSONRawMessage(`{"foo":"bar"}`),
					})
					id.Traits = identity.Traits("{}")

					require.NoError(t, p.CreateIdentity(context.Background(), id))
				}()
			}

			wg.Wait()
		})

	}
}
