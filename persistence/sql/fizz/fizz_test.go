package fizz

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/gobuffalo/packr/v2"
	"github.com/gobuffalo/pop/v5"
	"github.com/stretchr/testify/require"
)

var sqlite = fmt.Sprintf("sqlite3://%s.sqlite?_fk=true&mode=rwc", filepath.Join(os.TempDir(), strconv.FormatInt(time.Now().Unix(), 10)))

var migrations = packr.New("migrations", "migrations")

func TestFizzSoundness(t *testing.T) {
	_ = os.Setenv("TEST_DATABASE_SQLITE",sqlite)
	for db, dsn := range map[string]string{
		"sqlite":"TEST_DATABASE_SQLITE",
		"mysql":"TEST_DATABASE_MYSQL",
		"psql":"TEST_DATABASE_POSTGRESQL",
		"crdb":"TEST_DATABASE_COCKROACHDB",
	} {
		t.Run(fmt.Sprintf("case=%s",db), func(t *testing.T) {
			if len(os.Getenv(dsn)) == 0 {
				t.SkipNow()
			}

			c, err := pop.NewConnection(&pop.ConnectionDetails{URL: sqlite})
			require.NoError(t, err)

			require.NoError(t, c.Open())

			m, err := pop.NewMigrationBox(migrations, c)
			require.NoError(t, err)
			m.SchemaPath = db

			require.NoError(t, m.Status(os.Stderr))
			require.NoError(t, m.DumpMigrationSchema())
			require.NoError(t, m.Up())

			t.Logf("dsn: %s", sqlite)
		})
	}
}
