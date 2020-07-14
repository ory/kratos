package driver_test

import (
	"fmt"
	"testing"

	driver "github.com/ory/kratos/driver"
	"github.com/ory/kratos/driver/configuration"
	"github.com/stretchr/testify/assert"
)

func TestDriverDefault_SQLiteMemoryMode(t *testing.T) {
	t.Run("case=settings", func(t *testing.T) {
		for k, tc := range []struct {
			dsn string
			boo bool
		}{
			{dsn: configuration.DefaultSQLiteMemoryDSN, boo: true},
			{dsn: "sqlite://mem.db?mode=asd&_fk=true&cache=shared", boo: false},
			{dsn: "invalidurl", boo: false},
		} {
			t.Run(fmt.Sprintf("run=%d", k), func(t *testing.T) {
				isMem := driver.IsSQLiteMemoryMode(tc.dsn)
				assert.Equal(t, tc.boo, isMem)
			})
		}
	})
}
