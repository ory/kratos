package sql_test

import (
	"fmt"
	"testing"

	"github.com/ory/kratos/persistence/sql"

	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/identity"
)

func TestCredentialTypes(t *testing.T) {
	ps := createCleanDatabases(t)

	for name, p := range ps {
		t.Run(fmt.Sprintf("db=%s", name), func(t *testing.T) {
			for _, ct := range []identity.CredentialsType{identity.CredentialsTypeOIDC, identity.CredentialsTypePassword} {
				require.NoError(t, p.Persister().(*sql.Persister).Connection().Where("name = ?", ct).First(&identity.CredentialsTypeTable{}))
			}
		})

	}
}
