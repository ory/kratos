// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package sql_test

import (
	"context"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/cockroachdb/cockroach-go/v2/testserver"
	"github.com/gobuffalo/pop/v6"
	"github.com/gobuffalo/pop/v6/logging"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/sync/errgroup"

	continuity "github.com/ory/kratos/continuity/test"
	"github.com/ory/kratos/corpx"
	courier "github.com/ory/kratos/courier/test"
	"github.com/ory/kratos/driver"
	"github.com/ory/kratos/driver/config"
	ri "github.com/ory/kratos/identity"
	identity "github.com/ory/kratos/identity/test"
	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/internal/testhelpers"
	"github.com/ory/kratos/persistence/sql"
	sqltesthelpers "github.com/ory/kratos/persistence/sql/testhelpers"
	"github.com/ory/kratos/schema"
	errorx "github.com/ory/kratos/selfservice/errorx/test"
	lf "github.com/ory/kratos/selfservice/flow/login"
	login "github.com/ory/kratos/selfservice/flow/login/test"
	recovery "github.com/ory/kratos/selfservice/flow/recovery/test"
	registration "github.com/ory/kratos/selfservice/flow/registration/test"
	settings "github.com/ory/kratos/selfservice/flow/settings/test"
	verification "github.com/ory/kratos/selfservice/flow/verification/test"
	sessiontokenexchange "github.com/ory/kratos/selfservice/sessiontokenexchange/test"
	code "github.com/ory/kratos/selfservice/strategy/code/test"
	link "github.com/ory/kratos/selfservice/strategy/link/test"
	session "github.com/ory/kratos/session/test"
	"github.com/ory/kratos/x"
	"github.com/ory/kratos/x/xsql"
	"github.com/ory/x/sqlcon"
	"github.com/ory/x/sqlcon/dockertest"
	"github.com/ory/x/sqlxx"
	"github.com/ory/x/urlx"
)

func init() {
	corpx.RegisterFakes()
	pop.SetNowFunc(func() time.Time {
		return time.Now().UTC().Round(time.Second)
	})
	// pop.Debug = true
}

func TestMain(m *testing.M) {
	m.Run()
	dockertest.KillAllTestDatabases()
}

