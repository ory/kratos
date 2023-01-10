// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package migratest

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/ory/x/servicelocatorx"

	"github.com/ory/x/fsx"

	"github.com/ory/kratos/identity"

	"github.com/bradleyjkemp/cupaloy/v2"
	"github.com/stretchr/testify/assert"

	"github.com/ory/x/dbal"

	"github.com/ory/kratos/x/xsql"

	"github.com/ory/x/migratest"

	"github.com/gobuffalo/pop/v6"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/driver"
	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/selfservice/flow/login"
	"github.com/ory/kratos/selfservice/flow/recovery"
	"github.com/ory/kratos/selfservice/flow/registration"
	"github.com/ory/kratos/selfservice/flow/settings"
	"github.com/ory/kratos/selfservice/flow/verification"
	"github.com/ory/kratos/selfservice/strategy/code"
	"github.com/ory/kratos/selfservice/strategy/link"
	"github.com/ory/kratos/session"
	"github.com/ory/kratos/x"
	"github.com/ory/x/configx"
	"github.com/ory/x/logrusx"
	"github.com/ory/x/popx"
	"github.com/ory/x/sqlcon"
	"github.com/ory/x/sqlcon/dockertest"
)

func init() {
	dbal.RegisterDriver(func() dbal.Driver {
		return driver.NewRegistryDefault()
	})
}

func snapshotFor(paths ...string) *cupaloy.Config {
	return cupaloy.New(
		cupaloy.CreateNewAutomatically(true),
		cupaloy.FailOnUpdate(true),
		cupaloy.SnapshotFileExtension(".json"),
		cupaloy.SnapshotSubdirectory(filepath.Join(paths...)),
	)
}

func CompareWithFixture(t *testing.T, actual interface{}, prefix string, id string) {
	s := snapshotFor("fixtures", prefix)
	actualJSON, err := json.MarshalIndent(actual, "", "  ")
	require.NoError(t, err)
	assert.NoError(t, s.SnapshotWithName(id, actualJSON))
}

