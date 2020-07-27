package driver_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/viper"
	"github.com/ory/x/logrusx"

	driver "github.com/ory/kratos/driver"
	"github.com/ory/kratos/driver/configuration"
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

func TestDriverNew(t *testing.T) {
	viper.Set("dsn", "memory")
	d, err := driver.NewDefaultDriver(logrusx.New("", ""),
		"", "", "", true)
	require.NoError(t, err)
	assert.EqualValues(t, configuration.DefaultSQLiteMemoryDSN, d.Configuration().DSN())
	require.NoError(t, d.Registry().Persister().Ping(context.Background()))
}
