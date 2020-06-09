package migratest

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime/debug"
	"testing"

	"github.com/gobuffalo/packr/v2/plog"
	"github.com/sirupsen/logrus"

	"github.com/ory/viper"
	"github.com/ory/x/logrusx"
	"github.com/ory/x/popx"
	"github.com/ory/x/sqlcon/dockertest"

	"github.com/gobuffalo/pop/v5"
	"github.com/stretchr/testify/require"

	gobuffalologger "github.com/gobuffalo/logger"

	"github.com/ory/kratos/driver"
	"github.com/ory/kratos/driver/configuration"
	"github.com/ory/kratos/internal/testhelpers"
	"github.com/ory/kratos/x"
)

func TestMigrations(t *testing.T) {
	sqlite, err := pop.NewConnection(&pop.ConnectionDetails{
		URL: "sqlite3://" + filepath.Join(os.TempDir(), x.NewUUID().String()) + ".sql?mode=memory&_fk=true",
	})
	require.NoError(t, err)
	require.NoError(t, sqlite.Open())

	connections := map[string]*pop.Connection{
		"sqlite": sqlite,
	}
	l := logrusx.New("", "", logrusx.ForceLevel(logrus.TraceLevel))
	plog.Logger = gobuffalologger.Logrus{FieldLogger: l.Entry}

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
			defer func() {
				if err := recover(); err != nil {
					t.Fatalf("recovered: %+v\n\t%s", err, debug.Stack())
				}
			}()

			t.Logf("Cleaning up before migrations")
			_ = os.Remove("../migrations/schema.sql")
			testhelpers.CleanSQL(t, c)

			t.Cleanup(func() {
				t.Logf("Cleaning up after migrations")
				testhelpers.CleanSQL(t, c)
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

			tm := popx.NewTestMigrator(t, c, "../migrations", "./testdata")
			require.NoError(t, tm.Up())

			viper.Set(configuration.ViperKeyURLsSelfPublic, "https://www.ory.sh/")
			viper.Set(configuration.ViperKeyDefaultIdentityTraitsSchemaURL, "file://stub/default.schema.json")
			viper.Set(configuration.ViperKeyDSN, url)

			d, err := driver.NewDefaultDriver(l, "", "", "", true)
			require.NoError(t, err)

			ids, err := d.Registry().PrivilegedIdentityPool().ListIdentities(context.Background(), 100, 0)
			require.NoError(t, err)

			for _, id := range ids {
				_, err = d.Registry().PrivilegedIdentityPool().GetIdentityConfidential(context.Background(), id.ID)
				require.NoError(t, err)
			}
		}
	}

	for db, c := range connections {
		t.Run(fmt.Sprintf("database=%s", db), test(db, c))
	}
}
