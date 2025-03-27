// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package sql_test

import (
	"context"
	"fmt"
	"regexp"
	"strings"
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
	"github.com/ory/kratos/persistence/sql/batch"
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
		"sqlite": "sqlite://file:" + t.TempDir() + "/db.sqlite?_fk=true&max_conns=1&lock=false",
	}

	if !testing.Short() {
		funcs := map[string]func(t testing.TB) string{
			"postgres": func(t testing.TB) string {
				return dockertest.RunTestPostgreSQLWithVersion(t, "16")
			},
			"mysql": func(t testing.TB) string {
				return dockertest.RunTestMySQLWithVersion(t, "8.4")
			},
			"cockroach": newLocalTestCRDBServer,
		}

		var wg sync.WaitGroup
		wg.Add(len(funcs))

		for k, f := range funcs {
			go func(s string, f func(t testing.TB) string) {
				defer wg.Done()
				db := f(t)
				conns[s] = db
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

			if name != "sqlite" {
				require.EventuallyWithT(t, func(t *assert.CollectT) {
					c, err := pop.NewConnection(&pop.ConnectionDetails{URL: dsn})
					require.NoError(t, err)
					require.NoError(t, c.Open())
					dbName := "testdb" + strings.ReplaceAll(x.NewUUID().String(), "-", "")
					require.NoError(t, c.RawQuery("CREATE DATABASE "+dbName).Exec())
					dsn = regexp.MustCompile("/[a-z0-9]+\\?").ReplaceAllString(dsn, "/"+dbName+"?")
				}, 20*time.Second, 100*time.Millisecond)
			}

			t.Logf("Connecting to %s: %s", name, dsn)

			_, reg := internal.NewRegistryDefaultWithDSN(t, dsn)
			p := reg.Persister().(*sql.Persister)

			t.Logf("Applying %s migrations", name)
			pop.SetLogger(pl(t))
			require.NoError(t, p.MigrateUp(context.Background()))
			t.Logf("%s migrations applied", name)
			status, err := p.MigrationStatus(context.Background())
			require.NoError(t, err)
			require.False(t, status.HasPending())

			ps[name] = reg

			t.Logf("Database %s initialized successfully", name)
		}(name, dsn)
	}

	wg.Wait()
	return ps
}

func TestPersister(t *testing.T) {
	t.Parallel()

	conns := createCleanDatabases(t)
	ctx := testhelpers.WithDefaultIdentitySchema(context.Background(), "file://./stub/identity.schema.json")

	for name, reg := range conns {
		t.Run(fmt.Sprintf("database=%s", name), func(t *testing.T) {
			t.Parallel()

			_, p := testhelpers.NewNetwork(t, ctx, reg.Persister())

			t.Logf("DSN: %s", reg.Config().DSN(ctx))

			t.Run("racy identity creation", func(t *testing.T) {
				t.Parallel()

				var wg sync.WaitGroup

				_, ps := testhelpers.NewNetwork(t, ctx, reg.Persister())

				for i := range 10 {
					wg.Add(1)
					go func() {
						defer wg.Done()

						id := ri.NewIdentity("")
						id.SetCredentials(ri.CredentialsTypePassword, ri.Credentials{
							Type:        ri.CredentialsTypePassword,
							Identifiers: []string{fmt.Sprintf("racy identity %d", i)},
							Config:      sqlxx.JSONRawMessage(`{"foo":"bar"}`),
						})
						id.Traits = ri.Traits("{}")

						require.NoError(t, ps.CreateIdentity(ctx, id))
					}()
				}

				wg.Wait()
			})

			t.Run("case=credential types exist", func(t *testing.T) {
				t.Parallel()
				for _, ct := range []ri.CredentialsType{ri.CredentialsTypeOIDC, ri.CredentialsTypePassword} {
					require.NoError(t, p.(*sql.Persister).Connection(context.Background()).Where("name = ?", ct).First(&ri.CredentialsTypeTable{}))
				}
			})

			t.Run("contract=identity.TestPool", func(t *testing.T) {
				t.Parallel()
				identity.TestPool(ctx, p, reg.IdentityManager(), name)(t)
			})
			t.Run("contract=registration.TestFlowPersister", func(t *testing.T) {
				t.Parallel()
				registration.TestFlowPersister(ctx, p)(t)
			})
			t.Run("contract=errorx.TestPersister", func(t *testing.T) {
				t.Parallel()
				errorx.TestPersister(ctx, p)(t)
			})
			t.Run("contract=login.TestFlowPersister", func(t *testing.T) {
				t.Parallel()
				login.TestFlowPersister(ctx, p)(t)
			})
			t.Run("contract=settings.TestFlowPersister", func(t *testing.T) {
				t.Parallel()
				settings.TestFlowPersister(ctx, p)(t)
			})
			t.Run("contract=session.TestPersister", func(t *testing.T) {
				t.Parallel()
				session.TestPersister(ctx, reg.Config(), p)(t)
			})
			t.Run("contract=sessiontokenexchange.TestPersister", func(t *testing.T) {
				t.Parallel()
				sessiontokenexchange.TestPersister(ctx, p)(t)
			})
			t.Run("contract=courier.TestPersister", func(t *testing.T) {
				t.Parallel()
				upsert, insert := sqltesthelpers.DefaultNetworkWrapper(p)
				courier.TestPersister(ctx, upsert, insert)(t)
			})
			t.Run("contract=verification.TestFlowPersister", func(t *testing.T) {
				t.Parallel()
				verification.TestFlowPersister(ctx, p)(t)
			})
			t.Run("contract=recovery.TestFlowPersister", func(t *testing.T) {
				t.Parallel()
				recovery.TestFlowPersister(ctx, p)(t)
			})
			t.Run("contract=link.TestPersister", func(t *testing.T) {
				t.Parallel()
				link.TestPersister(ctx, p)(t)
			})
			t.Run("contract=code.TestPersister", func(t *testing.T) {
				t.Parallel()
				code.TestPersister(ctx, p)(t)
			})
			t.Run("contract=continuity.TestPersister", func(t *testing.T) {
				t.Parallel()
				continuity.TestPersister(ctx, p)(t)
			})
			t.Run("contract=batch.TestPersister", func(t *testing.T) {
				t.Parallel()
				batch.TestPersister(ctx, reg.Tracer(ctx), p)(t)
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
	t.Parallel()

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
			return errors.New(errMessage)
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
			return errors.New(errMessage)
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
	ts, err := testserver.NewTestServer(testserver.CustomVersionOpt("v23.1.13"))
	require.NoError(t, err)
	t.Cleanup(ts.Stop)

	require.NoError(t, ts.WaitForInit())

	ts.PGURL().Scheme = "cockroach"
	return ts.PGURL().String()
}
