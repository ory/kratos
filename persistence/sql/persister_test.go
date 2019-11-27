package sql_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/gobuffalo/pop"
	"github.com/google/uuid"
	"github.com/ory/x/sqlcon/dockertest"
	"github.com/pkg/errors"

	// "github.com/ory/x/sqlcon/dockertest"
	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/persistence/sql"
	"github.com/ory/kratos/selfservice/flow/login"
)

var sqlite = fmt.Sprintf("sqlite://%s.sql", filepath.Join(os.TempDir(), uuid.New().String()))

func init() {
	 internal.RegisterFakes()
}

func TestMain(m *testing.M) {
	atexit := dockertest.NewOnExit()
	atexit.Add(func() {
		_ = os.Remove(strings.TrimPrefix(sqlite, "sqlite://"))
		dockertest.KillAllTestDatabases()
	})
	atexit.Exit(m.Run())
}

func TestPersister(t *testing.T) {
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
				c, err = pop.NewConnection(&pop.ConnectionDetails{URL: dsn})
				if err != nil {
					t.Logf("Unable to connect to database: %+v", err)
					return errors.WithStack(err)
				}
				return nil
			}, bc))

			p, err := sql.NewPersister(c)
			require.NoError(t, err)

			t.Run("case=run up migrations", func(t *testing.T) {
				require.NoError(t, p.MigrateUp(context.Background()))
			})

			t.Run("contract=login.TestRequestPersister", login.TestRequestPersister(p))
			// t.Run("contract=registration.TestRequestPersister", registration.TestRequestPersister(p))
		})
	}
}