func pl(t testing.TB) func(lvl logging.Level, s string, args ...interface{}) {
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

func createCleanDatabases(t testing.TB) map[string]*driver.RegistryDefault {
	conns := map[string]string{
		"sqlite": "sqlite://file:" + t.TempDir() + "/db.sqlite?_fk=true",
	}

	var l sync.Mutex
	if !testing.Short() {
		funcs := map[string]func(t testing.TB) string{
			"postgres":  dockertest.RunTestPostgreSQL,
			"mysql":     dockertest.RunTestMySQL,
			"cockroach": newLocalTestCRDBServer,
		}

		var wg sync.WaitGroup
		wg.Add(len(funcs))

		for k, f := range funcs {
			go func(s string, f func(t testing.TB) string) {
				defer wg.Done()
				db := f(t)
				l.Lock()
				conns[s] = db
				l.Unlock()
			}(k, f)
		}

		wg.Wait()
	}

	ps := make(map[string]*driver.RegistryDefault, len(conns))
	var wg sync.WaitGroup
	wg.Add(len(conns))
	for name, dsn := range conns {
		go func(name, dsn string) {
			defer wg.Done()
			t.Logf("Connecting to %s: %s", name, dsn)
			_, reg := internal.NewRegistryDefaultWithDSN(t, dsn)
			p := reg.Persister().(*sql.Persister)

			t.Logf("Cleaning up %s", name)
			_ = os.Remove("migrations/schema.sql")
			xsql.CleanSQL(t, p.Connection(context.Background()))
			t.Cleanup(func() {
				xsql.CleanSQL(t, p.Connection(context.Background()))
				_ = os.Remove("migrations/schema.sql")
			})

			t.Logf("Applying %s migrations", name)
			pop.SetLogger(pl(t))
			require.NoError(t, p.MigrateUp(context.Background()))
			t.Logf("%s migrations applied", name)
			status, err := p.MigrationStatus(context.Background())
			require.NoError(t, err)
			require.False(t, status.HasPending())

			l.Lock()
			ps[name] = reg
			l.Unlock()

			t.Logf("Database %s initialized successfully", name)
		}(name, dsn)
	}

	wg.Wait()
	return ps
}

func TestPersister(t *testing.T) {
	conns := createCleanDatabases(t)
	ctx := context.Background()

	for name := range conns {
		name := name
		reg := conns[name]
		t.Run(fmt.Sprintf("database=%s", name), func(t *testing.T) {
			t.Parallel()

			_, p := testhelpers.NewNetwork(t, ctx, reg.Persister())
			conf := reg.Config()

			t.Logf("DSN: %s", conf.DSN(ctx))

			// This test must remain the first test in the test suite!
			t.Run("racy identity creation", func(t *testing.T) {
				defaultSchema := schema.Schema{
					ID:     config.DefaultIdentityTraitsSchemaID,
					URL:    urlx.ParseOrPanic("file://./stub/identity.schema.json"),
					RawURL: "file://./stub/identity.schema.json",
				}

				var wg sync.WaitGroup
				testhelpers.SetDefaultIdentitySchema(reg.Config(), defaultSchema.RawURL)
				_, ps := testhelpers.NewNetwork(t, ctx, reg.Persister())

				for i := 0; i < 10; i++ {
					wg.Add(1)
					// capture i
					ii := i
					go func() {
						defer wg.Done()

						id := ri.NewIdentity("")
						id.SetCredentials(ri.CredentialsTypePassword, ri.Credentials{
							Type:        ri.CredentialsTypePassword,
							Identifiers: []string{fmt.Sprintf("racy identity %d", ii)},
							Config:      sqlxx.JSONRawMessage(`{"foo":"bar"}`),
						})
						id.Traits = ri.Traits("{}")

						require.NoError(t, ps.CreateIdentity(context.Background(), id))
					}()
				}

				wg.Wait()
			})

			t.Run("case=credentials types", func(t *testing.T) {
				for _, ct := range []ri.CredentialsType{ri.CredentialsTypeOIDC, ri.CredentialsTypePassword} {
					require.NoError(t, p.(*sql.Persister).Connection(context.Background()).Where("name = ?", ct).First(&ri.CredentialsTypeTable{}))
				}
			})

			t.Run("contract=identity.TestPool", func(t *testing.T) {
				pop.SetLogger(pl(t))
				identity.TestPool(ctx, conf, p, reg.IdentityManager(), name)(t)
			})
			t.Run("contract=registration.TestFlowPersister", func(t *testing.T) {
				pop.SetLogger(pl(t))
				registration.TestFlowPersister(ctx, p)(t)
			})
			t.Run("contract=errorx.TestPersister", func(t *testing.T) {
				pop.SetLogger(pl(t))
				errorx.TestPersister(ctx, p)(t)
			})
			t.Run("contract=login.TestFlowPersister", func(t *testing.T) {
				pop.SetLogger(pl(t))
				login.TestFlowPersister(ctx, p)(t)
			})
			t.Run("contract=settings.TestFlowPersister", func(t *testing.T) {
				pop.SetLogger(pl(t))
				settings.TestFlowPersister(ctx, conf, p)(t)
			})
			t.Run("contract=session.TestPersister", func(t *testing.T) {
				pop.SetLogger(pl(t))
				session.TestPersister(ctx, conf, p)(t)
			})
			t.Run("contract=sessiontokenexchange.TestPersister", func(t *testing.T) {
				pop.SetLogger(pl(t))
				sessiontokenexchange.TestPersister(ctx, conf, p)(t)
			})
			t.Run("contract=courier.TestPersister", func(t *testing.T) {
				pop.SetLogger(pl(t))
				upsert, insert := sqltesthelpers.DefaultNetworkWrapper(p)
				courier.TestPersister(ctx, upsert, insert)(t)
			})
			t.Run("contract=verification.TestFlowPersister", func(t *testing.T) {
				pop.SetLogger(pl(t))
				verification.TestFlowPersister(ctx, conf, p)(t)
			})
			t.Run("contract=recovery.TestFlowPersister", func(t *testing.T) {
				pop.SetLogger(pl(t))
				recovery.TestFlowPersister(ctx, conf, p)(t)
			})
			t.Run("contract=link.TestPersister", func(t *testing.T) {
				pop.SetLogger(pl(t))
				link.TestPersister(ctx, conf, p)(t)
			})
			t.Run("contract=code.TestPersister", func(t *testing.T) {
				pop.SetLogger(pl(t))
				code.TestPersister(ctx, conf, p)(t)
			})
			t.Run("contract=continuity.TestPersister", func(t *testing.T) {
				pop.SetLogger(pl(t))
				continuity.TestPersister(ctx, p)(t)
			})
		})
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
		i := &ri.Identity{
			ID:     x.NewUUID(),
			State:  ri.StateActive,
			Traits: ri.Traits(`{}`),
		}
		errMessage := "failing because why not"
		err := p.Transaction(context.Background(), func(_ context.Context, connection *pop.Connection) error {
			require.NoError(t, connection.Create(i))
			return errors.Errorf(errMessage)
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), errMessage)
		_, err = p.GetIdentity(context.Background(), i.ID, ri.ExpandNothing)
		require.Error(t, err)
		assert.Equal(t, sqlcon.ErrNoRows.Error(), err.Error())
	})

	t.Run("case=functions should use the context connection", func(t *testing.T) {
		c := p.GetConnection(context.Background())
		errMessage := "some stupid error you can't debug"
		lr := &lf.Flow{
			ID: x.NewUUID(),
		}
		err := c.Transaction(func(tx *pop.Connection) error {
			ctx := sql.WithTransaction(context.Background(), tx)
			require.NoError(t, p.CreateLoginFlow(ctx, lr), "%+v", lr)
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

func Benchmark_BatchCreateIdentities(b *testing.B) {
	conns := createCleanDatabases(b)
	ctx := context.Background()
	batchSizes := []int{1, 10, 100, 500, 800, 900, 1000, 2000, 3000}
	parallelRequests := []int{1, 4, 8, 16}

	for name := range conns {
		name := name
		reg := conns[name]
		b.Run(fmt.Sprintf("database=%s", name), func(b *testing.B) {
			conf := reg.Config()
			_, p := testhelpers.NewNetwork(b, ctx, reg.Persister())
			multipleEmailsSchema := schema.Schema{
				ID:     "multiple_emails",
				URL:    urlx.ParseOrPanic("file://./stub/handler/multiple_emails.schema.json"),
				RawURL: "file://./stub/identity-2.schema.json",
			}
			conf.MustSet(ctx, config.ViperKeyIdentitySchemas, []config.Schema{
				{
					ID:  multipleEmailsSchema.ID,
					URL: multipleEmailsSchema.RawURL,
				},
			})
			conf.MustSet(ctx, config.ViperKeyPublicBaseURL, "http://localhost/")

			run := 0
			for _, batchSize := range batchSizes {
				b.Run(fmt.Sprintf("batch-size=%d", batchSize), func(b *testing.B) {
					for _, paralellism := range parallelRequests {
						b.Run(fmt.Sprintf("parallelism=%d", paralellism), func(b *testing.B) {
							start := time.Now()
							for i := 0; i < b.N; i++ {
								wg := new(errgroup.Group)
								for paralell := 0; paralell < paralellism; paralell++ {
									paralell := paralell
									wg.Go(func() error {
										identities := make([]*ri.Identity, batchSize)
										prefix := fmt.Sprintf("bench-insert-run-%d", run+paralell)
										for j := range identities {
											identities[j] = identity.NewTestIdentity(1, prefix, j)
										}

										return p.CreateIdentities(ctx, identities...)
									})
								}
								assert.NoError(b, wg.Wait())
								run += paralellism
							}
							end := time.Now()
							b.ReportMetric(float64(paralellism*batchSize*b.N), "identites_created")
							b.ReportMetric(float64(paralellism*batchSize*b.N)/end.Sub(start).Seconds(), "identities/s")
						})
					}
				})
			}
		})
	}
}

func newLocalTestCRDBServer(t testing.TB) string {
	ts, err := testserver.NewTestServer(testserver.CustomVersionOpt("23.1.13"))
	require.NoError(t, err)
	t.Cleanup(ts.Stop)

	require.NoError(t, ts.WaitForInit())

	ts.PGURL().Scheme = "cockroach"
	return ts.PGURL().String()
}
