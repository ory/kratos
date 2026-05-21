// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package sql_test

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/cockroachdb/cockroach-go/v2/testserver"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/sync/errgroup"

	continuity "github.com/ory/kratos/continuity/test"
	"github.com/ory/kratos/corpx"
	courier "github.com/ory/kratos/courier/test"
	"github.com/ory/kratos/driver/config"
	ri "github.com/ory/kratos/identity"
	identity "github.com/ory/kratos/identity/test"
	"github.com/ory/kratos/persistence/sql"
	"github.com/ory/kratos/persistence/sql/batch"
	sqltesthelpers "github.com/ory/kratos/persistence/sql/testhelpers"
	"github.com/ory/kratos/pkg"
	"github.com/ory/kratos/pkg/testhelpers"
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
	"github.com/ory/pop/v6"
	"github.com/ory/pop/v6/logging"
	"github.com/ory/x/dbal"
	"github.com/ory/x/popx"
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
}

func TestMain(m *testing.M) {
	os.Exit(m.Run())
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

var dbNames = []string{"sqlite", "postgres", "mysql", "cockroach"}

func setupDatabase(t testing.TB, name string) string {
	var dsn string
	switch name {
	case "sqlite":
		dsn = dbal.NewSQLiteTestDatabase(t)
	case "postgres":
		dsn = dockertest.RunTestPostgreSQLWithVersion(t, "16")
	case "mysql":
		dsn = dockertest.RunTestMySQLWithVersion(t, "8.4")
	case "cockroach":
		dsn = newLocalTestCRDBServer(t)
	default:
		t.Fatalf("unknown database: %s", name)
	}

	if name != "sqlite" {
		require.EventuallyWithT(t, func(t *assert.CollectT) {
			c, err := pop.NewConnection(&pop.ConnectionDetails{URL: dsn})
			require.NoError(t, err)
			require.NoError(t, c.Open())
			dbName := "testdb" + strings.ReplaceAll(x.NewUUID().String(), "-", "")
			require.NoError(t, c.RawQuery("CREATE DATABASE "+dbName).Exec())
			dsn = regexp.MustCompile(`/[a-z0-9]+\?`).ReplaceAllString(dsn, "/"+dbName+"?")
		}, 60*time.Second, 200*time.Millisecond)
	}

	t.Logf("Connecting to %s: %s", name, dsn)

	_, reg := pkg.NewRegistryDefaultWithDSN(t, dsn)
	p := reg.Persister().(*sql.Persister)

	t.Logf("Applying %s migrations", name)
	pop.SetLogger(pl(t))
	require.NoError(t, p.MigrateUp(t.Context()))
	t.Logf("%s migrations applied", name)
	status, err := p.MigrationStatus(t.Context())
	require.NoError(t, err)
	require.False(t, status.HasPending())

	t.Logf("Database %s initialized successfully", name)
	return dsn
}

func TestPersister(t *testing.T) {
	t.Parallel()

	ctx := testhelpers.WithDefaultIdentitySchema(t.Context(), "file://./stub/identity.schema.json")

	for _, name := range dbNames {
		t.Run(fmt.Sprintf("database=%s", name), func(t *testing.T) {
			t.Parallel()

			if name != "sqlite" && testing.Short() {
				t.Skip("skipping non-sqlite under -short")
			}

			dsn := setupDatabase(t, name)
			t.Logf("DSN: %s", dsn)

			t.Run("racy identity creation", func(t *testing.T) {
				t.Parallel()
				dsn := dsn
				// Just have a separate DB for sqlite to speed it up.
				if name == "sqlite" {
					dsn = dbal.NewSQLiteTestDatabase(t)
					// Pre-run migrations so concurrent goroutine startups below
					// don't race to apply the same migrations on a fresh DB.
					pkg.NewRegistryDefaultWithDSN(t, dsn)
				}

				var wg sync.WaitGroup

				for i := range 10 {
					wg.Add(1)
					go func() {
						defer wg.Done()

						_, reg := pkg.NewRegistryDefaultWithDSN(t, dsn)
						_, ps := testhelpers.NewNetwork(t, ctx, reg.Persister())
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
				dsn := dsn
				// Just have a separate DB for sqlite to speed it up.
				if name == "sqlite" {
					dsn = dbal.NewSQLiteTestDatabase(t)
				}

				_, reg := pkg.NewRegistryDefaultWithDSN(t, dsn)
				_, p := testhelpers.NewNetwork(t, ctx, reg.Persister())
				for _, ct := range []ri.CredentialsType{ri.CredentialsTypeOIDC, ri.CredentialsTypePassword} {
					require.NoError(t, p.(*sql.Persister).Connection(t.Context()).Where("name = ?", ct).First(&ri.CredentialsTypeTable{}))
				}
			})

			t.Run("contract=identity.TestPool", func(t *testing.T) {
				t.Parallel()
				dsn := dsn
				// Just have a separate DB for sqlite to speed it up.
				if name == "sqlite" {
					dsn = dbal.NewSQLiteTestDatabase(t)
				}

				_, reg := pkg.NewRegistryDefaultWithDSN(t, dsn)
				_, p := testhelpers.NewNetwork(t, ctx, reg.Persister())
				identity.TestPool(ctx, p, reg.IdentityManager(), name)(t)
			})
			t.Run("contract=registration.TestFlowPersister", func(t *testing.T) {
				t.Parallel()
				dsn := dsn
				// Just have a separate DB for sqlite to speed it up.
				if name == "sqlite" {
					dsn = dbal.NewSQLiteTestDatabase(t)
				}

				_, reg := pkg.NewRegistryDefaultWithDSN(t, dsn)
				_, p := testhelpers.NewNetwork(t, ctx, reg.Persister())
				registration.TestFlowPersister(ctx, p)(t)
			})
			t.Run("contract=errorx.TestPersister", func(t *testing.T) {
				t.Parallel()
				dsn := dsn
				// Just have a separate DB for sqlite to speed it up.
				if name == "sqlite" {
					dsn = dbal.NewSQLiteTestDatabase(t)
				}

				_, reg := pkg.NewRegistryDefaultWithDSN(t, dsn)
				_, p := testhelpers.NewNetwork(t, ctx, reg.Persister())
				errorx.TestPersister(ctx, p)(t)
			})
			t.Run("contract=login.TestFlowPersister", func(t *testing.T) {
				t.Parallel()
				dsn := dsn
				// Just have a separate DB for sqlite to speed it up.
				if name == "sqlite" {
					dsn = dbal.NewSQLiteTestDatabase(t)
				}

				_, reg := pkg.NewRegistryDefaultWithDSN(t, dsn)
				_, p := testhelpers.NewNetwork(t, ctx, reg.Persister())
				login.TestFlowPersister(ctx, p)(t)
			})
			t.Run("contract=settings.TestFlowPersister", func(t *testing.T) {
				t.Parallel()
				dsn := dsn
				// Just have a separate DB for sqlite to speed it up.
				if name == "sqlite" {
					dsn = dbal.NewSQLiteTestDatabase(t)
				}

				_, reg := pkg.NewRegistryDefaultWithDSN(t, dsn)
				_, p := testhelpers.NewNetwork(t, ctx, reg.Persister())
				settings.TestFlowPersister(ctx, p)(t)
			})
			t.Run("contract=session.TestPersister", func(t *testing.T) {
				t.Parallel()
				dsn := dsn
				// Just have a separate DB for sqlite to speed it up.
				if name == "sqlite" {
					dsn = dbal.NewSQLiteTestDatabase(t)
				}

				_, reg := pkg.NewRegistryDefaultWithDSN(t, dsn)
				_, p := testhelpers.NewNetwork(t, ctx, reg.Persister())
				session.TestPersister(ctx, reg.Config(), p)(t)
			})
			t.Run("contract=sessiontokenexchange.TestPersister", func(t *testing.T) {
				t.Parallel()
				dsn := dsn
				// Just have a separate DB for sqlite to speed it up.
				if name == "sqlite" {
					dsn = dbal.NewSQLiteTestDatabase(t)
				}

				_, reg := pkg.NewRegistryDefaultWithDSN(t, dsn)
				_, p := testhelpers.NewNetwork(t, ctx, reg.Persister())
				sessiontokenexchange.TestPersister(ctx, p)(t)
			})
			t.Run("contract=courier.TestPersister", func(t *testing.T) {
				t.Parallel()
				dsn := dsn
				// Just have a separate DB for sqlite to speed it up.
				if name == "sqlite" {
					dsn = dbal.NewSQLiteTestDatabase(t)
				}

				_, reg := pkg.NewRegistryDefaultWithDSN(t, dsn)
				_, p := testhelpers.NewNetwork(t, ctx, reg.Persister())
				upsert, insert := sqltesthelpers.DefaultNetworkWrapper(p)
				courier.TestPersister(ctx, upsert, insert)(t)
			})
			t.Run("contract=verification.TestFlowPersister", func(t *testing.T) {
				t.Parallel()
				dsn := dsn
				// Just have a separate DB for sqlite to speed it up.
				if name == "sqlite" {
					dsn = dbal.NewSQLiteTestDatabase(t)
				}

				_, reg := pkg.NewRegistryDefaultWithDSN(t, dsn)
				_, p := testhelpers.NewNetwork(t, ctx, reg.Persister())
				verification.TestFlowPersister(ctx, p)(t)
			})
			t.Run("contract=recovery.TestFlowPersister", func(t *testing.T) {
				t.Parallel()
				dsn := dsn
				// Just have a separate DB for sqlite to speed it up.
				if name == "sqlite" {
					dsn = dbal.NewSQLiteTestDatabase(t)
				}

				_, reg := pkg.NewRegistryDefaultWithDSN(t, dsn)
				_, p := testhelpers.NewNetwork(t, ctx, reg.Persister())
				recovery.TestFlowPersister(ctx, p)(t)
			})
			t.Run("contract=link.TestPersister", func(t *testing.T) {
				t.Parallel()
				dsn := dsn
				// Just have a separate DB for sqlite to speed it up.
				if name == "sqlite" {
					dsn = dbal.NewSQLiteTestDatabase(t)
				}

				_, reg := pkg.NewRegistryDefaultWithDSN(t, dsn)
				_, p := testhelpers.NewNetwork(t, ctx, reg.Persister())
				link.TestPersister(ctx, p)(t)
			})
			t.Run("contract=code.TestPersister", func(t *testing.T) {
				t.Parallel()
				dsn := dsn
				// Just have a separate DB for sqlite to speed it up.
				if name == "sqlite" {
					dsn = dbal.NewSQLiteTestDatabase(t)
				}

				_, reg := pkg.NewRegistryDefaultWithDSN(t, dsn)
				_, p := testhelpers.NewNetwork(t, ctx, reg.Persister())
				code.TestPersister(ctx, p)(t)
			})
			t.Run("contract=continuity.TestPersister", func(t *testing.T) {
				t.Parallel()
				dsn := dsn
				// Just have a separate DB for sqlite to speed it up.
				if name == "sqlite" {
					dsn = dbal.NewSQLiteTestDatabase(t)
				}

				_, reg := pkg.NewRegistryDefaultWithDSN(t, dsn)
				_, p := testhelpers.NewNetwork(t, ctx, reg.Persister())
				continuity.TestPersister(ctx, p)(t)
			})
			t.Run("contract=batch.TestPersister", func(t *testing.T) {
				t.Parallel()
				dsn := dsn
				// Just have a separate DB for sqlite to speed it up.
				if name == "sqlite" {
					dsn = dbal.NewSQLiteTestDatabase(t)
				}

				_, reg := pkg.NewRegistryDefaultWithDSN(t, dsn)
				_, p := testhelpers.NewNetwork(t, ctx, reg.Persister())
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

	_, reg := pkg.NewFastRegistryWithMocks(t)
	p := reg.Persister()

	t.Run("case=should not create identity because callback returned error", func(t *testing.T) {
		i := &ri.Identity{
			ID:     x.NewUUID(),
			State:  ri.StateActive,
			Traits: ri.Traits(`{}`),
		}
		errMessage := "failing because why not"
		err := p.Transaction(t.Context(), func(_ context.Context, connection *pop.Connection) error {
			require.NoError(t, connection.Create(i))
			return errors.New(errMessage)
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), errMessage)
		_, err = p.GetIdentity(t.Context(), i.ID, ri.ExpandNothing)
		require.Error(t, err)
		assert.Equal(t, sqlcon.ErrNoRows().Error(), err.Error())
	})

	t.Run("case=functions should use the context connection", func(t *testing.T) {
		c := p.GetConnection(t.Context())
		errMessage := "some stupid error you can't debug"
		lr := &lf.Flow{
			ID: x.NewUUID(),
		}
		err := c.Transaction(func(tx *pop.Connection) error {
			ctx := popx.WithTransaction(t.Context(), tx)
			require.NoError(t, p.CreateLoginFlow(ctx, lr), "%+v", lr)
			require.NoError(t, getErr(p.GetLoginFlow(ctx, lr.ID)), "%+v", lr)
			return errors.New(errMessage)
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), errMessage)
		_, err = p.GetLoginFlow(t.Context(), lr.ID)
		require.Error(t, err)
		assert.Equal(t, sqlcon.ErrNoRows().Error(), err.Error())
	})
}

func Benchmark_BatchCreateIdentities(b *testing.B) {
	ctx := b.Context()
	batchSizes := []int{1, 10, 100, 500, 800, 900, 1000, 2000, 3000}
	parallelRequests := []int{1, 4, 8, 16}

	for _, name := range dbNames {
		b.Run(fmt.Sprintf("database=%s", name), func(b *testing.B) {
			dsn := setupDatabase(b, name)
			_, reg := pkg.NewRegistryDefaultWithDSN(b, dsn)
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

										return p.CreateIdentities(ctx, identities)
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
	ts, err := testserver.NewTestServer(testserver.CustomVersionOpt("v25.4.1"))
	require.NoError(t, err)
	t.Cleanup(ts.Stop)

	require.NoError(t, ts.WaitForInit())

	ts.PGURL().Scheme = "cockroach"
	return ts.PGURL().String()
}