func TestMigrations(t *testing.T) {
	sqlite, err := pop.NewConnection(&pop.ConnectionDetails{
		URL: "sqlite3://" + filepath.Join(os.TempDir(), x.NewUUID().String()) + ".sql?_fk=true"})
	require.NoError(t, err)
	require.NoError(t, sqlite.Open())

	connections := map[string]*pop.Connection{
		"sqlite": sqlite,
	}
	if !testing.Short() {
		dockertest.Parallel([]func(){
			func() {
				connections["postgres"] = dockertest.ConnectToTestPostgreSQLPop(t)
			},
			func() {
				connections["mysql"] = dockertest.ConnectToTestMySQLPop(t)
			},
			func() {
				connections["cockroach"] = dockertest.ConnectToTestCockroachDBPop(t)
			},
		})
	}

	var test = func(db string, c *pop.Connection) func(t *testing.T) {
		return func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			l := logrusx.New("", "", logrusx.ForceLevel(logrus.ErrorLevel))

			t.Logf("Cleaning up before migrations")
			_ = os.Remove("../migrations/sql/schema.sql")
			xsql.CleanSQL(t, c)

			t.Cleanup(func() {
				t.Logf("Cleaning up after migrations")
				xsql.CleanSQL(t, c)
				require.NoError(t, c.Close())
			})

			url := c.URL()
			// workaround for https://github.com/gobuffalo/pop/issues/538
			switch db {
			case "mysql":
				url = "mysql://" + url
			case "sqlite":
				url = "sqlite3://" + url
			}
			t.Logf("URL: %s", url)

			t.Run("suite=up", func(t *testing.T) {
				tm, err := popx.NewMigrationBox(
					fsx.Merge(os.DirFS("../migrations/sql")),
					popx.NewMigrator(c, logrusx.New("", "", logrusx.ForceLevel(logrus.DebugLevel)), nil, 1*time.Minute),
					popx.WithTestdata(t, os.DirFS("./testdata")),
				)
				require.NoError(t, err)
				require.NoError(t, tm.Up(ctx))
			})

			t.Run("suite=fixtures", func(t *testing.T) {
				wg := &sync.WaitGroup{}

				d, err := driver.New(
					context.Background(),
					os.Stderr,
					servicelocatorx.NewOptions(),
					nil,
					[]configx.OptionModifier{
						configx.WithValues(map[string]interface{}{
							config.ViperKeyDSN:             url,
							config.ViperKeyPublicBaseURL:   "https://www.ory.sh/",
							config.ViperKeyIdentitySchemas: config.Schemas{{ID: "default", URL: "file://stub/default.schema.json"}},
							config.ViperKeySecretsDefault:  []string{"secret"},
						}),
						configx.SkipValidation(),
					},
				)
				require.NoError(t, err)

				t.Run("case=identity", func(t *testing.T) {
					wg.Add(1)
					defer wg.Done()
					t.Parallel()

					ids, err := d.PrivilegedIdentityPool().ListIdentities(context.Background(), 0, 1000)
					require.NoError(t, err)
					require.NotEmpty(t, ids)

					var found []string
					for _, id := range ids {
						found = append(found, id.ID.String())
						actual, err := d.PrivilegedIdentityPool().GetIdentityConfidential(context.Background(), id.ID)
						require.NoError(t, err, "ID: %s", id.ID)

						for _, a := range actual.VerifiableAddresses {
							CompareWithFixture(t, a, "identity_verification_address", a.ID.String())
						}

						for _, a := range actual.RecoveryAddresses {
							CompareWithFixture(t, a, "identity_recovery_address", a.ID.String())
						}

						// Prevents ordering to get in the way.
						actual.VerifiableAddresses = nil
						actual.RecoveryAddresses = nil
						CompareWithFixture(t, identity.WithCredentialsAndAdminMetadataInJSON(*actual), "identity", id.ID.String())
					}

					migratest.ContainsExpectedIds(t, filepath.Join("fixtures", "identity"), found)
				})

				t.Run("case=verification_token", func(t *testing.T) {
					wg.Add(1)
					defer wg.Done()
					t.Parallel()

					var ids []link.VerificationToken

					require.NoError(t, c.All(&ids))
					require.NotEmpty(t, ids)

					for _, id := range ids {
						CompareWithFixture(t, id, "verification_token", id.ID.String())
					}
				})

				t.Run("case=session", func(t *testing.T) {
					wg.Add(1)
					defer wg.Done()
					t.Parallel()

					var ids []session.Session
					require.NoError(t, c.Select("id").All(&ids))
					require.NotEmpty(t, ids)

					var found []string
					for _, id := range ids {
						found = append(found, id.ID.String())
						actual, err := d.SessionPersister().GetSession(context.Background(), id.ID, session.ExpandEverything)
						require.NoErrorf(t, err, "Trying to get session: %s", id.ID)
						require.NotEmpty(t, actual.LogoutToken, "check if migrations have generated a logout token for existing sessions")
						CompareWithFixture(t, actual, "session", id.ID.String())
					}
					migratest.ContainsExpectedIds(t, filepath.Join("fixtures", "session"), found)
				})

				t.Run("case=login", func(t *testing.T) {
					wg.Add(1)
					defer wg.Done()
					t.Parallel()

					var ids []login.Flow
					require.NoError(t, c.Select("id").All(&ids))
					require.NotEmpty(t, ids)

					var found []string
					for _, id := range ids {
						found = append(found, id.ID.String())
						actual, err := d.LoginFlowPersister().GetLoginFlow(context.Background(), id.ID)
						require.NoError(t, err)
						CompareWithFixture(t, actual, "login_flow", id.ID.String())
					}
					migratest.ContainsExpectedIds(t, filepath.Join("fixtures", "login_flow"), found)
				})

				t.Run("case=registration", func(t *testing.T) {
					wg.Add(1)
					defer wg.Done()
					t.Parallel()

					var ids []registration.Flow
					require.NoError(t, c.Select("id").All(&ids))
					require.NotEmpty(t, ids)

					var found []string
					for _, id := range ids {
						found = append(found, id.ID.String())
						actual, err := d.RegistrationFlowPersister().GetRegistrationFlow(context.Background(), id.ID)
						require.NoError(t, err)
						CompareWithFixture(t, actual, "registration_flow", id.ID.String())
					}
					migratest.ContainsExpectedIds(t, filepath.Join("fixtures", "registration_flow"), found)
				})

				t.Run("case=settings_flow", func(t *testing.T) {
					wg.Add(1)
					defer wg.Done()
					t.Parallel()

					var ids []settings.Flow
					require.NoError(t, c.Select("id").All(&ids))
					require.NotEmpty(t, ids)

					var found []string
					for _, id := range ids {
						found = append(found, id.ID.String())
						actual, err := d.SettingsFlowPersister().GetSettingsFlow(context.Background(), id.ID)
						require.NoError(t, err, id.ID.String())
						CompareWithFixture(t, actual, "settings_flow", id.ID.String())
					}
					migratest.ContainsExpectedIds(t, filepath.Join("fixtures", "settings_flow"), found)
				})

				t.Run("case=recovery_flow", func(t *testing.T) {
					wg.Add(1)
					defer wg.Done()
					t.Parallel()

					var ids []recovery.Flow
					require.NoError(t, c.Select("id").All(&ids))
					require.NotEmpty(t, ids)

					var found []string
					for _, id := range ids {
						found = append(found, id.ID.String())
						actual, err := d.RecoveryFlowPersister().GetRecoveryFlow(context.Background(), id.ID)
						require.NoError(t, err)
						CompareWithFixture(t, actual, "recovery_flow", id.ID.String())
					}
					migratest.ContainsExpectedIds(t, filepath.Join("fixtures", "recovery_flow"), found)
				})

				t.Run("case=verification_flow", func(t *testing.T) {
					wg.Add(1)
					defer wg.Done()
					t.Parallel()

					var ids []verification.Flow
					require.NoError(t, c.Select("id").All(&ids))
					require.NotEmpty(t, ids)

					var found []string
					for _, id := range ids {
						found = append(found, id.ID.String())
						actual, err := d.VerificationFlowPersister().GetVerificationFlow(context.Background(), id.ID)
						require.NoError(t, err)
						CompareWithFixture(t, actual, "verification_flow", id.ID.String())
					}
					migratest.ContainsExpectedIds(t, filepath.Join("fixtures", "verification_flow"), found)
				})

				t.Run("case=recovery_token", func(t *testing.T) {
					wg.Add(1)
					defer wg.Done()
					t.Parallel()

					var ids []link.RecoveryToken
					require.NoError(t, c.All(&ids))
					require.NotEmpty(t, ids)

					var found []string
					for _, id := range ids {
						found = append(found, id.ID.String())
						CompareWithFixture(t, id, "recovery_token", id.ID.String())
					}
					migratest.ContainsExpectedIds(t, filepath.Join("fixtures", "recovery_token"), found)
				})

				t.Run("case=recovery_code", func(t *testing.T) {
					wg.Add(1)
					defer wg.Done()
					t.Parallel()

					var ids []code.RecoveryCode
					require.NoError(t, c.All(&ids))
					require.NotEmpty(t, ids)

					var found []string
					for _, id := range ids {
						found = append(found, id.ID.String())
						CompareWithFixture(t, id, "recovery_code", id.ID.String())
					}
					migratest.ContainsExpectedIds(t, filepath.Join("fixtures", "recovery_code"), found)
				})

				t.Run("suite=constraints", func(t *testing.T) {
					// This is not really a parallel test, but we have to mark it parallel so the other tests run first.
					t.Parallel()
					wg.Wait()

					sr, err := d.SettingsFlowPersister().GetSettingsFlow(context.Background(), x.ParseUUID("a79bfcf1-68ae-49de-8b23-4f96921b8341"))
					require.NoError(t, err)

					require.NoError(t, d.PrivilegedIdentityPool().DeleteIdentity(context.Background(), sr.IdentityID))

					_, err = d.SettingsFlowPersister().GetSettingsFlow(context.Background(), x.ParseUUID("a79bfcf1-68ae-49de-8b23-4f96921b8341"))
					require.Error(t, err)
					require.ErrorIs(t, err, sqlcon.ErrNoRows)
				})
			})

			t.Run("suite=down", func(t *testing.T) {
				tm := popx.NewTestMigrator(t, c, os.DirFS("../migrations/sql"), os.DirFS("./testdata"), l)
				require.NoError(t, tm.Down(ctx, -1))
			})
		}
	}

	for db, c := range connections {
		t.Run(fmt.Sprintf("database=%s", db), test(db, c))
	}
}
